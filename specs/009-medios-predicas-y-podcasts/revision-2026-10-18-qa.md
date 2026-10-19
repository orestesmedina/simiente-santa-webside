# Revisión QA — F9 Medios: prédicas y podcasts (2026-10-18/19)

**Rol**: `qa-tester` · **Rama**: `009-medios-predicas-y-podcasts` · **Alcance**: verificación independiente contra la **spec v2 aprobada** (6 US, 30 FR, 17 SC), no fiada del informe de cierre T944.

## Veredicto: APRUEBA

Todo lo automatizable pasa en mi propia corrida, con **verificaciones en vivo contra la API real**: la búsqueda contra los feeds reales funcionó a demanda con su resumen correcto y **idempotente** (2ª pasada: 0 importados, 0 cambios), el fail-safe de FR-016 se observó en vivo (YouTube caído desde la IP actual → importó lo del podcast, informó el fallo y **ocultó/marcó 0**), el público expone solo lo publicable (ocultar uno en la base: 9→8 y **0 filtraciones**), el bloque de la portada devuelve **exactamente 5**, los permisos dan `401`/`403` con mensaje claro y el intento sin permiso queda **auditado como `denied`** (comprobado en vivo contra `admin_actions`), y el catálogo de permisos sigue en **10** con «medios» presente. No añadí pruebas nuevas: la cobertura automatizada existente ya cubre cada criterio automatizable de la spec, la revisé de código (las aserciones son las de la spec, sin debilitamientos) y la ejecuté con `--workers=1` (nota H1).

## Suites ejecutadas por mí

| Suite | Comando | Resultado |
|---|---|---|
| Backend unitarias | `cd backend && go test ./...` | ✅ ok (paquete `medios` incluido) |
| Backend integración (PostgreSQL real) | `go test -tags=integration -count=1 ./internal/medios/...` | ✅ ok (0,4 s) |
| Frontend unitarias | `npx vitest run` | ✅ **930/930** en 129 archivos |
| E2e de F9 | `npx playwright test --config e2e/playwright.config.ts --workers=1 medios- portada-medios` (4 tandas) | ✅ **25/25**: `medios-panel` 9, `medios-publica` 8, `portada-medios` 4, `medios-accesibilidad` 4 |
| Lint / seguridad | `make lint` · `make security` | ✅ (0 vulnerabilidades npm audit) |

## Verificación independiente en vivo (API + PostgreSQL, no repetida del informe)

| # | Verificación | Evidencia literal | Resultado |
|---|---|---|---|
| V-1 | Búsqueda real contra los feeds (POST `/api/v1/admin/medios/buscar-nuevos`, primera vez) | `{"importedCount":9,"joinedCount":0,"markedCount":0,"hiddenCount":0,…,"youtube":"failed","podcast":"ok"}` en ~1,8 s | ✅ FR-011/FR-018: a demanda, sin cron/worker, resumen completo |
| V-2 | **Fail-safe FEED CAÍDO en vivo (FR-016/SC-007)**: el feed de YouTube responde 404/500 desde esta IP (reproducido con `curl`: `HTTP/2 404`; research R9-1 lo anticipa) | Resumen arriba: importa lo del otro feed, reporta `youtube:"failed"`, y `hiddenCount:0`/`markedCount:0` — **nada se oculta ni se marca por un feed caído** | ✅ |
| V-3 | **Idempotencia real (FR-012/SC-004)**: 2ª búsqueda | `{"importedCount":0,…}`; BD: `9 \| 9` (9 filas, 9 visibles), `media_sync_runs` = 2 | ✅ |
| V-4 | Sección pública `GET /api/v1/medios`: fichas con `coverUrl`, `title`, `description` corta, `publishedAt`, `audioUrl` — **claves que no aplican se omiten** (sin `hidden` ni `reviewMark`) | `total 9`, p. ej. `La Fe y la confianza en el Abba \| 2026-07-03 \| desc 225 \| keys ['audioUrl','coverUrl','description','id','publishedAt','title']`; descripciones recortadas a ≤ ~300 con «…» | ✅ FR-002/FR-020 |
| V-5 | **SC-002 en vivo**: oculto por SQL un ítem visible → el público deja de verlo | públicos 9→**8**, `fuga: 0` (ninguna respuesta contiene el id oculto) | ✅ |
| V-6 | **Tope 5 del bloque en el servidor (FR-027/SC-015)** | `GET /api/v1/portada/medios?lang=es` → **`items 5`** con 9 visibles; los 5 primeros en el orden de la sección | ✅ |
| V-7 | Tras restaurar el oculto, el bloque vuelve a los 5 | `portada items tras restaurar: 5` | ✅ |
| V-8 | **401 sin sesión** | `POST buscar-nuevos` sin cookies → **401** `unauthenticated` | ✅ FR-021 |
| V-9 | **403 con cuenta sin «medios»** | creé rol `qa-sin-f9-oct18` (`noticias`) + cuenta `qa-sin-medios-oct18@…` con contraseña cambiada; `POST buscar-nuevos` → **403** `{"error":{"code":"forbidden","message":"No tienes permiso para acceder a este módulo"}}`; `GET /admin/medios` → **403** | ✅ FR-021/SC-009 |
| V-10 | **El 403 queda auditado (FR-022/SC-010)** | `admin_actions`: `media.sync\|denied│3` (incluye mi intento en vivo, junto con los de las e2e), y éxitos/denegaciones por cada operación (`media.hide`, `media.show`, `media.merge`, `media.description.update`, `media.mark.resolve` con `success` y `denied`) | ✅ |
| V-11 | **Catálogo de permisos intacto (SC-009)** | `SELECT count(*) FROM permissions` → **10**; `code='medios'` existe | ✅ |

