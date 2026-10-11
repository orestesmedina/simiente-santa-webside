import { existsSync } from 'node:fs';
import { readFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { expect, type Browser, type BrowserContext, type Page } from '@playwright/test';
import { ADMIN, API_URL, apiSend, ensureInitialized, uniqueSuffix } from './helpers';

/**
 * Siembra de la portada pública para los e2e de F10 (T1029–T1033).
 *
 * No es un `*.spec.ts` (Playwright no lo trata como prueba) y **no lo usan** las
 * suites congeladas de F2/F3: existe para no repetir el *setup* en cada spec
 * nueva. Publica una identidad con logo y portada, «quiénes somos», un servicio
 * de horario, un canal de WhatsApp, contacto y una red, con un sufijo único por
 * ejecución para que las aserciones no choquen con corridas previas.
 */

/** Imagen real del repositorio, usada para las subidas (patrón portada-publica). */
const FIXTURE_IMAGE = join(process.cwd(), '..', 'resources', 'simiente.jpeg');

export interface PublicSeed {
  identityName: string;
  mission: string;
  aboutText: string;
  serviceName: string;
  channelName: string;
  contactEmail: string;
  contactPhone: string;
  /** Teléfono normalizado (dígitos) que usa el `tel:` y su texto visible. */
  contactPhoneNormalized: string;
  logoAlt: string;
  coverAlt: string;
  instagramUrl: string;
  logoFileName: string;
}

/** Lee la cookie CSRF de la sesión abierta (misma que `src/api/client.ts`). */
async function csrfToken(page: Page): Promise<string> {
  const cookies = await page.context().cookies(API_URL);
  const csrf = cookies.find((cookie) => cookie.name === 'csrf_token');
  if (!csrf) {
    throw new Error('No hay cookie csrf_token: ¿se ha iniciado sesión en este contexto?');
  }
  return csrf.value;
}

/**
 * Contexto con sesión de administrador, **compartido por todas las specs de
 * F10**: la primera llamada inicia sesión (por API, sin preflight CORS) y guarda
 * el `storageState`; las siguientes lo reutilizan. Evita agotar el rate-limit de
 * `POST /api/v1/auth/login` (20/min por IP) que ya roza la suite de F2/F3 al
 * añadirse estas cinco specs.
 */
export async function createAdminContext(browser: Browser): Promise<BrowserContext> {
  const stateFile = join(tmpdir(), 'simiente-e2e-f10-admin.json');
  if (!existsSync(stateFile)) {
    const context = await browser.newContext();
    try {
      const credentials = { email: ADMIN.email, password: ADMIN.password };
      let login = await context.request.post(`${API_URL}/api/v1/auth/login`, { data: credentials });
      if (!login.ok()) {
        // Sistema sin inicializar: se crea el administrador y se reintenta.
        await ensureInitialized(context.request);
        login = await context.request.post(`${API_URL}/api/v1/auth/login`, { data: credentials });
      }
      expect(login.ok(), `login de administrador → ${login.status()}`).toBe(true);
      await context.storageState({ path: stateFile });
    } finally {
      await context.close();
    }
  }
  return browser.newContext({ storageState: stateFile });
}

/**
 * Publica la portada completa (todas las secciones) con la sesión del panel ya
 * abierta en `page`. Devuelve los textos sembrados para las aserciones.
 */
export async function seedPublicPortada(page: Page): Promise<PublicSeed> {
  const suffix = uniqueSuffix();
  const identityName = `Iglesia Simiente ${suffix}`;
  const mission = `Misión ${suffix}`;
  const aboutText = `Somos una iglesia cercana y abierta ${suffix}.`;
  const serviceName = `Culto dominical ${suffix}`;
  const channelName = `Escríbenos ${suffix}`;
  const contactEmail = `hola-${suffix}@ejemplo.com`;
  const contactPhone = '+506 7000 4321';
  const contactPhoneNormalized = contactPhone.replace(/[^\d+]/g, '');
  const instagramUrl = `https://www.instagram.com/simiente-${suffix}`;

  const upload = await page.request.post(`${API_URL}/api/v1/admin/portada/imagenes`, {
    headers: { 'X-CSRF-Token': await csrfToken(page) },
    multipart: {
      file: { name: 'logo.jpg', mimeType: 'image/jpeg', buffer: readFileSync(FIXTURE_IMAGE) },
    },
  });
  expect(upload.status(), 'la subida del logo').toBe(201);
  const logoFileName = ((await upload.json()) as { fileName: string }).fileName;

  // Las colecciones del backend tienen tope (horario ≤ 50, WhatsApp ≤ 20) y la
  // base es compartida: se retira lo viejo para que esta prueba sea repetible.
  const state = (await (await apiSend(page, 'GET', '/api/v1/admin/portada')).json()) as {
    schedule: { items: { id: string }[] };
    whatsapp: { items: { id: string }[] };
  };
  for (const item of state.schedule.items) {
    await apiSend(page, 'DELETE', `/api/v1/admin/portada/horario/${item.id}`);
  }
  for (const channel of state.whatsapp.items) {
    await apiSend(page, 'DELETE', `/api/v1/admin/portada/whatsapp/${channel.id}`);
  }

  for (const [method, path, body] of [
    [
      'PUT',
      '/api/v1/admin/portada/identidad',
      {
        nameEs: identityName,
        missionEs: mission,
        logoFile: logoFileName,
        logoAltEs: `Logotipo de ${identityName}`,
        coverImageFile: logoFileName,
        coverImageAltEs: `Imagen de portada de ${identityName}`,
        publicationState: 'published',
      },
    ],
    [
      'PUT',
      '/api/v1/admin/portada/quienes-somos',
      { textEs: aboutText, publicationState: 'published' },
    ],
    [
      'PUT',
      '/api/v1/admin/portada/contacto',
      {
        addressEs: `San José, ${suffix}`,
        email: contactEmail,
        phone: contactPhone,
        publicationState: 'published',
      },
    ],
    [
      'POST',
      '/api/v1/admin/portada/horario',
      {
        dayOfWeek: 0,
        startTime: '10:00',
        endTime: '12:00',
        nameEs: serviceName,
        placeEs: 'Templo principal',
        publicationState: 'published',
      },
    ],
    [
      'POST',
      '/api/v1/admin/portada/whatsapp',
      {
        nameEs: channelName,
        kind: 'direct',
        destination: contactPhone,
        publicationState: 'published',
      },
    ],
  ] as const) {
    const response = await apiSend(page, method, path, { ...body });
    expect(response.ok(), `${method} ${path} → ${response.status()}`).toBe(true);
  }

  // Una sola red por red (Q5): si una corrida previa ya dejó Instagram, se
  // actualiza en vez de crear (409).
  const adminState = (await (await apiSend(page, 'GET', '/api/v1/admin/portada')).json()) as {
    socials: { items: { id: string; network: string }[] };
  };
  const instagram = adminState.socials.items.find((link) => link.network === 'instagram');
  const socialResponse = instagram
    ? await apiSend(page, 'PATCH', `/api/v1/admin/portada/redes/${instagram.id}`, {
        url: instagramUrl,
        publicationState: 'published',
      })
    : await apiSend(page, 'POST', '/api/v1/admin/portada/redes', {
        network: 'instagram',
        url: instagramUrl,
        publicationState: 'published',
      });
  expect(socialResponse.ok(), 'la red publicada').toBe(true);

  return {
    identityName,
    mission,
    aboutText,
    serviceName,
    channelName,
    contactEmail,
    contactPhone,
    contactPhoneNormalized,
    logoAlt: `Logotipo de ${identityName}`,
    coverAlt: `Imagen de portada de ${identityName}`,
    instagramUrl,
    logoFileName,
  };
}
