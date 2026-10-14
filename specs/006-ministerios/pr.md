# PR — F6 Ministerios

> Descripción para el Pull Request de la rama `006-ministerios` (preparada por `documentador`,
> 2026-10-13; T638). **El merge y el despliegue requieren aprobación humana explícita** (fase 9/9
> del flujo `equipo-feature`; puertas «PR / merge» y «Despliegue» en `estado.md`). El `documentador`
> de esta sesión no tenía acceso a shell ni a `gh`: quien abre el PR es `devops`/el humano con esta
> descripción.

## feat: F6 — Ministerios (sección pública, panel y bloque de la portada)

### Resumen

La segunda sección de contenido dinámico del sitio: **las áreas de servicio de la iglesia**, visibles
sin cuenta, con el módulo del panel que las administra y su extracto en la portada. Cubre las **38
tareas (T601–T638)** de `tasks.md` (US1–US6 / FR-001…FR-029 / SC-001…SC-017 de la **spec v2**;
Q1–Q9 resueltos el 2026-10-13 y dos decisiones propias del humano en la puerta 2).

**Backend (Go + PostgreSQL, módulo `internal/ministerios`)**
- Documentos públicos **`GET /api/v1/ministerios?lang=es|en`** (sección completa) y
  **`GET /api/v1/portada/ministerios?lang=es|en`** (extracto del bloque, **tope 5 resuelto en el
  servidor**, `Cache-Control: no-store`, clave `items` omitida si está vacía, `lang` cerrado a
  `es|en` y DTO mínimo **sin campos de encargado**), ambos sin sesión y con fallback `en → es`.
- Subgrupo `/api/v1/admin/ministerios` (lista, alta, `GET`/`PATCH`/`DELETE` por id e imagen
  multipart) tras `AdminChain("ministerios")` + `RequirePermission`; contrato vivo
  **0.6.0 → 0.7.0** (`backend/api/openapi.yaml`, fusión aditiva del delta `contracts/openapi.yaml`,
  que no se edita) y tipos TS regenerados con `make api-gen`.
- **Privacidad del encargado en tres capas** (FR-026/SC-014): la consulta pública enmascara con
  `CASE WHEN leader_visible` (la fila **no transporta** el dato), el repositorio traduce con
  `maskedText` *fail-closed* y el DTO público solo compone `leader` si está publicado; el extracto
  de la portada **no tiene campos de encargado en ningún caso**. `leader_visible` nace en `FALSE`.
- **Regla de publicación** como filtro de lectura (borradores jamás expuestos, también en la imagen
  vía `IsFilePublished` ampliado de forma **aditiva** a `ministries`); **redirección silenciosa** de
  `/ministerios` sin publicados en el cliente (la API responde `200` sin `items` y no redirige).
- Validaciones completas en servidor: español por campo, solo nombre y descripción en español
  obligatorios, celular **o** correo (nunca ambos; el contacto exige nombre), `whatsappUrl` solo
  `https` seguro de cualquier dominio, duplicado exacto → `409`, `PATCH {}` → `400`
  (`minProperties: 1`) y `sortOrder` acotado a `int32` → `400` con detalle **sin truncar**.
- Migraciones **`000010_create_ministries`** y **`000011_extend_admin_actions_for_ministries`**
  (8 códigos del `CHECK` de `admin_actions`), ambas con `up`/`down`; **sin migración de permisos**:
  el catálogo de F2 **sigue en 10**. Auditoría transaccional (registro atómico con la mutación,
  solo lectura, con denegaciones y rechazos); el alta registra **solo `ministry.create`**.
- **0 dependencias nuevas** y **`.env.example` sin cambios** (`APP_TZ` es de F4); cobertura
  `internal/ministerios/service*.go` **96,83 %** (366/378) en la revisión de código.

**Frontend (React + TypeScript + Vite)**
- **Sección pública `/ministerios`** con el sistema de F10 (`MinisteriosPage`, `MinistryCard`,
  `EncargadoLine`, `WhatsappJoinButton`): fichas con nombre, descripción, horario/lugar **como
  texto libre** (0 fechas derivadas), imagen con `alt`, encargado opcional y botón **«Unirse al
  grupo de WhatsApp»**; entrada «Ministerios» condicional en barra, menú móvil y pie
  (`useSiteLinks`, deducida del documento del bloque); **sin publicados redirige a la portada**
  con `replace`, sin destello y sin anuncio accesible (decisión del humano, fuera de alcance).
- **Bloque «Ministerios» en la portada** (`MinisteriosBlock` + `MinistryItem`, tras «Lo que viene»
  y antes de «Quiénes somos»): hasta 5 ítems por `sortOrder, id`, chip «Ministerio», botón
  «Ver más ministerios», **oculto por completo si no hay publicados** y sin encargado ni enlace.
