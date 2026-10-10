import { expect, test, type Page, type Response } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { API_URL, ADMIN, apiSend, ensureInitialized, loginViaUi, uniqueSuffix } from './helpers';

/**
 * E2E de regresión por el bug «las imágenes de la portada no cargan».
 *
 * Causa raíz diagnosticada: la API devuelve URLs de media relativas
 * (`/api/v1/media/img_….jpg`) y la SPA las usaba como `src` **sin resolverlas**
 * contra la base de la API, de modo que el navegador las pedía a su propio
 * origen (`http://localhost:5173`), donde el nginx del frontend no proxya
 * `/api/`: `GET http://localhost:5173/api/v1/media/…` devolvía el fallback de
 * SPA (`200 text/html`, `index.html`) en vez de la imagen. Directo al backend
 * (`:8080`) sí responde `200 image/jpeg`.
 *
 * El arreglo (`mediaUrl` en `src/api/client.ts`) prefija las URLs relativas con
 * `API_BASE_URL`, así que el navegador las pide **al origen de la API**
 * (`http://localhost:8080`). La prueba lo verifica a través del navegador real:
 * siembra una identidad publicada con logo e imagen de portada, abre la portada
 * pública y comprueba (a) que la respuesta de red de esas URLs es `image/*`
 * (no `text/html` de `index.html`) y (b) que las imágenes realmente descodifican
 * (`naturalWidth > 0`), acotando los *locators* por zona (el logo aparece en
 * cabecera, hero y pie con el mismo `alt`), y (c) que los `src` quedan
 * **resueltos contra el origen de la API** (`API_BASE_URL`, I-1): cabecera y
 * pie llaman a `mediaUrl` de forma independiente (`PublicLayout` y
 * `PublicFooter`), así que una regresión parcial en cualquiera de las tres
 * llamadas (hero, cabecera, pie) hace fallar esta prueba.
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

test.describe('Portada: las imágenes de hero, cabecera y pie cargan (URLs de media resueltas contra la API)', () => {
  test.setTimeout(180_000);

  test('el logo y la imagen de portada publicados responden image/* y descodifican en todas las zonas', async ({
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

      // ── El navegador (origen de la API, vía mediaUrl) pide las imágenes ───
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

        // (a) La red devolvió la imagen (desde la API), no `index.html`.
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
        expect(logoRespuesta.status, 'el logo carga por HTTP desde la API').toBe(200);
        expect(logoRespuesta.contentType, 'el logo es una imagen (no index.html)').toMatch(
          /^image\//,
        );
        const coverRespuesta = requestsPorUrl.get(browserBase(coverUrl))!;
        expect(coverRespuesta.status, 'la portada carga por HTTP desde la API').toBe(200);
        expect(coverRespuesta.contentType, 'la portada es una imagen (no index.html)').toMatch(
          /^image\//,
        );

        // (b) Realmente descodifican como imagen dentro del `<img>` del hero.
        // El `alt` del logo se repite en cabecera, hero y pie (mismo contenido
        // publicado), así que se acota a la región del hero (nombre sembrado).
        const hero = page.getByRole('region', { name: identityName });
        await expect(hero.getByRole('img', { name: logoAlt })).toBeAttached();
        await expect
          .poll(async () =>
            hero
              .getByRole('img', { name: logoAlt })
              .evaluate((el) => (el as HTMLImageElement).naturalWidth),
          )
          .toBeGreaterThan(0);
        await expect(hero.getByRole('img', { name: coverAlt })).toBeAttached();
        await expect
          .poll(async () =>
            hero
              .getByRole('img', { name: coverAlt })
              .evaluate((el) => (el as HTMLImageElement).naturalWidth),
          )
          .toBeGreaterThan(0);

        // (c) Cabecera y pie también pintan el logo con llamadas independientes
        // a `mediaUrl` (PublicLayout y PublicFooter): una regresión parcial que
        // solo quite `mediaUrl` en una de ellas dejaría el hero intacto. Se
        // verifica cada zona por separado y acotada (banner/contentinfo), pues
        // el mismo `alt` del logo aparece tres veces en la página.
        const logoSrcResuelto = browserBase(logoUrl);
        for (const [zona, zonaNombre] of [
          [page.getByRole('banner'), 'la cabecera'],
          [page.getByRole('contentinfo'), 'el pie'],
        ] as const) {
          const logoZona = zona.getByRole('img', { name: logoAlt });
          await expect(logoZona, `el logo de ${zonaNombre}`).toBeAttached();
          // (c1) El `src` quedó resuelto contra el origen de la API por
          // `mediaUrl`, no relativo a la SPA: el punto exacto de la regresión.
          await expect(logoZona, `src del logo de ${zonaNombre}`).toHaveAttribute(
            'src',
            logoSrcResuelto,
          );
          // (c2) Y descodifica: el src relativo pediría `index.html` al nginx
          // de la SPA y `naturalWidth` sería 0 (el bug original).
          await expect
            .poll(async () => logoZona.evaluate((el) => (el as HTMLImageElement).naturalWidth))
            .toBeGreaterThan(0);
        }

        // (d) Ancla también los `src` resueltos del hero (logo y portada).
        await expect(hero.getByRole('img', { name: logoAlt })).toHaveAttribute(
          'src',
          logoSrcResuelto,
        );
        await expect(hero.getByRole('img', { name: coverAlt })).toHaveAttribute(
          'src',
          browserBase(coverUrl),
        );
      } finally {
        await visitorContext.close();
      }
    } finally {
      await adminContext.close();
    }
  });
});

/**
 * URL absoluta que pedirá el navegador para una URL de media relativa
 * (`/api/…`): la resuelve contra el **origen de la API** (`API_URL`), que es
 * donde el arreglo (`mediaUrl`) apunta el `src`. Antes del arreglo la SPA la
 * resolvía contra su propio origen (`:5173`), el punto exacto del bug.
 */
function browserBase(path: string): string {
  return new URL(path, API_URL).toString();
}
