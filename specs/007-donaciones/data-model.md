# Data Model — F7 Ofrendas y diezmos (Donaciones)

**Fecha**: 2026-10-14 · **Spec**: [spec.md](./spec.md) · **Decisiones**: [research.md](./research.md)
(R7-1…R7-12) · **Plantilla de estilo**: el `data-model.md` de F6 (que sigue el de F4).

Fuente de verdad del esquema: **este documento** → migraciones `backend/migrations/` → consultas
`backend/internal/db/queries/donation_methods.sql` y `donation_texts.sql` (sqlc). Convenciones:
skill `postgres-db`, §VI de la constitución y §8.1 de `docs/tecnico/arquitectura.md` (UUID,
`created_at`/`updated_at`, mapeo `pgtype` → dominio en el repository, consultas parametrizadas).

## Patrones transversales heredados de F3/F4/F6 (no se reabren)

- **Contenido bilingüe (R3-1/R7-8)**: cada campo traducible vive en dos columnas, `*_es` (idioma
  base) y `*_en` (**opcional**). `""`/espacios se guardan como `NULL` (normalización I6: trim en el
  service antes de persistir). **La regla es por campo** (Decisión 8): si un campo **opcional** se
  rellena en algún idioma, su versión en español debe existir (`*_en IS NULL OR *_es IS NOT NULL`,
  con `CHECK` en cada par). Traducibles en F7: `holder` (titular) y `body` de los textos. **No
  traducibles**: `kind`, `number`, `institution` (R7-5) y `currency` (R7-4).
- **Estado de publicación (R3-3/R7-11)**: `publication_state TEXT NOT NULL DEFAULT 'draft' CHECK
  (publication_state IN ('draft','published'))` **por método** y **por texto** (Q10). Las consultas
  del visitante filtran `publication_state = 'published'` en la **misma** consulta (`…Published`):
  es la única vía pública y no existe ninguna función que devuelva al visitante un borrador
  (FR-012/SC-002).
- **Orden manual (FR-004)**: `sort_order INTEGER NOT NULL DEFAULT 0` **dentro de cada grupo de
  tipo** (Q1), con desempate por `id`, `created_at` NO decide (patrón literal de `000005`,
  `000007` y `000010`). El orden completo es `CASE kind WHEN 'iban' THEN 0 ELSE 1 END, sort_order,
  id` — primero cuentas IBAN y luego SINPE Móvil, como exigen FR-004/SC-014.
- **Auditoría (R3-11/R4-10/R6-12/R7-10)**: las mutaciones insertan `admin_actions` con
  `InsertAdminAction` (consulta compartida de F2) **dentro de su transacción**; objetivo
  `target_kind='content'`, ambas FK en `NULL` y `target_label` descriptivo. Registro de solo
  lectura (sin `UPDATE`/`DELETE`).
- **Normalización (F2/P13)**: "el dato normalizado es el que se guarda". F7 normaliza **los números**
  (IBAN a los caracteres del número; SINPE a `+506XXXXXXXX`, R7-2/R7-3) y recorta los textos
  (`btrim` en el service; `CHECK` en las columnas cortas, igual que `000010`).

## Resumen de tablas

| Tabla | Cardinalidad | Qué guarda |
|---|---|---|
| `donation_methods` (**nueva**) | 0..N | Un método de donación (cuenta IBAN o número de SINPE Móvil): `kind`, `number` (normalizado), `institution` (obligatorios), `holder` opcional y traducible, `currency` solo en IBAN, `sort_order` **por grupo** y `publication_state` (FR-001, FR-019, Q1/Q2/Q3/Q4/Q10) |
| `donation_texts` (**nueva**) | 0..2 | Los dos textos libres de la sección —`intro` («cómo donar», Q11) y `usage` («en qué se usan las donaciones», Q5)— uno por clave (`text_key UNIQUE`): `body_es`/`body_en` y su **propio** `publication_state` (FR-021, Q5/Q10/Q11) |
| `admin_actions` (**existente, F2**) | — | Se **amplía** el registro cerrado: 8 códigos `donation.*` (FR-017, R7-10) |
| `permissions` (**existente, F2**) | — | **SIN CAMBIOS**: `('donaciones', 'Donaciones')` ya está sembrado por `000002` (línea 25); F7 **no** crea permisos nuevos (FR-011) |
| `roles` / `role_permissions` (**existentes, F2**) | — | **SIN CAMBIOS** (Decisión 9) |