## Cobertura automatizada por criterio (revisada de código y ejecutada)

| Criterio | Prueba(s) que lo cubren | Estado |
|---|---|---|
| FR-001..FR-009, SC-001/SC-002/SC-013 (público: fichas, orden, claves omitidas, 0 ocultos/marcados/absorbidos/par abierto) | `medios-publica.spec.ts` (SC-001/002/013, nav FR-007/007, Cargar más FR-006) + `service_public_test.go` + repositorio (integración) | ✅ pasa |
| FR-004/SC-005 (0 terceros, 0 iframes, carátulas del sitio) | `medios-publica.spec.ts:263` — intercepta **todas** las peticiones de `/medios` **y `/`**, asert `toEqual([])` sobre dominios youtube/ytimg/spotify/scdn/anchor/cloudfront; **0 atributos** `src/data/poster` de terceros; `iframe` = 0 | ✅ (aserciones fuertes, no debilitadas) |
| FR-005/SC-006 (enlaces anunciados y pestaña nueva anunciada) | `medios-publica.spec.ts:323` (ruta simulada de los servicios, comprueba pestaña nueva y URL sin red) | ✅ |
| FR-010..FR-012, SC-004 (a demanda, idempotencia) | `service_sync_test.go` (`TestBuscarNuevosImportaElSetReal`, `…EsIdempotente`) + **V-3 en vivo** | ✅ |
| FR-013/FR-014 (unión segura / dudoso separado oculto y marcado) | `TestBuscarNuevosUnionAutomatica`, `…DuplicadoDudoso`, `…SinCandidatura` + e2e unión con doble confirmación | ✅ |
| FR-015/SC-007 (no pisa lo del equipo) | `TestBuscarNuevosNoPisaLoDelEquipo` + e2e ocultar/mostrar | ✅ |
| FR-016 (feed caído) | `TestBuscarNuevosFeedCaido` (uno y **los dos** caídos: nada cambia) + **V-2 en vivo** | ✅ |
| FR-017/SC-014 (ventana del feed y disponibilidad) | `TestBuscarNuevosVentanaDelFeed` (muerto oculta y marca; vivo sigue visible y marca; **inconcluso nunca oculta**), `TestBuscarNuevosAgregacionPorFicha` (**el peor resultado gana: un enlace muerto oculta la ficha completa aun con dos enlaces**), `TestProbeSoloConfirma404Y410`, «misma clase ya resuelta + enlace vivo → **no se reabre**» + e2e `marca-unavailable`, `marca-resuelta`, `vivo-fuera`, `inconcluso` sembrados | ✅ |
| «Marca de disponibilidad no rompe el par de duplicados» | e2e `medios-panel:507` (par mixto: los dos ejes se resuelven por separado) + `TestBuscarNuevosPrecedenciaA3` | ✅ |
| FR-018 (resumen) | `TestBuscarNuevosResumenYRegistro` + V-1 en vivo | ✅ |
| FR-019/FR-020 (panel reducido, descripción, **vacía → se omite la línea**, decisión 28) | `medios-panel.spec.ts` (listado, editar es/en, **en sin español no guarda** FR-023) + fixture `sin-descripcion` verificada en `medios-publica` | ✅ |
| FR-021/SC-009 (permisos serv. + 403 auditado) | `medios-panel.spec.ts:565` (tres cuentas: con `medios`, otros módulos, ninguna) + V-8..V-11 en vivo | ✅ |
| FR-022/SC-010 (auditoría) | e2e `medios-panel:661` + `service_audit_test.go` (`result='denied'`) + rama de auditoría en integración | ✅ |
| FR-023 (validación, sin parciales) | e2e editar descripción (límite y es obligatorio) | ✅ |
| FR-024..FR-026/SC-011 (bilingüe con respaldo es, datos idénticos entre idiomas) | `medios-publica.spec.ts:380` (inglés: fallback es→siempre, datos sin alterar) | ✅ |
| FR-027/SC-015/SC-002 (bloque portada) | `portada-medios.spec.ts`: más de 5 → **exactamente 5** en su posición; menos → exactos; **0 publicables → el bloque no existe**; + V-6/V-7 en vivo | ✅ |
| FR-008/SC-016 (nada publicable: nav oculta y `/medios` redirige sin destello) | `medios-publica.spec.ts:488` + `medios-accesibilidad` (P9-16) | ✅ |
| FR-028..FR-030/SC-008/SC-017 (responsiva 320 px, teclado, foco visible, 200 %, reduced motion) | `medios-publica.spec.ts:437`, `medios-accesibilidad.spec.ts` (4 pruebas, claro/oscuro), `portada-medios:266` | ✅ (automatizable) |
| SC-008 real (lector de pantalla) y SC-012 (personas) | **Manual, tras despliegue.** Protocolo `pruebas-usabilidad-SC-012.md` **existe (78 líneas)** y cubre lo que la spec exige: muestra ≥6 personas en 3 franjas con poca experiencia, tarea 2 mide *antes de tocar el enlace* que sale del sitio y va en pestaña nueva, cronómetro 30 s de SC-001, checklist SC-008 (teclado, anuncio de enlaces, contraste claro/oscuro, 200 %, reduced motion), criterios ≥90 % y 0 hallazgos críticos, y firma humana. Estado registrado como **pendiente, no bloquea PR pero sí el cierre** — igual que F6/F7 | ⚠️ confirmado protocolo |

