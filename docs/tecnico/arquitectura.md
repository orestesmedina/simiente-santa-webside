# Arquitectura técnica — Sitio web de la Iglesia Simiente Santa

**Estado:** documento fundacional · **Fecha:** 2026-09-30 · **Decisiones que lo respaldan:** [decisiones.md](./decisiones.md)

Este documento describe **cómo se construye** el sistema: reglas de dependencia entre capas,
árbol de carpetas del backend, la interfaz que neutraliza el router, un ejemplo vertical completo
de dominio, el flujo de una petición, la estrategia de pruebas y la receta para agregar áreas de
negocio. Lo usan las personas del equipo y los agentes de desarrollo (`dev-backend`,
`dev-frontend`, `devops`) como referencia obligada al escribir código.

Fuentes que este documento respeta y no contradice:

- `.specify/memory/constitution.md` — reglas no negociables (§I…§VIII).
- `.agents/skills/go-backend/SKILL.md`, `.agents/skills/postgres-db/SKILL.md`,
  `.agents/skills/react-frontend/SKILL.md` — convenciones del stack.
- `specs/001-estructura-base/plan.md` y `research.md` — plan vigente de F1 (el router, `sqlc` y
  la versión de Go aquí **cambian** respecto de él: ver D-A3, D-A4 y D-A5 de `decisiones.md`).
- `docs/producto/roadmap.md` — F1…F9.

> **Nota sobre `specs/`**: las specs son la fuente de verdad del **qué** (constitución §I). Este
> documento manda sobre el **cómo**. Si una spec aprobada obliga a cambiar algo de aquí, se
> actualiza este documento en esa misma funcionalidad (lo hace el `documentador` al cerrarla).

---

## 1. Principios y reglas de dependencia

### 1.1 Diagrama de dependencias

```text
cmd/api (main.go) ──► internal/<dominio> ──► internal/platform
        │                    │
        │                    └──► internal/db (generado por sqlc) ──► pgx
        └──► internal/platform
```

### 1.2 Reglas (DEBE)

| # | Regla | Detalle |
|---|---|---|
| R1 | **`cmd` → dominio → `platform`** | La flecha solo apunta hacia abajo. `cmd/api/main.go` compone las piezas; los dominios usan `platform`; `platform` **no conoce ningún dominio** (ni por import ni por nombre). |
| R2 | **Un dominio no importa a otro dominio** | Si dos dominios necesitan lo mismo, eso **sube a `internal/platform/`** (o a `internal/db` si es SQL compartido, vía consulta nombrada). Nunca `contacto` importando `usuarios`. |
| R3 | **Las interfaces las define quien las consume** | El `service` define la interfaz que su `repository` cumple; el `handler` define la interfaz que su `service` cumple. La implementación concreta vive en su archivo y se entrega por constructor. |
| R4 | **El `handler` no sabe SQL y el `repository` no sabe HTTP** | El handler decodifica/valida/responde; el repository solo habla con PostgreSQL. El `service` no conoce ni `net/http` ni `pgx`. |
| R5 | **`context.Context` como primer parámetro de todo lo que hace E/S** | Repositorios, servicios y cualquier llamada que pueda bloquearse. Se respeta su cancelación (`ctx.Done()`), nunca se crea uno nuevo sin derivar del recibido. |
| R6 | **Sin estado global** | Ni variables de paquete mutables, ni singletons, ni `init()` con lógica. Todas las dependencias entran **por constructor** desde `main.go`: composición manual, sin `wire` ni `fx` (D-A6). |
| R7 | **Errores con contexto y traducción única** | Se envuelven con `fmt.Errorf("acción: %w", err)`; los errores de dominio (`apperr`) se traducen a HTTP **solo** en `httpserver.WriteError`, llamado desde el handler (ver §5.11). |
| R8 | **Contrato antes que código** | Todo endpoint nace en `backend/api/openapi.yaml` (o en el delta de `specs/<N>/contracts/`) **antes** de escribir el handler (constitución §II). |

### 1.3 Excepción declarada: `httpserver.Registrar`

La regla R3 dice "la interfaz la define quien la consume", y el consumidor de `Registrar` son
**todos** los `routes.go` de todos los dominios. Definirla una vez por dominio sería duplicación
purа, así que `Registrar` se define en `internal/platform/httpserver/` como **plumbing de
infraestructura neutral** (no expresa nada del negocio) y cada `routes.go` la recibe como
parámetro. Es la única excepción conocida a R3 y está justificada por su neutralidad (ver §4).

---

## 2. Árbol de carpetas del backend

```text
backend/
├── cmd/
│   └── api/
│       └── main.go                  # composición manual de dependencias + arranque/apagado
├── internal/
│   ├── db/                          # CÓDIGO GENERADO por sqlc — no se edita a mano
│   │   ├── queries/                 # consultas SQL anotadas (la fuente sí se edita)
│   │   │   └── contact.sql
│   │   ├── db.go / models.go / *.sql.go
│   ├── <dominio>/                   # status/, contacto/, usuarios/… un paquete por área
│   │   ├── model.go                 # entidades, estados y DTOs (con etiquetas de validación)
│   │   ├── repository.go            # implementación concreta de la interfaz del service (SQL)
│   │   ├── service.go               # interfaz Repository + reglas de negocio (sin HTTP, sin SQL)
│   │   ├── handler.go               # interfaz Service + HTTP (decodificar, validar, responder)
│   │   ├── routes.go                # RegisterPublic / RegisterAdmin
│   │   └── *_test.go
│   └── platform/                    # el "casi framework" del proyecto (ver tabla siguiente)
│       ├── config/
│       ├── logger/
│       ├── apperr/
│       ├── httpserver/
│       ├── middleware/
│       ├── database/
│       ├── migrate/
│       ├── validate/
│       ├── paginate/
│       └── testing/
├── migrations/                      # NNNNNN_nombre.up.sql / .down.sql (golang-migrate)
├── api/
│   └── openapi.yaml                 # CONTRATO VIVO (constitución §II)
├── sqlc.yaml                        # schema: migrations · queries: internal/db/queries · out: internal/db
├── go.mod / go.sum
└── Dockerfile
```