Ninguna tabla de F7 tiene FK hacia `users`: el "quién" vive en `admin_actions.actor_user_id`
(FR-017). **Ninguna FK nueva**: los textos no dependen de los métodos ni al revés (son dos
superficies de la misma sección) y F7 no tiene catálogos. **F7 no toca `media`** ni
`IsFilePublished`: Q6/FR-020 prohíben imágenes, así que no hay archivos que servir. Índices:
`(kind, sort_order, id)` para el orden público y de panel, y `publication_state` para el
filtro/lectura pública.

---

## Migraciones

Dos migraciones nuevas (F6 terminó en `000011`): **`000012`** crea `donation_methods` y
`donation_texts`; **`000013`** amplía el registro de auditoría con los 8 códigos `donation.*`.
**No hay migración de permisos** (el permiso `donaciones` ya existe) y ninguna migración existente
se toca (§VI).

### `000012_create_donation_content.up.sql`

```sql
-- 000012_create_donation_content (up)
-- F7 Ofrendas y diezmos: el listado administrable de métodos de donación
-- (cuentas IBAN y números de SINPE Móvil, agrupados por tipo con orden manual
-- dentro de cada grupo — Q1) y los dos textos libres de la sección, uno por
-- clave, cada uno con su propio estado de publicación (Q5/Q10/Q11).
-- Convenciones: skill `postgres-db` y specs/007-donaciones/data-model.md
-- (fuente de verdad). Decisiones: research R7-1…R7-5.

CREATE TABLE donation_methods (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Tipo de método (Q1/FR-019): 'iban' = cuenta bancaria, 'sinpe' = SINPE Móvil.
    kind              TEXT NOT NULL CHECK (kind IN ('iban', 'sinpe')),
    -- Número ya NORMALIZADO (research R7-2/R7-3): el dato normalizado es el que
    -- se guarda. IBAN = solo los caracteres del número, en mayúsculas
    -- (sin espacios ni guiones; longitud exacta por país la valida el service).
    -- SINPE = '+506' + 8 dígitos. El CHECK por tipo refuerza la forma aunque el
    -- service ya la valida (defensa en profundidad).
    number            TEXT NOT NULL CHECK (char_length(number) BETWEEN 1 AND 34),
    -- Entidad financiera (nombre del banco), obligatoria (Q2). NO traducible
    -- (FR-007/US4 esc. 5/SC-006; research R7-5 señala la contradicción de FR-019).
    institution       TEXT NOT NULL CHECK (char_length(institution) BETWEEN 1 AND 160),
    -- Titular de la cuenta, OPCIONAL (Q2: "si se rellena, se publica") y único
    -- campo traducible del método (FR-007).
    holder_es         TEXT NULL CHECK (char_length(holder_es) BETWEEN 1 AND 160),
    holder_en         TEXT NULL CHECK (char_length(holder_en) BETWEEN 1 AND 160),
    -- Moneda SOLO en cuentas IBAN (Q2/FR-014): obligatoria en IBAN y prohibida
    -- en SINPE (los dos CHECK de abajo). Código CRC/USD; la etiqueta visible la
    -- localiza el frontend (patrón day_of_week de 000005; no es contenido
    -- traducible, SC-006).
    currency          TEXT NULL CHECK (currency IN ('CRC', 'USD')),
    publication_state TEXT NOT NULL DEFAULT 'draft'
                      CHECK (publication_state IN ('draft', 'published')),
    -- Orden manual DENTRO del grupo de tipo (Q1/FR-004); el orden entre grupos
    -- lo fija el CASE del ORDER BY (iban primero, sinpe después).
    sort_order        INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Formato exacto por tipo (el dato normalizado es el que se guarda):
    --   IBAN  → 2 letras de país + 2 dígitos de control + BBAN alfanumérico
    --            (la longitud exacta por país la valida el service: R7-2)
    --   SINPE → '+506' + 8 dígitos (celular costarricense: R7-3)
    CHECK ((kind = 'iban'  AND number ~ '^[A-Z]{2}[0-9]{2}[A-Z0-9]{10,30}$')
        OR (kind = 'sinpe' AND number ~ '^\+506[0-9]{8}$')),
    -- Moneda obligatoria en IBAN y prohibida en SINPE (Q2/FR-014).
    CHECK ((kind = 'iban' AND currency IS NOT NULL)
        OR (kind = 'sinpe' AND currency IS NULL)),
    -- Español base por campo: holder_en exige holder_es (FR-007/Decisión 8).
    CHECK (holder_en IS NULL OR holder_es IS NOT NULL),
    -- Normalización de F2 (P13): el dato guardado ya viene recortado
    -- (precedente literal de 000010; los anulables con IS NULL OR … = btrim(…)).
    CHECK (institution = btrim(institution)),
    CHECK (holder_es IS NULL OR holder_es = btrim(holder_es)),
    CHECK (holder_en IS NULL OR holder_en = btrim(holder_en))
);

-- Orden de presentación dentro del grupo y filtro público por estado.
CREATE INDEX donation_methods_kind_sort_order_id_idx
    ON donation_methods (kind, sort_order, id);
CREATE INDEX donation_methods_publication_state_idx
    ON donation_methods (publication_state);

-- Los dos textos libres de la sección (FR-021, Q5/Q11): un renglón por clave
-- ('intro' = «cómo donar», 'usage' = «en qué utilizamos esas donaciones»),
-- con su PROPIO estado de publicación (Q10). Es el patrón "singleton por clave":
-- UNIQUE (text_key) garantiza ≤1 fila por clave y las filas se crean en el
-- primer guardado (upsert del PUT), como los singletons de 000005; la migración
-- NO siembra contenido.
CREATE TABLE donation_texts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text_key          TEXT NOT NULL UNIQUE CHECK (text_key IN ('intro', 'usage')),
    body_es           TEXT NOT NULL CHECK (char_length(body_es) BETWEEN 1 AND 2000),
    body_en           TEXT NULL CHECK (char_length(body_en) BETWEEN 1 AND 2000),
    publication_state TEXT NOT NULL DEFAULT 'draft'
                      CHECK (publication_state IN ('draft', 'published')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Español base por campo (FR-007/Decisión 8): body_en exige body_es
    -- (redundante con body_es NOT NULL; se deja como documentación viva de la
    -- regla, igual que los CHECK de español por campo de 000010).
    CHECK (body_en IS NULL OR body_es IS NOT NULL)
);
```