- **Panel `/panel/ministerios`** (`MinistriesPage`, `MinistriesList`, `MinistryForm`,
  `ContactoEncargado`, `ShowContactSwitch`): alta/edición con pestañas es/en, **interruptor que
  nace apagado** («Oculto por defecto», D-13), reordenación, publicación/retirada, eliminación con
  confirmación y aviso de enlace compartido; entrada de menú, `AccesosRapidos` y rutas protegidas.
- **`CardImage` promovida a `frontend/src/components/publico/`** (inventario único; F4 la consume
  desde ahí, sin copia duplicada) e i18n `es`/`en` con claves tipadas nuevas y las **8 etiquetas de
  auditoría** `ministry.*`.
- **Vitest 652/652 (94 archivos)**, `lint`, `typecheck` y `build` en verde (649 antes de las
  correcciones del ciclo 1: +3 de `6328cff`).

**Infraestructura / Docker (devops)**
- **Sin cambios** en `docker-compose.yml`, Dockerfiles, `.github/`, `.env.example` ni en el kit
  compartido; sin variables de entorno nuevas.

### Decisiones firmes de la puerta 2 (aplicadas, no reabiertas)

- **El interruptor de la ficha del encargado nace «Oculto por defecto»** (`leader_visible BOOLEAN
  NOT NULL DEFAULT FALSE`): publicar nombre o contacto exige encenderlo de forma expresa; el
  **dato guardado** se conserva íntegro en el panel y el interruptor alterna sin reescribirlo.
- **`/ministerios` sin ningún ministerio publicado redirige a la portada**, de forma **silenciosa
  y sin anuncio accesible**; **distinta a propósito** de `/eventos` de F4 (aviso sin redirigir), que
  **no se toca** y cuyo `HiddenNotice` **no** se promueve al inventario único.

### Decisiones del cliente en las aclaraciones Q1–Q9 (2026-10-13)

Encargado **opcional** con nombre y **un** medio de contacto (celular **o** correo, nunca ambos;
contacto opcional aunque haya encargado) e **interruptor que oculta también el nombre** · enlace
de WhatsApp **opcional** · **cualquier `https` seguro** válido (sin restringir a dominios) · imagen
**opcional** cargada por el equipo · orden **manual** por columna numérica (`sort_order, id`) ·
**horario y lugar como textos libres** sin recurrencia calculada · **bloque de ministerios en la
portada** con **tope de 5**.

### Cambios acotados y documentados (previamente autorizados, no debilitamiento)

| ID | Qué es | Por qué |
|---|---|---|
| **RG6-6** (T626) | **4 pruebas de frontend** actualizadas al activarse el módulo: `lib/permissions.test.ts` (lista `MODULE_PERMISSIONS`), `features/roles/permissions.test.ts`, `features/roles/roles.test.tsx` (el sello «Disponible más adelante» deja de aplicarse a `ministerios`) y `features/panel/panel.test.tsx` (tarjeta nueva en accesos rápidos) | Mismo gesto autorizado que F4 hizo con los suyos (RG4-10): solo recuentos, listas y sellos. **El backend no toca ninguna prueba de F2** y el catálogo sigue en 10 |
| **RG6-5** (T610) | `IsFilePublished` ampliado **aditivamente** a `ministries` | Regresión de imágenes de F3/F4 verificada: `portada-imagenes` y `agenda-*` en verde; `media` de F3/F4 intacta |

### Cómo probarlo (quickstart)

Recorrido completo con evidencia por sección en [`quickstart.md`](./quickstart.md) (§0–§16).
Resumen:

```bash
make up                 # db, redis, backend, frontend
make db-migrate         # 000010 y 000011
cd frontend && npm install
# Portada (visitante): http://localhost:5173/  -> bloque «Ministerios»
curl -s 'http://localhost:8080/api/v1/portada/ministerios?lang=es' | jq
# Sección completa:    http://localhost:5173/ministerios
curl -s 'http://localhost:8080/api/v1/ministerios?lang=es' | jq
# Panel: login con la cuenta sembrada ana@ejemplo.com / Semilla.2026 (helpers de e2e)
# -> /panel/ministerios con el permiso «ministerios»
```

Pruebas automatizadas: `make ci` (lint + backend unit/integración + frontend + security) y las
e2e **en cuatro tandas** con `--workers=1` (comandos exactos en la cabecera de
`frontend/e2e/ministerios-publica.spec.ts`).

### Evidencia de pruebas

**Verificación de cierre (T637, 2026-10-13)**

