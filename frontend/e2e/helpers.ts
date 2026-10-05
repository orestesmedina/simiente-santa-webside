import type { APIRequestContext, APIResponse, Page } from '@playwright/test';

/**
 * Utilidades compartidas por los e2e de F2 (`acceso.spec.ts` y
 * `auditoria.spec.ts`). Este archivo no es un `*.spec.ts`, así que Playwright no
 * lo trata como prueba.
 *
 * El stack completo corre en local (`make up`): la SPA se sirve en
 * `http://localhost:5173` (baseURL del config) y la API en
 * `http://localhost:8080` (VITE_API_URL horneada en el build del frontend).
 */

/** API del backend publicada por docker-compose (VITE_API_URL). */
export const API_URL = process.env.E2E_API_URL ?? 'http://localhost:8080';

/**
 * Token de la inicialización única. Debe coincidir con `BOOTSTRAP_TOKEN` del
 * `.env` local (valor de desarrollo documentado: `dev-bootstrap-token`). Se
 * puede sobreescribir con `E2E_SETUP_TOKEN`.
 */
export const SETUP_TOKEN =
  process.env.E2E_SETUP_TOKEN ?? process.env.BOOTSTRAP_TOKEN ?? 'dev-bootstrap-token';

/** Administrador inicial creado por `POST /setup/initialize` (quickstart §1). */
export const ADMIN = {
  firstName: 'Ana',
  lastName: 'Responsable',
  email: 'ana@ejemplo.com',
  phone: '+34 612 345 678',
  password: 'Semilla.2026',
} as const;

/** Sufijo único por ejecución: hace repetibles los nombres y correos creados. */
export function uniqueSuffix(): string {
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 7)}`;
}

function isHttpStatus(response: APIResponse, expected: number): boolean {
  return response.status() === expected;
}

/**
 * Inicializa el administrador una sola vez (FR-007/quickstart §1). Es
 * idempotente: `201` la primera vez y `409` en las siguientes (la
 * inicialización ya se hizo), de modo que el e2e es repetible sin limpiar la
 * base de datos. Cualquier otro estado se reporta con un mensaje accionable.
 */
export async function ensureInitialized(request: APIRequestContext): Promise<void> {
  const response = await request.post(`${API_URL}/api/v1/setup/initialize`, {
    headers: { 'Content-Type': 'application/json', 'X-Setup-Token': SETUP_TOKEN },
    data: ADMIN,
  });

  if (isHttpStatus(response, 201) || isHttpStatus(response, 409)) {
    return;
  }
  if (response.status() === 401 || response.status() === 403) {
    throw new Error(
      `La inicialización rechazó el X-Setup-Token (${response.status()}). ` +
        `Revisa que BOOTSTRAP_TOKEN del .env local valga "${SETUP_TOKEN}" y reconstruye/levanta el stack.`,
    );
  }
  if (response.status() === 429) {
    throw new Error(
      'POST /setup/initialize respondió 429 (rate-limit por IP): espera ~1 minuto y reintenta.',
    );
  }
  throw new Error(`La inicialización falló con ${response.status()}: ${await response.text()}`);
}

/** Inicia sesión rellenando la pantalla de acceso (ux.md §3.1). */
export async function loginViaUi(page: Page, email: string, password: string): Promise<void> {
  await page.goto('/login');
  await page.getByLabel('Correo').fill(email);
  // El campo obligatorio añade " *" al texto de la etiqueta; se ancla el nombre.
  await page.getByLabel(/^Contraseña\s*\*?$/).fill(password);
  await page.getByRole('button', { name: 'Entrar' }).click();
}

/**
 * Lee el valor de la cookie CSRF *double-submit* de la sesión abierta en el
 * contexto del navegador. La cookie no es `HttpOnly` (el cliente la lee desde
 * `document.cookie`), así que Playwright la puede recuperar.
 */
async function csrfToken(page: Page): Promise<string> {
  const cookies = await page.context().cookies(API_URL);
  const csrf = cookies.find((cookie) => cookie.name === 'csrf_token');
  if (!csrf) {
    throw new Error('No hay cookie csrf_token: ¿se ha iniciado sesión en este contexto?');
  }
  return csrf.value;
}

/**
 * Envía una petición autenticada a la API desde el contexto del navegador
 * (comparte sus cookies): los métodos no seguros llevan la cabecera
 * `X-CSRF-Token`, igual que `src/api/client.ts` (P10). Se usa para generar
 * actividad de auditoría sin recorrer toda la interfaz.
 */
export async function apiSend(
  page: Page,
  method: 'GET' | 'POST' | 'PATCH' | 'DELETE',
  path: string,
  data?: unknown,
): Promise<APIResponse> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (method !== 'GET') {
    headers['X-CSRF-Token'] = await csrfToken(page);
  }
  return page.request.fetch(`${API_URL}${path}`, {
    method,
    headers,
    ...(data === undefined ? {} : { data }),
  });
}