### `000012_create_donation_content.down.sql`

```sql
-- 000012_create_donation_content (down) — revierte 000012 por completo (§VI).
DROP TABLE IF EXISTS donation_texts;
DROP TABLE IF EXISTS donation_methods;
```

### `000013_extend_admin_actions_for_donations.up.sql`

Amplía el registro de auditoría de F2 (FR-017/R7-10). El `CHECK` que se sustituye es
`admin_actions_action_check`, **con nombre propio** desde `000006` de F3 (F3, F4 y F6 ya hicieron
este mismo ejercicio y su prueba de integración lo verifica con up → down → up). El `target_kind`
**no cambia**: los métodos y los textos son `content` como los elementos de F3/F4/F6. Los **8
códigos** nuevos son los de R7-10.

```sql
-- 000013_extend_admin_actions_for_donations (up)
-- F7: la auditoría de F2 se amplía a ofrendas y diezmos (FR-017). El registro de
-- códigos de acción sigue siendo CERRADO y lo garantiza la BD.

ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_action_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_action_check CHECK (action IN (
    'user.create', 'user.update', 'user.activate', 'user.deactivate', 'user.password_reset',
    'role.create', 'role.update', 'role.delete',
    'home.identity.update', 'home.about.update', 'home.contact.update',
    'home.schedule.create', 'home.schedule.update', 'home.schedule.delete',
    'home.whatsapp.create', 'home.whatsapp.update', 'home.whatsapp.delete',
    'home.social.create', 'home.social.update', 'home.social.delete',
    'home.image.upload',
    'home.publish', 'home.unpublish',
    'event.create', 'event.update', 'event.delete', 'event.publish', 'event.unpublish',
    'event.image.upload',
    'activity.create', 'activity.update', 'activity.delete', 'activity.publish',
    'activity.unpublish', 'activity.image.upload',
    'category.create', 'category.update', 'category.delete',
    'ministry.create', 'ministry.update', 'ministry.delete', 'ministry.publish',
    'ministry.unpublish', 'ministry.leader.show', 'ministry.leader.hide', 'ministry.image.upload',
    'donation.create', 'donation.update', 'donation.delete', 'donation.publish',
    'donation.unpublish', 'donation.text.update', 'donation.text.publish', 'donation.text.unpublish'
));
```