| Verificación | Resultado |
|---|---|
| `make ci` | **en verde**: lint (gofmt/vet/golangci-lint), backend unitarias + integración contra PostgreSQL real, frontend, `govulncheck` **0** vulnerabilidades que afecten al código y `npm audit --audit-level=high` **0** |
| Cobertura `internal/ministerios/service*.go` | **96,83 %** (366/378): `service.go` 97,4 % · `service_admin.go` 95,7 % · `service_public.go` 97,6 % · `service_audit.go` 100 % (≥ 80 % exigido) |
| Frontend | **Vitest 652/652** (94 archivos), `lint`, `typecheck` y `build` en verde |
| e2e Playwright | **F6 19/19** (`ministerios-publica`, `ministerios-panel`, `ministerios-accesibilidad`, `portada-ministerios`) en la tanda 4, dentro de la suite completa **repartida en 4 tandas** por el límite de login; **tandas de regresión** de F2/F3/F10/F4 ejecutadas en serie (`--workers=1`) |
| Migraciones | `000010`/`000011` **up → down → up** (incluidas en las pruebas de integración) y catálogo de permisos **intacto en 10** |
| Deriva de generados | `make sqlc-verify` y `make api-gen` **sin diff**; contrato vivo **0.7.0** y `schema.d.ts` byte a byte igual al regenerado |
| Entorno | `.env.example` **sin cambios** · sin dependencias nuevas (`go.mod`/`package.json` intactos) · sin cambios de F6 en compose, Dockerfiles, `.github/` ni kit |

*Nota de método (T638)*: esta sesión del `documentador` **no tuvo acceso a shell ni a git**, así que
las cifras de esta tabla provienen de la ficha **T637 marcada como completa en `tasks.md`** y de los
informes de validación citados más abajo (la cobertura 96,83 % y `api-gen` sin deriva los verificó
`revisor-codigo`; las 652 unitarias y las 19 e2e de F6, `qa-tester` en dos corridas; las migraciones
up→down→up, ambas revisiones). **No se re-ejecutaron en esta sesión.**

**Validación del equipo (2026-10-13) — un ciclo de corrección**

| Informe | Veredicto | Nota |
|---|---|---|
| [`revision-2026-10-13-qa.md`](./revision-2026-10-13-qa.md) | **APROBADO** | 28 verificaciones en vivo con datos reales sembrados (privacidad 8/8 greps = 0, permisos 9×403, orden, tope 5, redirección silenciosa, `https` rechazados/aceptados); e2e F6 **19/19**; integración `ministerios` en verde; 0 defectos |
| [`revision-2026-10-13-codigo.md`](./revision-2026-10-13-codigo.md) | **APROBADO CON CAMBIOS MENORES** | sin CRÍTICOS; **1 IMPORTANTE (I1)** + 6 menores/sugerencias; cobertura **96,83 %**, `golangci-lint` **0 issues**, contrato 0.7.0 coherente con el delta, sin `skip`/`only` ni `any` injustificado |
| [`revision-2026-10-13-seguridad.md`](./revision-2026-10-13-seguridad.md) | **APROBADO** | sin críticos/altos/medios; privacidad del encargado **cerrada en todas las vías** con evidencia en vivo; `govulncheck` 0 y `npm audit` 0; SQL parametrizado y sin secretos |
| [`revision-2026-10-13.md`](./revision-2026-10-13.md) (consolidado) | **APROBADO CON CAMBIOS MENORES** | 1 bloqueante (**I1**) que abre el ciclo 1; el resto, deuda aceptada |

**Ciclo de corrección 1 (2026-10-13) — APROBADO por los tres roles, sin bloqueantes**:

- **I1** (contrato que desmentía su prosa en `whatsappUrl` y `minProperties` ausente) →
  `1e0d9e5` (contrato anulable + `minProperties: 1`, `schema.d.ts` regenerado) y `6328cff`
  (el formulario manda `null`, nunca `""`).
- **M1/I4** (`sortOrder` sin cota → truncamiento silencioso) → `2d75367` (cota `int32` con
  `minimum`/`maximum` y pruebas en las cuatro capas).
- Re-validación: `qa-tester` **APROBADO** (en vivo contra el stack **reconstruido**: `PATCH {}` →
  400, quitar el enlace conservando el resto, privacidad **12/12 greps = 0**, integración
  **125 PASS / 0 FAIL**, e2e F6 **19/19**, Vitest **652/652**), `revisor-codigo` **APROBADO CON
  CAMBIOS MENORES** (I1 y M1 **cerrados**, `api:gen` byte a byte igual) y `seguridad` **APROBADO
  (ratificado)** (I4 cerrado; ningún riesgo nuevo).

### Excepción aprobada (cambio de producto, no debilitamiento)