El módulo Go es `simiente-santa/backend` en los ejemplos de este documento; el path definitivo se
fija cuando exista el remoto (acción humana pendiente, ver `plan.md` §Riesgos R1).

**Distinción importante**: `internal/platform/database/` **construye y administra el pool** de
conexiones (infraestructura); `internal/db/` es **código generado por sqlc** a partir de
`internal/db/queries/` (acceso a datos de negocio). No se mezclan.

### 2.1 Paquetes de `internal/platform/`

| Paquete | Responsabilidad | ¿Cuándo existe? |
|---|---|---|
| `config/` | Carga de variables de entorno con valores por defecto y **validación al arrancar** (falla rápido con mensaje claro si algo obligatorio falta o es inválido). `Load() (Config, error)`. | **F1** |
| `logger/` | Constructor de `*slog.Logger` en **JSON** con nivel desde `LOG_LEVEL`, y **logger por petición** (hij con `request_id`, ruta y método). | **F1** (la parte de petición llega con `middleware/request-id`) |
| `apperr/` | Errores tipados de dominio: `Invalid` (400), `Unauthenticated` (401), `Forbidden` (403), `NotFound` (404), `Conflict` (409), `Internal` (500). Llevan `Message` (seguro para el cliente), detalle de campos opcional y el error interno **envuelto** (`%w`), que nunca se serializa. | **F1** (los kinds que usa `/healthz` + `Internal`; crecen bajo demanda) |
| `httpserver/` | Servidor HTTP con timeouts (`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`) y **apagado ordenado** (`signal.NotifyContext` + `Shutdown` con periodo de gracia); la interfaz `Registrar` y el tipo `Middleware` (§4); helpers `WriteJSON` y `WriteError` (§5.11). | **F1** |
| `middleware/` | Cadena transversal: `request-id`, `recover`, `logging`, `CORS`, `rate-limit`, `authn`, `authz` por módulo, `CSRF`. | **F1**: `request-id`, `recover`, `logging`, `CORS` mínimo (una cabecera, sin dependencias — D12 del plan). **Se difiere**: `rate-limit` y `CSRF` con el primer endpoint público escribible; `authn`/`authz` con **F2** (usuarios y permisos) |
| `database/` | Construcción del `*pgxpool.Pool` desde `DATABASE_URL`, helper `WithTx(ctx, fn)` para transacciones y helper de salud (`Ping` con timeout) usado por `/healthz`. | **F1** |
| `migrate/` | Runner de migraciones **embebidas** (`//go:embed`), **opt-in por entorno** (`RUN_MIGRATIONS=true`). | **Se difiere**: en F1 mandan el CLI `golang-migrate` (target `make db-migrate` y el paso "Migraciones" del CI). El runner embebido llega cuando un entorno desplegado necesite auto-migrar sin CLI — decisión del plan de esa funcionalidad |
| `validate/` | Validación de DTOs a partir de sus etiquetas (`required`, `max`, `email`…), produciendo `apperr.Invalid` con detalle por campo. | **Se difiere a F2** (primera vez que hay entradas de usuario de verdad). El mecanismo concreto (librería vs. implementación propia mínima) se decide en el plan de F2 — ver "Preguntas abiertas" de `decisiones.md` |
| `paginate/` | Parseo y topes de paginación desde query string (`limit`, `offset` con máximos) y helper de cursor opaco para listados grandes (convención `postgres-db`). | **Se difiere a F2** (primer listado del panel) |
| `testing/` | Helpers compartidos de prueba: conexión a `DATABASE_URL_TEST` con *skip* automático si no hay BD, builders/fixtures de dominio, utilidades de `httptest`. | **F1** (lo mínimo que use la prueba de `/healthz`; crece con las pruebas) |

Fuera de `platform` pero igual de normativo: `internal/db/` (sqlc, §5.3) y `backend/migrations/`
(convenciones de la skill `postgres-db`: numeración secuencial, `up`/`down` completos, **nunca**
editar una migración aplicada).

### 2.2 Frontend (resumen)

La estructura completa la fija `.agents/skills/react-frontend/SKILL.md`; aquí basta recordar la
regla de oro: **ningún componente llama a `fetch`**. Cada feature expone hooks
(`features/<feature>/hooks/`) que usan funciones tipadas de `frontend/src/api/`, y los tipos de
respuesta salen **generados** del contrato (`frontend/src/api/schema.d.ts`, `npm run api:gen`).
Prohibido `any`; sesión en cookie `HttpOnly`, nunca en `localStorage`.

---

## 3. Qué hay en F1 y qué se difiere (resumen)

| | F1 (estructura base) | Después |
|---|---|---|
| Dominios | `internal/status/` (solo `/healthz`) | `contacto`, `usuarios`… siguiendo §5 |
| Platform | `config`, `logger`, `apperr`, `httpserver`, `middleware` (request-id, recover, logging, CORS), `database`, `testing` | `validate`, `paginate`, `migrate` (F2); `authn`/`authz`/`CSRF`/`rate-limit` (F2); i18n (**F3**) |
| Datos | pool `pgx` + `Ping` (sin tablas de negocio; migración baseline `000001`) | sqlc con la primera consulta de negocio (§5.3) |
| Contrato | `backend/api/openapi.yaml` con `/healthz` | deltas por funcionalidad fusionados en el documento vivo |

---

## 4. La pieza clave: `httpserver.Registrar` (neutraliza el router)

Ningún dominio debe saber qué router hay debajo. La única puerta para publicar rutas es la
interfaz `Registrar`, definida en `internal/platform/httpserver/registrar.go`:

```go
package httpserver

import (
	"net/http"
	"strings"
)

// Middleware envuelve un handler HTTP. Es el tipo de la stdlib, así que cualquier
// middleware compatible con net/http entra sin adaptación.
type Middleware = func(http.Handler) http.Handler

// Registrar es la única puerta que tienen los dominios para publicar rutas.
// No expone el router concreto: F1 la implementa con net/http (patrones
// "GET /ruta" de Go 1.22+). Si mañana se adopta chi, solo cambia el adaptador.
type Registrar interface {
	// Handle publica un handler en method + path (patrón estilo Go: "GET /contactos/{id}").
	Handle(method, path string, h http.HandlerFunc)
	// Group devuelve un Registrar con prefijo y middlewares acumulados.
	Group(prefix string, mws ...Middleware) Registrar
}
```

Adaptador de F1 sobre `net/http` (`muxRegistrar`, mismo archivo):

```go
type muxRegistrar struct {
	mux  *http.ServeMux
	base string      // prefijo del grupo ("" en la raíz)
	mws  []Middleware // middlewares acumulados del grupo
}

// NewMuxRegistrar adapta un http.ServeMux estándar a la interfaz Registrar.
func NewMuxRegistrar(mux *http.ServeMux) Registrar {
	return &muxRegistrar{mux: mux}
}

func (r *muxRegistrar) Handle(method, path string, h http.HandlerFunc) {
	full := joinPath(r.base, path)
	var handler http.Handler = h
	for i := len(r.mws) - 1; i >= 0; i-- { // el primer middleware de la lista queda más fuera
		handler = r.mws[i](handler)
	}
	r.mux.Handle(method+" "+full, handler) // patrón de Go 1.22+: método + ruta
}

func (r *muxRegistrar) Group(prefix string, mws ...Middleware) Registrar {
	acumulados := make([]Middleware, 0, len(r.mws)+len(mws))
	acumulados = append(acumulados, r.mws...)
	acumulados = append(acumulados, mws...)
	return &muxRegistrar{mux: r.mux, base: joinPath(r.base, prefix), mws: acumulados}
}

func joinPath(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	default:
		return strings.TrimSuffix(a, "/") + "/" + strings.TrimPrefix(b, "/")
	}
}
```

**Qué garantiza esto:**

1. **Los dominios no conocen el router.** Sus `routes.go` reciben un `Registrar` y llaman
   `Handle`/`Group`; no importan `net/http` más que por los tipos `http.HandlerFunc`.
2. **F1 se implementa con `net/http`** (patrones `GET /ruta` y comodines `{id}` de Go 1.22+):
   cero dependencias de ruteo (D-A4).
3. **Adoptar chi (u otro) después no toca ningún dominio.** Bastaría un adaptador como este:

   ```go
   // Ejemplo del adaptador chi (solo se escribe si F2 decide adoptarlo — D-A4).
   type chiRegistrar struct {
	r    chi.Router
	mws  []Middleware
   }

   func (c *chiRegistrar) Handle(method, path string, h http.HandlerFunc) {
	var handler http.Handler = h
	for i := len(c.mws) - 1; i >= 0; i-- {
		handler = c.mws[i](handler)
	}
	c.r.Method(method, path, handler)
   }

   func (c *chiRegistrar) Group(prefix string, mws ...Middleware) Registrar {
	return &chiRegistrar{r: c.r, mws: append(append([]Middleware{}, c.mws...), mws...)} // + c.r.Route(prefix, …)
   }
   ```

   El día que esto ocurra, el diff es **un archivo de `internal/platform/httpserver/`** más la
   línea que construye el `Registrar` en `main.go`.

**Detalles deliberados:**

- `Handle` recibe `method` y `path` por separado en lugar de un patrón compuesto: obliga a que
  cada ruta declare su método, y hace el adaptador trivial de portar.
- `Group` acumula middlewares **por grupo**; así las rutas del panel llevan `authn`+`authz`+`CSRF`
  sin que los dominios los mencionen (§5.8).
- `Middleware` es un alias del tipo stdlib (`type Middleware = func(http.Handler) http.Handler`),
  no un tipo nuevo: compatibilidad total con el ecosistema.

---

## 5. Ejemplo vertical: dominio `contacto`

> **Alcance del ejemplo.** Es el patrón de referencia para todos los dominios de negocio. El
> dominio concreto "formulario de contacto del visitante + bandeja del panel" **no está
> comprometido en `docs/producto/roadmap.md`**: si producto lo decide, se construirá con su propia
> spec siguiendo exactamente estos archivos. Se elige porque es el ejemplo más simple que ejercita
> las dos superficies de la API: una ruta **pública** (`POST /api/v1/contacto`, un visitante envía
> un mensaje) y una ruta de **panel** (`GET /api/v1/admin/contactos`, requiere permiso del módulo
> `"contactos"`). Las rutas del panel asumen **F2 entregado** (authn/authz existentes).

Notación de los ejemplos: módulo `simiente-santa/backend`; JSON de la API en `camelCase`
(campos), identificadores Go/TS en inglés, columnas en `snake_case`.

### 5.1 Contrato primero (`backend/api/openapi.yaml`, fragmento)