### `000013_extend_admin_actions_for_donations.down.sql`

```sql
-- 000013_extend_admin_actions_for_donations (down): restaura el CHECK de F6
-- (8 códigos de F2 + 15 de F3 + 15 de F4 + 8 de F6). Como en el down de
-- 000006/000008/000011, es destructivo a propósito: las filas con códigos F7 no
-- caben en el CHECK anterior y se eliminan antes de restaurar. Solo tiene
-- sentido en desarrollo/rollback.
DELETE FROM admin_actions WHERE action IN (
    'donation.create', 'donation.update', 'donation.delete', 'donation.publish',
    'donation.unpublish', 'donation.text.update', 'donation.text.publish', 'donation.text.unpublish'
);

ALTER TABLE admin_actions DROP CONSTRAINT admin_actions_action_check;
ALTER TABLE admin_actions ADD CONSTRAINT admin_actions_action_check CHECK (action IN (
    'user.create', 'user.update', 'user.activate', 'user.deactivate', 'user.password_reset',
    'role.create', 'role.update', 'role.delete',
    'home.identity.update', 'home.about.update', 'home.contact.update',
    'home.schedule.create', 'home.schedule.update', 'home.schedule.delete',
    'home.whatsapp.create', 'home.whatsapp.update', 'home.whatsapp.delete',
    'home.social.create', 'home.social.update', 'home.social.delete',
    'home.image.upload',
    'home.publish', 'home.unpublish',
    'event.create', 'event.update', 'event.delete', 'event.publish', 'event.unpublish',
    'event.image.upload',
    'activity.create', 'activity.update', 'activity.delete', 'activity.publish',
    'activity.unpublish', 'activity.image.upload',
    'category.create', 'category.update', 'category.delete',
    'ministry.create', 'ministry.update', 'ministry.delete', 'ministry.publish',
    'ministry.unpublish', 'ministry.leader.show', 'ministry.leader.hide', 'ministry.image.upload'
));
```

**Notas de diseño.**

- **`number` guardado normalizado** (R7-2/R7-3): lo que entra al formulario (con espacios, guiones
  y `+506` opcional) se normaliza **antes** de persistir y **ese** valor es el que se publica
  (FR-015: "el número publicado DEBE ser el número válido guardado"). Por eso la comparación de
  duplicados (R7-9) es una igualdad simple y el `CHECK` por tipo puede ser estricto.
- **`institution` sin par `*_es`/`*_en`** (R7-5): la entidad financiera **no** es contenido
  traducible (FR-007/US4 esc. 5/SC-006). La contradicción con FR-019/Key Entities está señalada
  en research R7-5 y en «Preguntas para el humano» de plan.md; si el humano decide lo contrario,
  el cambio es `institution` → `institution_es`/`institution_en` con su `CHECK` de español por
  campo (migración nueva; nada más se rompe).