**Ninguna: F6 no necesitó saltarse ninguna regla del proceso.** A diferencia de F4 (que tuvo que
actualizar las pruebas de F2 que fijaban el catálogo en 9, con la excepción RG4-10/T405), aquí **el
catálogo de permisos no crece**: F6 **activa** el permiso `ministerios` que F2 ya sembró, de modo
que **no se tocó ninguna prueba de backend de F2**. Los dos cambios sobre pruebas ajenas que sí hubo
son **RG6-6** (las 4 pruebas de frontend que fijan la lista de permisos y accesos) y **RG6-5**
(`IsFilePublished` ampliada aditivamente), ambos **previstos y autorizados en `plan.md`** como
cambios acotados de producto, no como relajación de aserciones (verificados por los tres roles:
«solo ampliaron recuentos/listas; ninguna aserción se relajó ni se borró»).

### Deuda aceptada (registrada; no bloquea el PR)

Del [consolidado](./revision-2026-10-13.md): **M2** (`RecordAction` vive en `service.go`,
precedente literal de F4) · **M3** (`mergeInvalid` conserva solo el último motivo de error; el
contrato devuelve el detalle por campo) · **M4** (`setTimeout(20)` en una prueba de portada:
espera frágil; conviene fijarla antes de fusionar a `main` si molesta) · **M5** (el listado del panel
se corta en 100 sin avisar el total; el catálogo real es de 5–10 ministerios) · **M6** (`Optional[T]`
llega a su tercer dominio: candidato a promoverse a `platform/`) · **S1/S2** (redacción de prosa y
3 fragmentos de snapshot superados por las correcciones: registrarlos, no editarlos) · **B1**
(seguridad: `imageFile` aceptado por patrón sin comprobar su existencia en el almacén; efecto real
una imagen rota, no una fuga; igual que el B2 de F4) · **B2** (contraseñas de prueba en los e2e,
cuentas efímeras locales) · **I1–I3** de seguridad (informativos: `GO-2026-5932` preexistente y no
invocado, superficie de escritura de auditoría por 403 heredada de F2–F4, `target_label` con el
nombre del ministerio y nunca datos del encargado) · **deuda de proceso**: no hay anuncio accesible
al redirigir `/ministerios` (decisión del humano, fuera de alcance de la spec).

### Pendiente manual (no automatizable; no bloquea abrir el PR)

- **SC-012** — prueba de usabilidad con **≥ 6 personas** y el **cronómetro de los 30 s de SC-001**;
  protocolo listo en [`pruebas-usabilidad-SC-012.md`](./pruebas-usabilidad-SC-012.md). Responsable:
  el equipo/humano.
- **SC-008** — lectura con **lector de pantalla real** (NVDA/VoiceOver) y pantalla táctil física de
  `/ministerios` y del bloque; lo automatizable (teclado, foco, jerarquía, contraste, 200 %, 320 px,
  ambos temas) ya está en verde.
- **Flaky preexistente (no de F6)**: `frontend/e2e/qa-bugs-004.spec.ts` «Bug 1» (scroll a
  `#whatsapp`, ~25 %) es de F3/F4; F6 no cambió ese spec.

### Notas de entrega

- **Tareas**: 37 de 38 completas (T601–T637) y **T638 marcada en este commit** (documentación y
  este borrador de PR). *Cubre el cierre de US1–US6 / FR-001…FR-029 / SC-001…SC-017 como
  entregable*, con los dos pendientes manuales arriba.
- **Documentación actualizada por `documentador` en este commit**: `CHANGELOG.md` (**0.6.0**, con
  «Agregado» de la sección/panel/bloque/privacidad, «Cambiado» del permiso, la navegación, la
  portada y el contrato 0.7.0, «Corregido» con las **2** correcciones del ciclo 1 y «No entra en
  esta versión»), `README.md` (sección «Qué es F6», `/ministerios` y el bloque en el quickstart,
  migraciones `000010`/`000011`, consultas `ministries.sql`, tandas e2e y enlaces),
  `docs/entrega/F6-ministerios.md` (nota para el cliente, en lenguaje no técnico) y este `pr.md`.
  **No se tocó** `spec.md`, `plan.md`, `ux.md`, `estado.md`, el roadmap ni `docs/plantillas/`.
- **Contrato**: `backend/api/openapi.yaml` en **0.7.0** y sin deriva de `make api-gen` /
  `make sqlc-verify`.
- **Aprobaciones humanas registradas**: spec v1 y **v2** («Re-apruebo la spec v2») y plan («Apruebo
  el plan»), todos el 2026-10-13; **no queda nada pendiente de re-aprobación**.
- **Merge y despliegue**: solo con aprobación humana explícita (puertas pendientes en `estado.md`);
  después, `make costos CERRAR=1`.