```yaml
paths:
  /api/v1/contacto:
    post:
      tags: [contact]
      operationId: createContact
      summary: Enviar un mensaje de contacto (público, sin autenticación)
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/ContactInput"
      responses:
        "201":
          description: Mensaje recibido.
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ContactCreated"
        "400":
          $ref: "#/components/responses/Invalid"
        "429":
          $ref: "#/components/responses/TooManyRequests"
  /api/v1/admin/contactos:
    get:
      tags: [contact]
      operationId: listContacts
      summary: Listar mensajes recibidos (panel; requiere permiso del módulo "contactos")
      security: [{ sessionCookie: [] }]
      parameters:
        - name: status
          in: query
          schema: { type: string, enum: [new, read, archived] }
        - name: limit
          in: query
          schema: { type: integer, minimum: 1, maximum: 100, default: 20 }
        - name: offset
          in: query
          schema: { type: integer, minimum: 0, default: 0 }
      responses:
        "200":
          description: Listado paginado de mensajes.
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ContactList"
        "401":
          $ref: "#/components/responses/Unauthenticated"
        "403":
          $ref: "#/components/responses/Forbidden"
components:
  schemas:
    ContactInput:
      type: object
      required: [name, email, message]
      properties:
        name:    { type: string, minLength: 1, maxLength: 120 }
        email:   { type: string, format: email, maxLength: 254 }
        phone:   { type: string, maxLength: 32 }
        message: { type: string, minLength: 1, maxLength: 2000 }
      additionalProperties: false
    ContactCreated:
      type: object
      required: [id, received]
      properties:
        id:       { type: string, format: uuid }
        received: { type: boolean }
      additionalProperties: false
    ContactList:
      type: object
      required: [items, total, limit, offset]
      properties:
        items:
          type: array
          items: { $ref: "#/components/schemas/ContactItem" }
        total:  { type: integer, format: int64 }
        limit:  { type: integer }
        offset: { type: integer }
      additionalProperties: false
    # ContactItem: id, name, email, phone, message, status, createdAt
```

Los esquemas de error (`Invalid`, `Unauthenticated`, `Forbidden`, `TooManyRequests`) son globales y
usan el sobre estándar de §5.11. Los nombres de esquema y operación van en inglés (igual que F1:
`SystemStatus`, `getSystemStatus`); los segmentos de URL en español cuando la superficie pública lo
pide (`/api/v1/contacto`).

### 5.2 Migración (`backend/migrations/`)

Nunca se edita una migración aplicada: la siguiente libre (F1 ya usó `000001_baseline`).

```sql
-- 000002_create_contacts.up.sql
CREATE TABLE contacts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
    email      TEXT NOT NULL CHECK (char_length(email) <= 254),
    phone      TEXT NOT NULL DEFAULT '' CHECK (char_length(phone) <= 32),
    message    TEXT NOT NULL CHECK (char_length(message) BETWEEN 1 AND 2000),
    status     TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'read', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX contacts_status_created_at_idx ON contacts (status, created_at DESC, id DESC);
```

```sql
-- 000002_create_contacts.down.sql
DROP TABLE IF EXISTS contacts;
```

Decisiones del esquema (skill `postgres-db`): tabla en plural e inglés; UUID con
`gen_random_uuid()`; `created_at`/`updated_at` `TIMESTAMPTZ`; `NOT NULL` por defecto (el teléfono
ausente se guarda como `''`, no como `NULL`: no existe el estado "desconocido vs. ausente");
restricciones `CHECK` **en la base**, no solo en el código; índice para el orden del listado del
panel. Se aplica con `make db-migrate` (CLI `golang-migrate`).

### 5.3 Consultas sqlc (`backend/internal/db/queries/contact.sql`)

SQL explícito y parametrizado; **nunca** concatenación de strings (constitución §IV, CWE-89).

```sql
-- name: InsertContact :one
INSERT INTO contacts (name, email, phone, message)
VALUES ($1, $2, $3, $4)
RETURNING id, name, email, phone, message, status, created_at, updated_at;

-- name: ListContacts :many
SELECT id, name, email, phone, message, status, created_at, updated_at
FROM contacts
WHERE (sqlc.arg('filter_status')::text = '' OR status = sqlc.arg('filter_status'))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountContacts :one
SELECT count(*)::bigint AS total
FROM contacts
WHERE (sqlc.arg('filter_status')::text = '' OR status = sqlc.arg('filter_status'));
```

`backend/sqlc.yaml` apunta `schema: migrations`, `queries: internal/db/queries`, `out: internal/db`.
Se ejecuta `sqlc generate` en desarrollo y **el código generado se commitea** (mismo criterio que
`frontend/src/api/schema.d.ts`: el CI no necesita la herramienta). Con `sqlc`, todo cambio de
consulta o de esquema **falla en compilación** si el código no acompaña (D-A3).

> **Dependencia de tipos UUID.** Por defecto sqlc emite `github.com/google/uuid` para las columnas
> `uuid`. Es una entrada nueva en `go.mod` que habrá que justificar en el `plan.md` de la
> funcionalidad que la introduzca (D-A8); si se quiere cero dependencias nuevas, basta un
> *override* en `sqlc.yaml` para emitir `pgtype.UUID`, que ya viene dentro de `pgx/v5`. Decisión
> del plan de F2.

### 5.4 `model.go` — entidad y DTOs

```go
package contacto

import (
	"time"

	"github.com/google/uuid" // ver nota de §5.3
)

// Status es el estado de un mensaje en la bandeja del panel.
type Status string

const (
	StatusNew      Status = "new"
	StatusRead     Status = "read"
	StatusArchived Status = "archived"
)

// Contact es la entidad de dominio. No se serializa tal cual hacia el API.
type Contact struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Phone     string // "" si el visitante no lo dio
	Message   string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateContactInput es el DTO de entrada de Create.
// Las etiquetas "validate" las aplica platform/validate en la frontera HTTP (F2);
// el service vuelve a comprobar sus invariantes (defensa en profundidad, §IV).
type CreateContactInput struct {
	Name    string `json:"name" validate:"required,min=1,max=120"`
	Email   string `json:"email" validate:"required,email,max=254"`
	Phone   string `json:"phone,omitempty" validate:"omitempty,max=32"`
	Message string `json:"message" validate:"required,min=1,max=2000"`
}

// ListFilter acota el listado del panel.
type ListFilter struct {
	Status Status // "" = todos
	Limit  int
	Offset int
}

// ContactItemResponse es un mensaje en el listado del panel (DTO de salida).
type ContactItemResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// ContactListResponse es la respuesta de GET /api/v1/admin/contactos.
type ContactListResponse struct {
	Items  []ContactItemResponse `json:"items"`
	Total  int64                 `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