- **`currency` con doble `CHECK` sobre `kind`** (R7-4): la moneda es obligatoria en IBAN y **no
  aplica** en SINPE (Q2). La BD lo garantiza aunque el formulario no ofrezca el campo en SINPE.
- **`sort_order` es por grupo** (Q1/FR-004): no hay una lista de orden aparte; dos métodos de
  grupos distintos pueden tener el mismo `sort_order` sin conflicto porque el orden entre grupos lo
  fija el `CASE kind`. El panel edita `sortOrder` como un campo más del `PATCH` y el visitante ve
  exactamente ese orden (US2 esc. 10).
- **`donation_texts` sin `sort_order`** (R7-1): los dos textos tienen posición fija en la sección
  (la composición la define `ux.md`, FR-021) y no se reordenan. Tampoco hay operación de borrado:
  un texto se retira con `publication_state = 'draft'` (FR-010 solo habla de "mantener" los
  textos; eliminarlos no está en la spec). **Si el humano quisiera borrarlos**, sería un
  `DELETE` nuevo en el contrato (no se inventa).
- **Sin `UNIQUE (kind, number)`** (R7-9): el duplicado exacto lo rechaza el service con
  `409 conflict` (mismo gesto que F6/R6-11), y así el default revisable "se rechaza **o se
  señala**" sigue siendo una línea que cambiar. Si el humano prefiere blindar la BD frente a dos
  altas simultáneas, `CREATE UNIQUE INDEX … ON donation_methods (kind, number)` es la adición de
  la migración siguiente.
- **Sin `deleted_at`**: la eliminación es definitiva, con confirmación y registrada (default
  revisable de la spec, igual que F3/F4/F6).
- **`CHECK` de `btrim`**: los textos cortos se guardan ya recortados (normalización de F2/P13,
  precedente `000010`); los cuerpos largos (`body_es`/`body_en`) no llevan `CHECK` de `btrim` —
  igual que `description_es` de `000010`—, pero el service los recorta igual (I6).
- **Sin imágenes** (Q6/FR-020): ninguna columna de archivo, ninguna FK a `media` y **cero
  cambios** en `IsFilePublished` (a diferencia de F6/R6-7).

---

## Consultas sqlc

### `backend/internal/db/queries/donation_methods.sql` (dominio `donaciones`, nuevo)

Listas con `ORDER BY` explícito y determinista (§8.1.7):
`ORDER BY CASE kind WHEN 'iban' THEN 0 ELSE 1 END, sort_order, id`. Sin `SELECT *`. Todo
parametrizado (§IV, CWE-89). **Separación público/panel en consultas distintas**: la única función
que ve borradores es la del panel (FR-012).

| Consulta | Uso |
|---|---|
| `ListPublishedDonationMethods` (`:limit`) | **Público** (sección y bloque de la portada): solo `publication_state = 'published'`, orden de sección, `LIMIT :limit`. El límite lo pasa el service (`maxPublicItems = 200` en la sección, `maxPortadaItems = 5` en el bloque): el tope 5 del bloque es de servidor (FR-022) |
| `ListDonationMethods` (`:limit`, `:offset`, `sqlc.narg('publication_state')`) | Panel paginado: **todo** (incluidos borradores), el **mismo orden que ve el visitante** (US2 esc. 10), filtro opcional por estado |
| `CountDonationMethods` (`sqlc.narg('publication_state')`) | Total del listado del panel (sobre de paginación) |
| `GetDonationMethodByID` | Panel (detalle y respuestas de escritura): todas las columnas |
| `InsertDonationMethod` / `UpdateDonationMethod` / `DeleteDonationMethod` | Mutaciones del panel (el `Update` escribe **la fila completa ya fusionada por el service** — campos presentes del `PATCH` sobre la fila actual, ausentes intactos y opcionales como `NULL` cuando se borran: R7-15) |
| `ExistsDonationMethodDuplicate` (`:kind`, `:number`) | Duplicado (mismo tipo y mismo número **normalizado**, R7-9) → `409 conflict` |

