# Estado: 2026-10-17-ci-migraciones-serializadas

<!-- Bug: el CI falla de forma intermitente con deadlock (40P01) en las pruebas de migración. -->

## Resumen

| Campo | Valor |
|---|---|
| Rama | fix/ci-migraciones-serializadas |
| Flujo | equipo-bug |
| Fase | Cerrado · **PR #12 fusionado en `main`** (`5e9a44a`, 2026-10-17) |
| Ciclo de corrección | 0/3 |
| Próximo paso | Nada. El **defecto conocido 23** del roadmap queda **resuelto** por este arreglo |
| Bloqueado por | — |
| Actualizado | 2026-10-17 |

## Aprobaciones

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | n/a (bug) | | | |
| Plan | n/a (bug) | | | |
| PR / merge | aprobado y fusionado | humano (cliente) | 2026-10-17 | PR #12 fusionado en `main` por el humano |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

- (ninguno)

## Diagnóstico (causa raíz)

- En CI las pruebas de integración corren **en paralelo contra una sola base** (`DATABASE_URL_TEST`). Varias pruebas `TestIntegrationMigration*UpDownUp` ejecutan un `down` que hace `DELETE FROM admin_actions` (RowExclusive) seguido de `ALTER TABLE admin_actions` (AccessExclusive) **en la misma transacción**; dos de esos ciclos solapados se bloquean en ciclo → **deadlock (40P01)**. Modo de fallo secundario: `23514` (el `down` de `000006` restaura el CHECK de F2 y otra prueba pudo confirmar filas con códigos de F3–F7).
- Es **preexistente** (F3/F4/F6 ya tenían tres ciclos; F7 añadió el cuarto). Falló dos veces en el CI (PR #10 y PR #11), así que deja de ser aceptable como flaky. Registrado como **defecto conocido 23** del roadmap.
- **Descartados por producir efectos peores:** (a) `ACCESS EXCLUSIVE` + borrado dentro de la transacción (bloquea `admin_actions` todo el ciclo → timeouts en cascada en otras pruebas); (b) base de datos aislada por prueba con `CREATE/DROP DATABASE` (satura el servidor).

## Arreglo

- **Serializar** los ciclos de migración destructivos con un **lock de aviso transaccional** de PostgreSQL (`pg_advisory_xact_lock`) adquirido como **primera sentencia** de la transacción del ciclo (`testutil.SerializeMigrations`), de modo que nunca se solapen. Se aplica a todas las pruebas que ejecutan `down` destructivos: portada `000006`, agenda `000008`/`000009`, ministerios `000010`/`000011`, donaciones `000012`/`000013` (y los `PermissionsCatalogUnchanged`).
- Incluye una **prueba de regresión determinista** en `testutil` (el lock queda en `pg_locks`; una segunda transacción con `lock_timeout` corto espera; al revertir, no queda ningún lock).
- No toca el CI (`.github/workflows/ci.yml` es archivo del kit), ni migraciones aplicadas, ni código de producción.
- **Recuperado** de los commits `d864e10` y `41388ef` (validados y luego revertidos durante F7 al probar los otros dos enfoques).

## Decisiones y aclaraciones

- 2026-10-17 (humano): tras el **segundo** fallo del flaky en el CI, elige **relanzar el job del PR #11 y preparar este arreglo definitivo** (en vez de seguir dejándolo como defecto conocido).

## Verificación

- **Estrés local con la base compartida** (`DATABASE_URL_TEST`): **10/10** corridas de los paquetes que compiten (`agenda`, `donaciones`, `ministerios`, `portada`, `testutil`) sin ningún `40P01` ni `23514`.
- **Suite de integración completa**: en verde (un único fallo transitorio de `session`/`TestIntegrationStoreIdleTTLRefresh` bajo carga, que pasa 3/3 en aislamiento y en paralelo; no se toca en este cambio).
- `make lint` limpio; `go vet`/`gofmt` limpios.
- **Revisión de código previa** (sobre los mismos commits, durante F7): **APRUEBA**, con la prueba de regresión verificada por mutación (helper no-op y clave distinta fallan).

## Bitácora

- 2026-10-17 — **Cerrado**: PR #12 fusionado en `main` (`5e9a44a`) por el humano.
- 2026-10-17 — Arreglo recuperado y verificado (10/10 estrés, lint limpio); registro creado. Pendiente el PR.