// ContactCreatedResponse es la respuesta de POST /api/v1/contacto.
type ContactCreatedResponse struct {
	ID       string `json:"id"`
	Received bool   `json:"received"`
}
```

### 5.5 `service.go` — reglas de negocio + interfaz del repository

La define **quien la consume**: `Service` necesita un `Repository`, así que la interfaz vive aquí
y su implementación en `repository.go`. El service no conoce `net/http` ni `pgx`.

```go
package contacto

import (
	"context"
	"fmt"
	"strings"
	"time"

	"simiente-santa/backend/internal/platform/apperr"
)

// Repository es la interfaz que este servicio necesita del almacén de datos.
// La define quien consume (este archivo); la cumple repository.go.
type Repository interface {
	Insert(ctx context.Context, c Contact) (Contact, error)
	List(ctx context.Context, f ListFilter) ([]Contact, error)
	Count(ctx context.Context, f ListFilter) (int64, error)
}

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

// service implementa la interfaz Service (declarada en handler.go, quien la consume).
type service struct {
	repo Repository
	now  func() time.Time // inyectable en pruebas
}

// NewService construye el servicio de contacto.
func NewService(repo Repository) Service {
	return &service{repo: repo, now: time.Now}
}

// Create normaliza el mensaje del visitante y lo guarda como "new".
func (s *service) Create(ctx context.Context, in CreateContactInput) (Contact, error) {
	name := normalizeSpaces(in.Name)
	message := normalizeSpaces(in.Message)
	email := strings.ToLower(strings.TrimSpace(in.Email))
	phone := normalizePhone(in.Phone)

	// Invariantes de negocio (además de la validación de la frontera HTTP).
	if name == "" || message == "" {
		return Contact{}, apperr.Invalid("el nombre y el mensaje no pueden quedar vacíos")
	}

	now := s.now().UTC()
	c := Contact{
		Name:    name,
		Email:   email,
		Phone:   phone,
		Message: message,
		Status:  StatusNew,
		CreatedAt: now,
		UpdatedAt: now,
	}

	saved, err := s.repo.Insert(ctx, c)
	if err != nil {
		return Contact{}, fmt.Errorf("create contact: %w", err)
	}
	return saved, nil
}

// List devuelve la bandeja del panel, acotada y filtrada.
func (s *service) List(ctx context.Context, f ListFilter) ([]Contact, int64, error) {
	if f.Limit <= 0 {
		f.Limit = defaultListLimit
	}
	if f.Limit > maxListLimit {
		f.Limit = maxListLimit
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	if f.Status != "" && f.Status != StatusNew && f.Status != StatusRead && f.Status != StatusArchived {
		return nil, 0, apperr.Invalid("estado de filtro inválido")
	}

	items, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, fmt.Errorf("list contacts: %w", err)
	}
	total, err := s.repo.Count(ctx, f)
	if err != nil {
		return nil, 0, fmt.Errorf("count contacts: %w", err)
	}
	return items, total, nil
}

// normalizeSpaces recorta y colapsa espacios: "  Ana   María " → "Ana María".
func normalizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// normalizePhone conserva solo dígitos y el '+' inicial.
func normalizePhone(s string) string {
	var b strings.Builder
	for i, r := range strings.TrimSpace(s) {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '+' && i == 0:
			b.WriteRune(r)
		}
	}
	return b.String()
}
```

### 5.6 `repository.go` — implementación concreta sobre PostgreSQL

Implementa la interfaz `Repository` del paso anterior. No conoce HTTP ni los DTOs de salida.

```go
package contacto

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	gendb "simiente-santa/backend/internal/db"
)

// repository implementa Repository (definida en service.go) sobre PostgreSQL,
// con las consultas generadas por sqlc desde internal/db/queries/contact.sql.
type repository struct {
	q    *gendb.Queries
	pool *pgxpool.Pool
}

// NewRepository construye el repository a partir del pool compartido.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{q: gendb.New(pool), pool: pool}
}

func (r *repository) Insert(ctx context.Context, c Contact) (Contact, error) {
	row, err := r.q.InsertContact(ctx, gendb.InsertContactParams{
		Name:    c.Name,
		Email:   c.Email,
		Phone:   c.Phone,
		Message: c.Message,
	})
	if err != nil {
		return Contact{}, fmt.Errorf("insert contact: %w", err)
	}
	return mapRow(row), nil
}