### `backend/internal/db/queries/donation_texts.sql` (dominio `donaciones`, nuevo)

| Consulta | Uso |
|---|---|
| `ListPublishedDonationTexts` | **Público**: solo `publication_state = 'published'`, `ORDER BY text_key` (determinista; `intro` < `usage` alfabéticamente y la composición la decide el cliente); el service las empaqueta como `intro`/`usage` |
| `ListDonationTexts` | Panel: las dos filas (existentes o no) en ambos idiomas con su estado |
| `GetDonationTextByKey` (`:text_key`) | Panel (detalle y respuesta del `PUT`) |
| `UpsertDonationText` (`:text_key`, …) | `INSERT … ON CONFLICT (text_key) DO UPDATE …` (crea o reemplaza el singleton; patrón F3 de `PUT /admin/portada/quienes-somos`) + `updated_at = now()` |

La auditoría **reutiliza** `InsertAdminAction` (F2): el repository de `donaciones` la llama con
`gendb.New(tx)` dentro de la transacción de cada mutación. No se escribe ningún `UPDATE`/`DELETE`
sobre `admin_actions` (solo lectura).

## Escritura de la auditoría desde F7 (detalle)

- Mutación y fila de `admin_actions` comparten transacción (o ambos o ninguno). Si el INSERT del
  registro falla, **no se aplica** la operación (edge case "no aplicarse sin registro").
- Objetivo: `target_kind='content'`, `target_user_id`/`target_role_id` = NULL, `target_label` =
  `Donaciones · <institution> · <number>` (métodos) o `Donaciones · Texto introductorio` /
  `Donaciones · Texto de uso` (textos; máx. 254, `CHECK` de F2).
- `action` ∈ los 8 códigos nuevos (regla exacta en R7-10): `donation.create` en el alta —**una
  sola fila**, también con `publicationState: published` (el alta **nunca** emite
  `donation.publish`/`unpublish`)—; `donation.publish`/`donation.unpublish` **solo** para cambios
  de estado de un método; `donation.update` para cualquier otro cambio de datos (incluido
  `sortOrder`); datos + estado en la **edición** → **dos filas**; `donation.text.update` en el
  guardado de un texto (alta o edición del singleton) y `donation.text.publish`/
  `donation.text.unpublish` **solo** para cambios de estado de un texto. `result`: `success` en la
  transacción, `failure` en rechazos y `denied` en denegaciones (best-effort).
- `actor_user_id` siempre la cuenta del panel que ejecutó (el `CHECK` de F2 sigue vigente).

## Invariantes de negocio y dónde se verifican

