import { expect, test, type Page, type Response } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { API_URL, ADMIN, apiSend, ensureInitialized, loginViaUi, uniqueSuffix } from './helpers';

/**
 * E2E de regresión por el bug «las imágenes de la portada no cargan».
 *
 * Causa raíz diagnosticada: la API devuelve URLs de media relativas
 * (`/api/v1/media/img_….jpg`) y la SPA las usa como `src` en su propio origen
 * (`http://localhost:5173`), pero el nginx del frontend no proxya `/api/`:
 * `GET http://localhost:5173/api/v1/media/…` responde el fallback de SPA
 * (`200 text/html`, `index.html`) en vez de la imagen. Directo al backend
 * (`:8080`) sí responde `200 image/jpeg`.
 *
 * La prueba lo reproduce a través del navegador real: siembra una identidad
 * publicada con logo e imagen de portada, abre la portada pública en el origen
 * de la SPA y comprueba (a) que la respuesta de red de esas URLs es `image/*`
 * (no `text/html` de `index.html`) y (b) que las imágenes realmente descodifican
 * (`naturalWidth > 0`).
 */

/** Imagen real del repositorio, usada para las subidas (patrón portada-publica). */
const FIXTURE_IMAGE = join(process.cwd(), '..', 'resources', 'simiente.jpeg');

interface PublicPortada {
  identity: {
    logoUrl?: string;
    coverImageUrl?: string;
  };
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

/** Registro mínimo de la respuesta de red que pidió el navegador. */
interface MediaResponse {
  url: string;
  status: number;
  contentType: string | null;
}

/** Sube un archivo de imagen y devuelve el `fileName` que asigna la API. */
async function subirImagen(page: Page, name: string): Promise<string> {
  const upload = await page.request.post(`${API_URL}/api/v1/admin/portada/imagenes`, {
    headers: { 'X-CSRF-Token': await csrfToken(page) },
    multipart: {
      file: { name, mimeType: 'image/jpeg', buffer: readFileSync(FIXTURE_IMAGE) },
    },
  });
  expect(upload.status(), `la subida de ${name}`).toBe(201);
  return ((await upload.json()) as { fileName: string }).fileName;
}

test.describe('Portada: imágenes del hero cargan desde el origen de la SPA', () => {
  test.setTimeout(180_000);

  test('el logo y la imagen de portada publicados responden image/* y descodifican', async ({
    browser,
    request,
  }) => {
    await ensureInitialized(request);

    const suffix = uniqueSuffix();
    const identityName = `Iglesia Simiente ${suffix}`;
    const logoAlt = `Logotipo de ${identityName}`;
    const coverAlt = `Imagen de portada de ${identityName}`;

    const adminContext = await browser.newContext();
    try {
      // ── Sembrar: subir logo + portada y publicar la identidad con ambos ──
      const adminPage = await adminContext.newPage();
      await loginViaUi(adminPage, ADMIN.email, ADMIN.password);
      await expect(adminPage).toHaveURL(/\/panel$/);

      const logoFileName = await subirImagen(adminPage, 'logo.jpg');
      const coverFileName = await subirImagen(adminPage, 'portada.jpg');

      expect(
        (
          await apiSend(adminPage, 'PUT', '/api/v1/admin/portada/identidad', {
            nameEs: identityName,
            missionEs: `Misión ${suffix}`,
            logoFile: logoFileName,
            logoAltEs: logoAlt,
            coverImageFile: coverFileName,
            coverImageAltEs: coverAlt,
            publicationState: 'published',
          })
        ).ok(),
      ).toBe(true);

      // Los URL relativos que la API pública expone al visitante.
      const portaPublica = (
        (await (await request.get(`${API_URL}/api/v1/portada`)).json()) as PublicPortada
      ).identity;
      const logoUrl = portaPublica.logoUrl ?? '';
      const coverUrl = portaPublica.coverImageUrl ?? '';
      expect(logoUrl, 'la identidad publicada tiene logoUrl').toMatch(/^\/api\/v1\/media\//);
      expect(coverUrl, 'la identidad publicada tiene coverImageUrl').toMatch(/^\/api\/v1\/media\//);

      // Saneamiento: el backend sirve la imagen de verdad (control negativo).
      // Si esto falla, el problema es del siembra, no del proxy del frontend.
      // (El logo sí puede 404 mientras la identidad esté en borrador: ya se publicó.)
      const fromBackend = await request.get(`${API_URL}${logoUrl}`);
      expect(fromBackend.status(), 'la media se sirve directo del backend').toBe(200);
      expect(fromBackend.headers()['content-type']).toMatch(/^image\//);

      // ── El navegador (origen de la SPA) pide las imágenes del hero ────────
      const requestsPorUrl = new Map<string, MediaResponse>();
      const registrar = (response: Response): void => {
        const url = response.url();
        if (url === `${browserBase(logoUrl)}` || url === `${browserBase(coverUrl)}`) {
          requestsPorUrl.set(url, {
            url,
            status: response.status(),
            contentType: response.headers()['content-type'] ?? null,
          });
        }
      };

      const visitorContext = await browser.newContext();
      try {
        const page = await visitorContext.newPage();
        page.on('response', registrar);

        await page.goto('/');
        // El `h1` del nombre siembrado confirma que la portada renderizó.
        await expect(page.getByRole('heading', { level: 1, name: identityName })).toBeVisible();

        // (a) La red del origen de la SPA devolvió la imagen, no `index.html`.
        await expect
          .poll(() => requestsPorUrl.get(browserBase(logoUrl)), {
            message: `el navegador pidió ${browserBase(logoUrl)}`,
          })
          .toBeDefined();
        await expect
          .poll(() => requestsPorUrl.get(browserBase(coverUrl)), {
            message: `el navegador pidió ${browserBase(coverUrl)}`,
          })
          .toBeDefined();
        const logoRespuesta = requestsPorUrl.get(browserBase(logoUrl))!;
        expect(logoRespuesta.status, 'el logo carga por HTTP desde la SPA').toBe(200);
        expect(logoRespuesta.contentType, 'el logo es una imagen (no index.html)').toMatch(
          /^image\//,
        );
        const coverRespuesta = requestsPorUrl.get(browserBase(coverUrl))!;
        expect(coverRespuesta.status, 'la portada carga por HTTP desde la SPA').toBe(200);
        expect(coverRespuesta.contentType, 'la portada es una imagen (no index.html)').toMatch(
          /^image\//,
        );

        // (b) Realmente descodifican como imagen dentro del `<img>` del hero.
        await expect(page.getByRole('img', { name: logoAlt })).toBeAttached();
        await expect
          .poll(async () =>
            page
              .getByRole('img', { name: logoAlt })
              .evaluate((el) => (el as HTMLImageElement).naturalWidth),
          )
          .toBeGreaterThan(0);
        await expect(page.getByRole('img', { name: coverAlt })).toBeAttached();
        await expect
          .poll(async () =>
            page
              .getByRole('img', { name: coverAlt })
              .evaluate((el) => (el as HTMLImageElement).naturalWidth),
          )
          .toBeGreaterThan(0);
      } finally {
        await visitorContext.close();
      }
    } finally {
      await adminContext.close();
    }
  });
});

/**
 * Resuelve la URL absoluta que pedirá el navegador para un `src` relativo
 * (`/api/…`) contra el origen de la SPA (`baseURL` del config: `:5173`).
 * Es justo el punto del bug: el mismo recurso servido en el origen de la SPA.
 */
function browserBase(path: string): string {
  return new URL(path, 'http://localhost:5173').toString();
}