func (r *repository) List(ctx context.Context, f ListFilter) ([]Contact, error) {
	rows, err := r.q.ListContacts(ctx, gendb.ListContactsParams{
		FilterStatus: string(f.Status),
		Lim:          int32(f.Limit),
		Off:          int32(f.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	out := make([]Contact, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapRow(row))
	}
	return out, nil
}

func (r *repository) Count(ctx context.Context, f ListFilter) (int64, error) {
	total, err := r.q.CountContacts(ctx, gendb.CountContactsParams{FilterStatus: string(f.Status)})
	if err != nil {
		return 0, fmt.Errorf("count contacts: %w", err)
	}
	return total, nil
}

// mapRow traduce la fila generada por sqlc a la entidad de dominio.
func mapRow(row gendb.Contact) Contact {
	return Contact{
		ID:        row.ID,
		Name:      row.Name,
		Email:     row.Email,
		Phone:     row.Phone,
		Message:   row.Message,
		Status:    Status(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
```

Si la operación toca varias tablas, va en una transacción con `database.WithTx(ctx, pool, fn)` de
`internal/platform/database/`. Los errores de integridad que el dominio conoce (p. ej.
`unique_violation`) se traducen a `apperr.Conflict` **aquí**; el resto se envuelve con `%w` y
llega como `internal` al cliente.

### 5.7 `handler.go` — solo HTTP

Define la interfaz `Service` (quien la consume) y **nada de SQL**. Traduce el mundo HTTP al mundo
del dominio y viceversa; los errores salen todos por `httpserver.WriteError`.

```go
package contacto

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/validate"
)

// Service es la interfaz que este handler necesita del servicio (quien la consume).
// La implementa service.go.
type Service interface {
	Create(ctx context.Context, in CreateContactInput) (Contact, error)
	List(ctx context.Context, f ListFilter) ([]Contact, int64, error)
}

// Handler contiene solo HTTP: decodifica, valida, delega y responde.
type Handler struct {
	svc    Service
	logger *slog.Logger
}

// NewHandler construye el handler.
func NewHandler(svc Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// Create maneja POST /api/v1/contacto (público).
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var in CreateContactInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpserver.WriteError(ctx, w, h.logger, apperr.Invalid("el cuerpo debe ser JSON válido"))
		return
	}
	if err := validate.Struct(in); err != nil {
		httpserver.WriteError(ctx, w, h.logger, err) // 400 con detalle por campo
		return
	}

	c, err := h.svc.Create(ctx, in)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err) // único punto de traducción a HTTP
		return
	}
	httpserver.WriteJSON(w, http.StatusCreated, ContactCreatedResponse{ID: c.ID.String(), Received: true})
}

// List maneja GET /api/v1/admin/contactos (panel; el permiso lo exigió middleware.Authz).
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page := paginate.FromRequest(r, paginate.Defaults{Limit: 20, MaxLimit: 100})
	f := ListFilter{
		Status: Status(r.URL.Query().Get("status")),
		Limit:  page.Limit,
		Offset: page.Offset,
	}

	items, total, err := h.svc.List(ctx, f)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, toListResponse(items, total, page))
}

func toListResponse(items []Contact, total int64, p paginate.Params) ContactListResponse {
	out := make([]ContactItemResponse, 0, len(items))
	for _, c := range items {
		out = append(out, ContactItemResponse{
			ID:        c.ID.String(),
			Name:      c.Name,
			Email:     c.Email,
			Phone:     c.Phone,
			Message:   c.Message,
			Status:    string(c.Status),
			CreatedAt: c.CreatedAt,
		})
	}
	return ContactListResponse{Items: out, Total: total, Limit: p.Limit, Offset: p.Offset}
}
```

### 5.8 `routes.go` — públicas y de panel, por separado

```go
package contacto

import (
	"net/http"

	"simiente-santa/backend/internal/platform/httpserver"
)

// RegisterPublic publica las rutas PÚBLICAS del dominio (sin autenticación).
// Van separadas de RegisterAdmin para que quede auditable de un vistazo qué
// superficie expone cada dominio al mundo.
func RegisterPublic(r httpserver.Registrar, h *Handler) {
	r.Handle(http.MethodPost, "/api/v1/contacto", h.Create)
}

