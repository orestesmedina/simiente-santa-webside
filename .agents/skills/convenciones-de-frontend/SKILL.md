---
name: convenciones-de-frontend
description: Convenciones propias de la parte frontend (nombres, estructura, patrones, bibliotecas). Usar siempre que se cree o modifique código en frontend/.
---
# Convenciones de frontend

Redactadas el 2026-10-20. Estándar de referencia: skill `react-frontend`. Rutas relativas a `frontend/`.

## Nombres
- El código nuevo se nombra en español (decisión del 2026-10-20): `useBuscarEpisodios` (`src/features/medios/hooks/useBuscarEpisodios.ts:16`), `quitarSeparadores` (`src/features/donaciones/validation.ts:153`). Lo que ya está en inglés no se renombra sin una tarea propia: `useCreateUser` (`src/features/usuarios/hooks/useCreateUser.ts:11`).
- Componente en PascalCase y página con sufijo `Page`, uno por archivo: `export function DonacionForm` (`src/features/donaciones/components/DonacionForm.tsx:118`).
- Hook con `use`, un archivo por hook con el nombre del hook: `export function useBuscarEpisodios()` (`src/features/medios/hooks/useBuscarEpisodios.ts:16`).
- Textos de interfaz en `messages.ts` de la feature, constantes SCREAMING_SNAKE: `export const LOAD_ERROR = 'No se pudo cargar los episodios.'` (`src/features/medios/messages.ts:143`).
- Claves de caché `*_QUERY_KEY = [...] as const`: `DONACIONES_ADMIN_QUERY_KEY` (`src/features/donaciones/hooks/queryKeys.ts:10`).
- Tipos del API re-exportados del esquema generado: `components['schemas']['DonacionAdmin']` (`src/api/ofrendas.ts:34`).
- Comentarios en español que citan la spec, decisión o tarea que los origina: `(D-6)`, `(FR-012)` (`src/app/router.tsx:155`).

## Estructura
- Cada feature vive en `src/features/<modulo>/` con `pages/`, `components/`, `hooks/` y, en la raíz, `validation.ts`, `errors.ts` y `messages.ts` (`src/features/donaciones/`).
- El acceso al servidor va en `src/api/<modulo>.ts`, uno por módulo del backend, sobre `src/api/client.ts`: `export const DONACIONES_BASE = '/api/v1/admin/donaciones'` (`src/api/ofrendas.ts:44`).
- El armazón en `src/app/`; lo compartido en `src/components/` y el sistema del sitio público en `src/components/publico/`, con su catálogo en `README.md` (`src/components/publico/README.md:212`).
- Utilidades puras sin React en `src/lib/`: `export function hasPermission(...)` (`src/lib/permissions.ts:74`).
- Idioma, tema y movimiento del sitio público en `src/features/publico/{i18n,theme,motion}/`: `const { t } = useLanguage()` (`src/features/publico/components/IdentityHero.tsx:39`).
- Pruebas: componentes y páginas junto al archivo (`src/features/donaciones/components/DonacionForm.test.tsx`); las de hooks, agregadas en `hooks/*Hooks.test.tsx` (`src/features/donaciones/hooks/donacionesHooks.test.tsx`).
- e2e de Playwright en `e2e/`: specs `*-publica`, `*-panel`, `*-accesibilidad` y `qa-*`; datos en `*Seed.ts` y utilidades en `helpers.ts` (`e2e/portada-panel.spec.ts:3`).

## Patrones
- Todo dato del servidor por TanStack Query: `useQuery` con `queryFn` sobre `src/api/`, y cada mutación invalida sus claves al triunfar (`src/features/donaciones/hooks/useCreateDonacion.ts:17`).
- Formularios con estado controlado y Zod a mano (decisión del 2026-10-20): `const parsed = donacionSchema(kind).safeParse(fields)` (`src/features/donaciones/components/DonacionForm.tsx:169`). react-hook-form queda solo en F2, como legado.
- La validación de cliente es espejo de la del servidor, que es la autoridad (`src/features/donaciones/validation.ts:21`).
- Errores del API traducidos en `errors.ts` de la feature con `isForbidden`, `detailMessage` y `generalErrorMessage` (`src/features/donaciones/errors.ts:34`).
- Toda vista con datos cubre carga, vacío y error con `EmptyState`, y el resultado de una acción con `Notice`: `<Notice variant="error">{NO_PERMISSION_ERROR}</Notice>` (`src/features/medios/pages/MediosPage.tsx:177`).
- Los campos bilingües se editan con pestañas Español / English que solo muestran lo traducible: `const [langTab, setLangTab] = useState<'es' | 'en'>('es')` (`src/features/ministerios/components/MinistryForm.tsx:128`).
- El acceso se decide en las rutas con `RequirePermission code={...}` y los permisos de `src/lib/permissions.ts`; ocultar no protege, el servidor manda (`src/app/router.tsx:169`).
- Tras guardar o publicar en el panel se invalidan también las consultas públicas del módulo (`useInvalidateDonaciones`, `src/features/donaciones/hooks/useDonaciones.ts:54`).

## Bibliotecas y herramientas
- React 19 con TypeScript `strict` y Vite; Tailwind CSS 4 con los tokens de marca en `@theme` y modo oscuro por `data-theme` (`src/index.css:57`).
- react-router-dom 7 con el árbol de rutas como `RouteObject[]` y `useRoutes` (`src/app/router.tsx:81`).
- TanStack Query 5 sin reintento, sin refresco al enfocar y `staleTime: 0` (`src/app/providers.tsx:10`).
- zod 4 para validar; el contrato tipado se genera con `openapi-typescript` (`npm run api:gen`).
- `apiFetch` propio con `ApiError`, timeout de 5 s, cookie de sesión y doble envío CSRF (`src/api/client.ts:151`).
- Vitest (jsdom) + Testing Library + MSW para pruebas, y Playwright contra el stack real para e2e (`src/test/setup.ts:7`).

## Dónde se aparta del estándar
| El estándar dice | Aquí se hace | En código nuevo |
|---|---|---|
| Para llegar a un elemento, `useRef` | El foco tras validar va por `document.getElementById(...).focus()` (`src/features/donaciones/components/DonacionForm.tsx:136`) | Se sigue el proyecto |
| Cada prueba junto a lo que prueba | Los hooks agrupan sus pruebas en `hooks/*Hooks.test.tsx` (`src/features/donaciones/hooks/donacionesHooks.test.tsx`) | Se sigue el proyecto |

## Deuda
- Seguridad: URLs que vienen de datos se pintan en `href` sin comprobar el esquema (`src/features/publico/components/SocialSection.tsx:33`, `src/features/medios/components/FuentesList.tsx:33`, `src/features/publico/medios/MedioLinks.tsx:38`, `src/features/publico/ministerios/WhatsappJoinButton.tsx:47`); las de F9 vienen de feeds de terceros. Falta un `safeUrl()` en el render.
- Seguridad: credenciales de desarrollo en el código de e2e (`e2e/helpers.ts:21-31`); deben venir solo de variables de entorno.
- `npm audit` no corre en la integración continua del frontend: el verbo existe en el perfil, pero el workflow no lo ejecuta.
