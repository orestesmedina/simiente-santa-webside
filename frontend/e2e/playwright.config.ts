import { defineConfig, devices } from '@playwright/test';

/**
 * Configuración de Playwright para ejecución **local** (D17): el `ci.yml` del
 * kit no ejecuta e2e y no se edita.
 *
 * Requiere el stack levantado (`make up`): el `baseURL` apunta al frontend
 * publicado en `http://localhost:5173` por `docker-compose.yml` (WEB_PORT por
 * defecto). No arranca ningún servidor web: las pruebas corren contra el stack
 * real (nginx → API) para validar el flujo completo de F1.
 */
export default defineConfig({
  testDir: '.',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: 0,
  reporter: [['list']],
  use: {
    baseURL: 'http://localhost:5173',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