// RegisterAdmin publica las rutas del panel. El Registrar que se recibe YA viene
// envuelto por los middlewares de sesión y de permiso del módulo "contactos"
// (ver cmd/api/main.go): el dominio ni los menciona. Ruta completa:
// GET /api/v1/admin/contactos
func RegisterAdmin(r httpserver.Registrar, h *Handler) {
	r.Handle(http.MethodGet, "/contactos", h.List)
}
```

### 5.9 Cableado en `cmd/api/main.go` (fragmento)

Composición manual, de abajo arriba: repository → service → handler → rutas. Sin estado global.

```go
func main() {
	ctx := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	cfg, err := config.Load() // valida al arrancar y falla rápido si algo falta
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	appLog := logger.New(cfg.LogLevel) // slog JSON

	pool, err := database.NewPool(ctx, cfg.DatabaseURL) // pgxpool; no bloquea si la BD tarda
	if err != nil {
		appLog.Error("crear pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Migraciones: opt-in por entorno (F1 usa el CLI; ver platform/migrate).
	if cfg.RunMigrations {
		if err := migrate.Run(ctx, cfg.DatabaseURL, appLog); err != nil {
			appLog.Error("migrar", "error", err)
			os.Exit(1)
		}
	}

	mux := http.NewServeMux()
	root := httpserver.NewMuxRegistrar(mux)

	// dominio status (F1)
	status.RegisterPublic(root,
		status.NewHandler(status.NewService(status.NewRepository(pool)), appLog))

	// dominio contacto
	contactRepo := contacto.NewRepository(pool)
	contactSvc := contacto.NewService(contactRepo)
	contactH := contacto.NewHandler(contactSvc, appLog)

	contacto.RegisterPublic(root, contactH) // POST /api/v1/contacto

	admin := root.Group("/api/v1/admin",
		middleware.Authn(sessionStore),          // F2: cookie httpOnly → sesión
		middleware.AuthzByModule("contactos"),   // F2: permiso del módulo
		middleware.CSRF,                         // F2: métodos inseguros con sesión
	)
	contacto.RegisterAdmin(admin, contactH) // GET /api/v1/admin/contactos

	// Cadena global (fuera de todo grupo): request-id → recover → logging → CORS → rate-limit
	srv := httpserver.New(mux,
		httpserver.Options{Addr: cfg.HTTPAddr, ReadHeaderTimeout: 5 * time.Second, /* … */},
		appLog,
		middleware.RequestID, middleware.Recover(appLog), middleware.Logging(appLog), middleware.CORS(cfg.CORSAllowedOrigins),
	)

	if err := srv.Run(ctx); err != nil { // apagado ordenado con Shutdown
		appLog.Error("servidor", "error", err)
		os.Exit(1)
	}
}
```

Variables de entorno que toca este cableado (todas documentadas en `.env.example`, valores por
defecto de desarrollo): `HTTP_ADDR`/`HTTP_PORT`, `DATABASE_URL`, `LOG_LEVEL`,
`CORS_ALLOWED_ORIGINS`, `RUN_MIGRATIONS` (opt-in). Ningún secreto en el repo (constitución §IV).

### 5.10 Frontend del ejemplo

```text
frontend/src/
├── api/
│   ├── client.ts              # fetch tipado base (URL desde import.meta.env.VITE_API_URL)
│   ├── schema.d.ts            # GENERADO con openapi-typescript desde backend/api/openapi.yaml
│   └── contacto.ts            # funciones tipadas del dominio (único lugar que hace fetch)
└── features/
    └── contacto/
        ├── pages/
        │   ├── ContactoPage.tsx         # formulario público del visitante
        │   └── AdminContactosPage.tsx   # bandeja del panel
        ├── components/
        │   ├── ContactoForm.tsx         # React Hook Form + Zod, errores junto al campo
        │   └── ContactoTable.tsx
        ├── hooks/
        │   ├── useEnviarContacto.ts     # useMutation → api/contacto.ts
        │   └── useContactos.ts          # useQuery    → api/contacto.ts
        └── contacto.test.tsx            # Vitest + Testing Library + MSW
```

`frontend/src/api/contacto.ts` (los tipos salen del contrato, prohibido `any`):

```ts
import { apiFetch } from "./client";
import type { components } from "./schema";

export type ContactInput = components["schemas"]["ContactInput"];
export type ContactCreated = components["schemas"]["ContactCreated"];
export type ContactList = components["schemas"]["ContactList"];

export function createContact(input: ContactInput): Promise<ContactCreated> {
  return apiFetch<ContactCreated>("/api/v1/contacto", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function listContacts(params: {
  status?: "new" | "read" | "archived";
  limit?: number;
  offset?: number;
} = {}): Promise<ContactList> {
  const qs = new URLSearchParams();
  if (params.status) qs.set("status", params.status);
  if (params.limit != null) qs.set("limit", String(params.limit));
  if (params.offset != null) qs.set("offset", String(params.offset));
  return apiFetch<ContactList>(`/api/v1/admin/contactos?${qs}`, {
    credentials: "include", // la sesión viaja en cookie HttpOnly
  });
}
```

```ts
// frontend/src/features/contacto/hooks/useEnviarContacto.ts
import { useMutation } from "@tanstack/react-query";
import { createContact, type ContactInput } from "../../../api/contacto";

export function useEnviarContacto() {
  return useMutation({ mutationFn: (input: ContactInput) => createContact(input) });
}
```

Reglas que hereda el frontend (skill `react-frontend`): estados **cargando, vacío, error y éxito**
en cada vista con datos; validación de formularios con Zod espejando las restricciones del
contrato; accesibilidad (`label`, `button`, foco visible); i18n español/inglés en **F3**.

### 5.11 Errores: traducción a HTTP en un único punto

El dominio devuelve `apperr.*` (con el error interno envuelto con `%w`); **solo**
`httpserver.WriteError` decide el código de estado y el cuerpo de error. El handler no escribe
códigos de error a mano nunca.

```go
// internal/platform/apperr (esquema): Error{Kind, Code, Message, Details, Err}
//   Invalid → KindInvalid · NotFound → KindNotFound · Conflict → KindConflict
//   Unauthenticated · Forbidden · Internal (envuelve el error interno con %w)

// internal/platform/httpserver/error.go
// WriteError es el ÚNICO punto donde un error de dominio se traduce a HTTP.
func WriteError(ctx context.Context, w http.ResponseWriter, logger *slog.Logger, err error)
```

| `apperr` | HTTP | `error.code` | `error.message` | Detalle interno |
|---|---|---|---|---|
| `Invalid` | 400 | `invalid` | del dominio (seguro) | en log nivel `warn` |
| `Unauthenticated` | 401 | `unauthenticated` | del dominio | — |
| `Forbidden` | 403 | `forbidden` | del dominio | — |
| `NotFound` | 404 | `not_found` | del dominio | — |
| `Conflict` | 409 | `conflict` | del dominio | — |
| cualquier otro error | 500 | `internal` | `"Error interno del servidor"` (genérico) | **solo en log**, con `request_id` |

Sobre de error estándar (convención de la skill `go-backend`), con `details` opcional para
errores de validación por campo:

```json
{
  "error": {
    "code": "invalid",
    "message": "Datos inválidos",
    "details": { "email": "no tiene un formato de correo válido" }
  }
}
```

Nunca se exponen errores internos al cliente (stack, SQL, mensajes de la BD): eso va al log
estructurado con su `request_id`.

---

## 6. Flujo de una petición, paso a paso

Petición ejemplo: `POST /api/v1/contacto` con `{"name":"Ana","email":"ana@ejemplo.com","message":"Hola"}`.

1. **Despacho del router.** `http.ServeMux` (tras el adaptador `Registrar`) casa el patrón
   `"POST /api/v1/contacto"` con el handler ya envuelto por la cadena de middlewares.
2. **Cadena global** (la monta `httpserver.New`, en este orden; el primero es el más externo):

   ```text
   request-id → recover → logging → CORS → rate-limit → [authn → authz → CSRF] → handler
   ```

   - `request-id`: toma `X-Request-ID` del cliente o genera uno; lo guarda en el `context` y crea
     el **logger por petición** (hij de `slog` con `request_id`, método y ruta).
   - `recover`: captura cualquier `panic` de la cadena interna y responde 500 vía `WriteError`
     (nunca se cae el proceso por una petición).
   - `logging`: registra al terminar método, ruta, status, duración y `request_id`.
   - `CORS`: responde los preflight `OPTIONS` y fija los headers permitidos
     (`CORS_ALLOWED_ORIGINS`); corta ahí los preflight.
   - `rate-limit`: satura a la baja peticiones abusivas (429); con el primer endpoint público
     escribible.
   - `[authn → authz → CSRF]`: solo en **grupos** (p. ej. `/api/v1/admin`), añadidos con
     `Group(...)` — el dominio no los ve.
3. **Handler** (`handler.go`): decodifica el JSON a `CreateContactInput`; si el cuerpo no es JSON
   válido → `WriteError(apperr.Invalid(...))` y se acabó. Valida con `platform/validate`
   (etiquetas del DTO) → 400 con `details` si falla. Llama a `Service.Create(ctx, in)`.
4. **Service** (`service.go`): normaliza (espacios, email en minúsculas, teléfono), comprueba sus
   invariantes y llama a `Repository.Insert(ctx, contact)` con `ctx` como primer parámetro.
5. **Repository** (`repository.go`): ejecuta la consulta **parametrizada** generada por sqlc sobre
   el pool `pgx`. Envuelve cualquier error con `fmt.Errorf("insert contact: %w", err)`.
6. **Respuesta de éxito**: `httpserver.WriteJSON(w, 201, ContactCreatedResponse{…})`
   (`Content-Type: application/json`).
7. **Respuesta de error**: el error sube envuelto hasta el handler, que hace
   `httpserver.WriteError(ctx, w, h.logger, err)` — único punto de traducción (§5.11).
8. **Cierre**: `logging` escribe la línea estructurada; si el cliente se desconectó, `ctx` ya está
   cancelado y las capas internas lo respetan.

En las rutas del panel, entre 2 y 3 ocurren `authn` (cookie `httpOnly` → sesión, F2), `authz`
(permiso del módulo `"contactos"`, F2) y `CSRF` (F2); un fallo en cualquiera de los tres responde
401/403 y **nunca llega al handler**.

---

## 7. Pruebas por capa

| Capa | Tipo | Herramienta | Qué verifica | Dónde |
|---|---|---|---|---|
| `service` | Unitaria | `testing` + **fake** de la interfaz `Repository` | Reglas de negocio con tabla de casos (`tests := []struct{...}` + `t.Run`): normalización, límites, errores `apperr`, `%w` envuelto | `internal/<dominio>/service_test.go` |
| `handler` | Unitaria | `httptest.NewRecorder` + **fake** de `Service` | Decodificación, validación, códigos de estado, sobre de error, contenido de la respuesta | `internal/<dominio>/handler_test.go` |
| `repository` | **Integración** | PostgreSQL real (`DATABASE_URL_TEST`), `//go:build integration` | SQL real, restricciones de la tabla, orden/paginación, traducción de errores | `internal/<dominio>/repository_test.go` |
| `platform` | Unitaria | tabla de casos + `httptest` para middlewares | `apperr`, `validate`, `paginate`, cadena de middlewares, `Registrar` (incluido `Group`) | `internal/platform/*/**_test.go` |
| Frontend | Unitaria | Vitest + Testing Library + **MSW** | Comportamiento visible: estados cargando/vacío/error/éxito, formularios con errores de campo | `frontend/src/features/<feature>/*.test.tsx` |
| E2E | End-to-end | **Playwright** | Flujos críticos contra el stack levantado (`make up`) | `frontend/e2e/*.spec.ts` |

Comandos: `go test ./...` (unitarias) · `go test -tags=integration ./...` (integración;
*skip* si no hay `DATABASE_URL_TEST`) · `npm test -- --run` · `npx playwright test` ·
`make ci` (el veredicto completo: lint, pruebas, migraciones, `govulncheck`, `npm audit`).

Cobertura mínima exigida: **80 % en `service/`** (constitución §III); la verifican `qa-tester` y
`revisor-codigo`. Ningún cambio se integra con pruebas fallando.

---

## 8. Receta: cómo agregar un área de negocio (10 pasos)

1. **Contrato primero.** Redacta el delta (paths, esquemas, códigos de error) en
   `specs/<N>-<feature>/contracts/` y fúndelo en `backend/api/openapi.yaml` **antes** de escribir
   código (constitución §II). Regenera los tipos del frontend: `npm run api:gen`.
2. **Migración nueva** en `backend/migrations/`: `00000N_create_<tabla>.up.sql` + `.down.sql`
   completo. **Nunca** edites una migración aplicada. Ejecuta `make db-migrate`.
3. **Consultas sqlc** en `internal/db/queries/<dominio>.sql`, siempre parametrizadas, sin
   `SELECT *`; luego `sqlc generate` (código generado se commitea).
4. **`model.go`**: entidad, estados, DTOs de entrada/salida con sus etiquetas de validación y
   límites (espejo del contrato).
5. **`repository.go`**: implementación concreta sobre `internal/db` + `pgxpool`, errores envueltos
   con `%w`; integridad conocida → `apperr.Conflict`. (La interfaz la escribe el paso 6: si
   compilas antes, deja el tipo pendiente y complétalo con el compilador de guía.)
6. **`service.go`**: la interfaz `Repository` (la define quien consume) + las reglas de negocio.
   Sin HTTP, sin SQL. Inyecta el reloj si se usan fechas.
7. **`handler.go`**: la interfaz `Service` (la define quien consume) + decodificar, validar,
   delegar y responder. **Todos** los errores por `httpserver.WriteError`; ningún código de error
   escrito a mano.
8. **`routes.go`**: `RegisterPublic(...)` y `RegisterAdmin(...)` por separado, para que quede
   auditable qué es público y qué exige permiso.
9. **Cableado en `cmd/api/main.go`**: `NewRepository(pool)` → `NewService(repo)` →
   `NewHandler(svc, logger)` → `RegisterPublic`/`RegisterAdmin(adminGroup, h)` (el grupo lleva
   `authn`, `authz` por módulo y `CSRF`).
10. **Pruebas de las tres capas** (service con fake, handler con `httptest`, repository con
    `//go:build integration`) + frontend con MSW si hay UI + `make ci` en verde.

El PR de cada tarea lleva sus pruebas (constitución §III) y pasa por `qa-tester`,
`revisor-codigo` y `seguridad` antes de integrarse (quien escribe no aprueba).
