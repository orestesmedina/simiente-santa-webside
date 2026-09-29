---
name: react-frontend
description: Convenciones de la empresa para escribir frontend en React + TypeScript + Vite (estructura, estado, llamadas a la API, estilos, pruebas). Usar siempre que se cree o modifique código en frontend/.
---
# Frontend en React — convenciones

## Herramientas
- React 19 + TypeScript (`strict: true`) + Vite.
- Rutas: React Router. Datos del servidor: TanStack Query. Formularios: React Hook Form + Zod.
- Estilos: Tailwind CSS. Componentes base accesibles: shadcn/ui (o los que defina el plan).
- Pruebas: Vitest + Testing Library; end-to-end con Playwright.
- Lint y formato: ESLint + Prettier.

## Estructura
```
frontend/src/
├── app/            # router, providers, layout principal
├── api/            # cliente HTTP y tipos generados del openapi.yaml
├── components/     # componentes reutilizables (UI genérica)
├── features/
│   └── <feature>/  # pantallas, hooks y componentes propios de la funcionalidad
│       ├── pages/
│       ├── components/
│       ├── hooks/  # useUsers(), useCreateUser()... sobre TanStack Query
│       └── *.test.tsx
└── lib/            # utilidades
frontend/e2e/       # pruebas Playwright
```

## Reglas
- Componentes funcionales, pequeños y con props tipadas. Un componente por archivo.
- **Nunca** llames a `fetch` desde un componente: usa un hook de `features/<feature>/hooks/` que use el cliente de `api/`.
- Tipos de la API generados desde `backend/api/openapi.yaml` con `openapi-typescript`. Prohibido `any`.
- Valida formularios con esquemas Zod; muestra errores junto al campo.
- Cada vista con datos implementa los estados: cargando, vacío, error y éxito.
- La URL base de la API viene de `import.meta.env.VITE_API_URL`.
- Accesibilidad: elementos semánticos (`button`, `label`, `nav`), `alt` en imágenes, foco visible, navegable por teclado.
- Nada de `dangerouslySetInnerHTML`. Nada de tokens en `localStorage`; la sesión va en cookie `HttpOnly` gestionada por el backend.

## Pruebas
- Prueba comportamiento visible para el usuario (`getByRole`, `userEvent`), no detalles de implementación.
- Simula la API con MSW (Mock Service Worker) en pruebas de componentes.
- Comandos: `npm run lint`, `npm run typecheck`, `npm test -- --run`, `npx playwright test`.