| Invariante | Dónde |
|---|---|
| Obligatorios de un método: `kind`, `number`, `institution` (y `currency` si es IBAN); **el titular es opcional** (Q2) | `required` en los DTOs (`platform/validate`) + `NOT NULL` en la tabla + regla condicional del service |
| El SINPE **no** lleva moneda y la moneda del IBAN es obligatoria (Q2/FR-014) | `CHECK` doble sobre `kind` + `400 invalid` con `details` en el service |
| IBAN por estructura y longitud (costarricense e internacional), guardado normalizado (Q3/FR-015) | `internal/donaciones/numero.go` (regex + tabla de longitudes por país) + `CHECK` por tipo + trim/normalización antes de persistir |
| SINPE = celular costarricense de 8 dígitos, con o sin `+506`, prefijo normalizado (Q4/FR-016) | `numero.go` (8 dígitos, primer dígito 6/7/8) + `CHECK '^\+506[0-9]{8}$'` |
| La entidad financiera **no** cambia con el idioma (FR-007/SC-006) | Una sola columna `institution` (R7-5) + prueba "idéntico en es y en" |
| Español por campo: `holder_en` exige `holder_es`, `body_en` exige `body_es` (FR-007) | `CHECK` por par + regla del service (el formulario la refleja) |
| Solo lo publicado es visible al visitante (FR-012), **también por texto** (Q10) | Consultas `…Published` (única vía pública) + pruebas "0 borradores expuestos" (ni métodos ni textos) |
| Orden `iban → sinpe` y, dentro de cada grupo, `sort_order, id` (FR-004/SC-014) | `ORDER BY CASE kind … END, sort_order, id` + índice `donation_methods_kind_sort_order_id_idx` + prueba de orden |
| Tope **5** del bloque, resuelto en el servidor (FR-022) | `maxPortadaItems = 5` (constante del service, con prueba) + `LIMIT :limit` en SQL; sin parámetro de límite en el contrato |
| Duplicado exacto rechazado (mismo tipo y mismo número normalizado, R7-9) | `ExistsDonationMethodDuplicate` → `409 conflict` |
| Auditoría atómica con la mutación | Transacción única (`InsertAdminAction` sobre `tx`) + prueba de integración |
| Registro de auditoría de solo lectura | Sin consultas `UPDATE`/`DELETE` sobre `admin_actions` en `queries/` |
| Permisos: `donaciones` ya existe y **no cambia** el catálogo de F2 | Sin migración de permisos; `000002` no se toca; `AdminChain("donaciones")` en `routes.go` |
| Sin imágenes: nada que servir ni que filtrar (Q6/FR-020) | Sin columnas de archivo; `media`/`IsFilePublished` **sin cambios** |

## Validación del modelo (qué comprobará la implementación)

1. `migrate up` limpio desde F6 (`000011`) y `down` completo de `000012`/`000013` sin errores
   (incluida la restauración del `CHECK` de `admin_actions`: up → down → up).
2. Pruebas de integración del repository (PostgreSQL real, `//go:build integration`):
   - el catálogo de `permissions` **sigue en 10 filas** y `donaciones` entre ellas (F7 no lo toca;
     el `down` de `000012`/`000013` no lo roza);
   - `CHECK` de formato por tipo: un IBAN con espacios (`'CR01 000…'`) falla en la BD; un
     `'cr01…'` en minúsculas falla; un SINPE sin `+506` (`'88888888'`) falla; un
     `'+50688888888'` procede;
   - `CHECK` de moneda: una cuenta `iban` sin `currency` falla; una `sinpe` **con** `currency`
     falla; `iban` + `CRC` y `sinpe` sin moneda proceden;
   - `CHECK` de español por campo: `holder_en` sin `holder_es` falla (ídem `body_en` sin `body_es`
     —este es imposible por `body_es NOT NULL`—);
   - `CHECK` de `btrim`: `institution`/`holder_es`/`holder_en` con espacios sobrantes fallan;
   - `text_key` fuera del catálogo (`'otro'`) falla y `UNIQUE (text_key)` impide dos renglones de
     la misma clave (el `UpsertDonationText` actualiza en vez de duplicar);
   - `ListPublishedDonationMethods`: **nunca** devuelve borradores; el orden es
     `CASE kind …, sort_order, id` (iban primero) y el `LIMIT` se respeta (5 en el bloque);
   - `ListPublishedDonationTexts`: solo textos publicados;
   - `ExistsDonationMethodDuplicate` detecta el repetido tras normalizar (`'cr01 0000-…'` vs
     `'CR010000…'`) y **no** dispara por el mismo número en otro tipo (imposible por forma, pero
     la consulta es `kind + number`);
   - cada mutación deja su fila en `admin_actions` en la misma transacción (y al forzar un fallo en
     el registro, la mutación no se aplica), con los 8 códigos y `target_label` correctos;
   - inserción de `admin_actions` con los 8 códigos dentro del `CHECK` ampliado (y rechazo de un
     código fuera del registro).
3. Tipos sqlc: `pgtype.*` traducido en `repository.go` (§8.1.6); `internal/db` regenerado y
   commiteado (`make sqlc-verify`).