## Hallazgos

| # | Severidad | Hallazgo | Evidencia |
|---|---|---|---|
| H1 | **medio** (solo infraestructura de pruebas, no producto) | Los **4 specs e2e de F9 fallan si se corren en paralelo** (la config `fullyParallel` de `e2e/playwright.config.ts` sin `workers=1`): la primera corrida dio `4 failed / 21 did not run` con error en la siembra (`psql` vía `docker compose exec` concurrente — `mediosSeed.ts:217`). Restricción **preexistente y declarada** en `mediosSeed.ts` (base compartida): las 4 tandas del cierre (T944) ya se corrían de a uno. Reproduje 25/25 solo con `--workers=1`. | `test-results/portada-medios-…SC-015…/error-context.md`; la misma prueba pasa aislada y toda la tanda en verde con `--workers=1` |
| H2 | **bajo** (entorno) | Desde esta IP el **feed de YouTube** responde 404/500 (no alcanzable desde datacenter), así que en mi sesión no pude comprobar en vivo la unión automática (FR-013) ni las comprobaciones de disponibilidad reales; T944 ya lo hizo cuando YouTube respondía (importó 10 y unió 8). El **fail-safe** sí quedó observado en vivo (V-2). | `curl` al feed → `HTTP/2 404`; resumen V-1 con `"youtube":"failed"` |
| H3 | bajo (residuo de la sesión de QA) | Mi verificación en vivo dejó en la BD de desarrollo 9 episodios reales importados del podcast y las cuentas/rol `qa-sin-f9-oct18`/`qa-sin-medios-oct18@…` (fines de prueba, datos reales del cliente de origen). Ninguna altera el comportamiento; avisar para decidir si se conservan o se limpian siembras. | V-1..V-4 |

**No encontré** ningún defecto de producto: 0 filtraciones de ocultos, 0 duplicados visibles, 0 peticiones a terceros, redirección silenciosa, tope 5 respetado, permisos íntegros y auditados.

## Qué no pude verificar y por qué

1. **SC-008 (lector de pantalla real) y SC-012 (usabilidad con personas)**: post-despliegue por diseño de la spec; solo confirmé la existencia y el contenido del protocolo (ver tabla, última fila).
2. **Unión automática sobre los feeds reales con YouTube accesible (FR-013)** y **comprobaciones de disponibilidad reales (FR-017)**: el feed de YouTube no responde desde esta red ahora mismo (H2). Quedan cubiertos por los dobles en Go (revisados línea a línea: las aserciones corresponden exactamente a los tres casos de FR-017 y a la agregación por ficha de la fase 5) y por la experiencia T944.
3. **El evento «enlace público deja de estar disponible» en producción real** (p. ej. Spotify retirando un episodio): solo simulable con dobles; el flujo está cubierto en Go y por fixtures sembradas (`marca-unavailable`).

## Pruebas nuevas añadidas

**Ninguna.** Tras revisar la cobertura contra los 30 FR y 17 SC, cada criterio automatizable tiene al menos una prueba que lo cubre con aserciones equivalentes a la spec; los dos criterios manuales tienen su protocolo. Lo novedoso lo aporta la batería en vivo V-1..V-11 (documentada arriba, reproducible con `curl`/psql contra el stack levantado).

---
*Preparado por `qa-tester`, 2026-10-18/19, F9 fase 7 (validación). Sin commit: informe entregado tal cual al orquestador.*
