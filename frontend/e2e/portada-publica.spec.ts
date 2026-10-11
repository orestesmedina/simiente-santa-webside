import { expect, test, type Page } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { ADMIN, API_URL, apiSend, ensureInitialized, loginViaUi, uniqueSuffix } from './helpers';

/**
 * E2E de la portada pública (T337, US1 · US3 lado público · US4 · US5) sobre
 * `quickstart.md` §1/§5/§7/§10. Cubre SC-001 (información sin cuenta), SC-002
 * (ningún borrador, ni sus imágenes — C4), SC-006 (inglés completo con
 * fallback `en → es` sin huecos), SC-007/SC-008 (tres anchos sin scroll
 * horizontal y recorrido por teclado), SC-009 (enlaces con un clic) y SC-012
 * (secciones sin publicados ocultas).
 *
 * Se ejecuta contra el stack real (`make up` + `make db-migrate`). El contenido
 * se siembra por API con una cuenta con permiso (`ADMIN` de F2), y el visitante
 * se recorre en contextos **sin sesión**. La identidad de la iglesia es un
 * singleton compartido: **ejecutar con `--workers=1`** (`make e2e`) para que
 * esta spec no compita con `portada-panel.spec.ts`.
 */

/** Imagen real del repositorio, usada para la subida del logotipo (R3-8). */
const FIXTURE_IMAGE = join(process.cwd(), '..', 'resources', 'simiente.jpeg');

const WIDTHS = [320, 360, 768, 1280] as const;

interface AdminPortada {
  whatsapp: { items: { id: string; publicationState: string }[] };
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

test.describe('Portada pública (visitante, idioma, accesibilidad)', () => {
  test.setTimeout(180_000);

  test('secciones publicadas, idioma es/en con fallback, sin borradores y responsiva', async ({
    browser,
    request,
  }) => {
    await ensureInitialized(request);

    const suffix = uniqueSuffix();
    const identityName = `Iglesia Simiente ${suffix}`;
    const mission = `Misión ${suffix}`;
    const aboutText = `Somos una iglesia cercana y abierta ${suffix}.`;
    const serviceName = `Culto dominical ${suffix}`;
    const directChannel = `Escríbenos ${suffix}`;
    const draftService = `Borrador interno ${suffix}`;
    const directPhone = '+506 7000 1234';
    // El backend normaliza el teléfono a dígitos (contract `Contact`, FR-015);
    // el enlace `tel:` y su texto usan ese valor normalizado.
    const directPhoneNormalized = directPhone.replace(/[^\d+]/g, '');
    const contactEmail = `hola-${suffix}@ejemplo.com`;
    const instagramUrl = `https://www.instagram.com/simiente-${suffix}`;

    const adminContext = await browser.newContext();
    const adminPage = await adminContext.newPage();
    let logoFileName: string;

    try {
      // ── Sembrar contenido con una cuenta con permiso (API) ────────────────
      await loginViaUi(adminPage, ADMIN.email, ADMIN.password);
      await expect(adminPage).toHaveURL(/\/panel$/);

      const upload = await adminPage.request.post(`${API_URL}/api/v1/admin/portada/imagenes`, {
        headers: { 'X-CSRF-Token': await csrfToken(adminPage) },
        multipart: {
          file: { name: 'logo.jpg', mimeType: 'image/jpeg', buffer: readFileSync(FIXTURE_IMAGE) },
        },
      });
      expect(upload.status()).toBe(201);
      logoFileName = ((await upload.json()) as { fileName: string }).fileName;

      // Identidad publicada **sin** `nameEn`: en inglés el servidor resuelve al
      // español (FR-009/SC-006), nunca un hueco.
      expect(
        (
          await apiSend(adminPage, 'PUT', '/api/v1/admin/portada/identidad', {
            nameEs: identityName,
            missionEs: mission,
            logoFile: logoFileName,
            logoAltEs: `Logotipo de ${identityName}`,
            publicationState: 'published',
          })
        ).ok(),
      ).toBe(true);

      expect(
        (
          await apiSend(adminPage, 'PUT', '/api/v1/admin/portada/quienes-somos', {
            textEs: aboutText,
            textEn: null,
            publicationState: 'published',
          })
        ).ok(),
      ).toBe(true);

      expect(
        (
          await apiSend(adminPage, 'PUT', '/api/v1/admin/portada/contacto', {
            addressEs: `San José, ${suffix}`,
            email: contactEmail,
            phone: directPhone,
            publicationState: 'published',
          })
        ).ok(),
      ).toBe(true);

      expect(
        (
          await apiSend(adminPage, 'POST', '/api/v1/admin/portada/horario', {
            dayOfWeek: 0,
            startTime: '10:00',
            endTime: '12:00',
            nameEs: serviceName,
            placeEs: 'Templo principal',
            publicationState: 'published',
          })
        ).ok(),
      ).toBe(true);

      expect(
        (
          await apiSend(adminPage, 'POST', '/api/v1/admin/portada/whatsapp', {
            nameEs: directChannel,
            kind: 'direct',
            destination: directPhone,
            publicationState: 'published',
          })
        ).ok(),
      ).toBe(true);

      // Un elemento en borrador que jamás debe verse (SC-002).
      expect(
        (
          await apiSend(adminPage, 'POST', '/api/v1/admin/portada/horario', {
            dayOfWeek: 1,
            startTime: '09:00',
            nameEs: draftService,
            placeEs: 'Anexo',
            publicationState: 'draft',
          })
        ).ok(),
      ).toBe(true);

      // Redes: una sola por red (Q5); si una corrida previa ya dejó Instagram,
      // se actualiza en vez de crear (409).
      const adminState = (await (
        await apiSend(adminPage, 'GET', '/api/v1/admin/portada')
      ).json()) as { socials: { items: { id: string; network: string }[] } };
      const instagram = adminState.socials.items.find((link) => link.network === 'instagram');
      const socialResponse = instagram
        ? await apiSend(adminPage, 'PATCH', `/api/v1/admin/portada/redes/${instagram.id}`, {
            url: instagramUrl,
            publicationState: 'published',
          })
        : await apiSend(adminPage, 'POST', '/api/v1/admin/portada/redes', {
            network: 'instagram',
            url: instagramUrl,
            publicationState: 'published',
          });
      expect(socialResponse.ok()).toBe(true);

      // ── SC-001: el visitante sin sesión ve las secciones publicadas ───────
      const visitorContext = await browser.newContext();
      const page = await visitorContext.newPage();
      try {
        await page.goto('/');

        await expect(page.getByRole('heading', { level: 1, name: identityName })).toBeVisible();
        await expect(page.getByRole('heading', { name: 'Nuestra misión' })).toBeVisible();

        const sectionsNav = page.getByRole('navigation', { name: 'Secciones de la portada' });
        await expect(sectionsNav.getByRole('link', { name: 'Quiénes somos' })).toBeVisible();
        await expect(sectionsNav.getByRole('link', { name: 'Horario de servicios' })).toBeVisible();
        await expect(sectionsNav.getByRole('link', { name: 'WhatsApp' })).toBeVisible();
        await expect(sectionsNav.getByRole('link', { name: 'Contacto' })).toBeVisible();
        await expect(sectionsNav.getByRole('link', { name: 'Redes sociales' })).toBeVisible();

        await expect(page.locator('#quienes-somos').getByText(aboutText)).toBeVisible();
        // El horario se siembra con un sufijo único, pero la base conserva los
        // servicios de corridas previas (la identidad es un singleton
        // compartido): se apunta a la fila del servicio creado para que el
        // locator no resuelva a varias filas con la misma hora.
        const scheduleTable = page.getByRole('table', { name: 'Horario de servicios' });
        const serviceRow = scheduleTable.getByRole('row').filter({ hasText: serviceName });
        await expect(serviceRow.getByText(serviceName)).toBeVisible();
        await expect(serviceRow.getByText('10:00 a. m. − 12:00 m.')).toBeVisible();

        // SC-009: los enlaces de WhatsApp y redes llevan al destino correcto.
        const whatsappLink = page
          .locator('#whatsapp li')
          .filter({ hasText: directChannel })
          .getByRole('link');
        await expect(whatsappLink).toHaveAttribute('href', /^https:\/\/wa\.me\/\d+$/);
        await expect(whatsappLink).toHaveAttribute('target', '_blank');
        await expect(whatsappLink).toHaveAttribute('rel', /noopener/);

        const socialLink = page
          .locator('#redes li')
          .filter({ hasText: 'Ver en Instagram' })
          .getByRole('link');
        await expect(socialLink).toHaveAttribute('href', instagramUrl);
        await expect(socialLink).toHaveAttribute('target', '_blank');

        await expect(
          page.locator('#contacto').getByRole('link', { name: contactEmail }),
        ).toHaveAttribute('href', `mailto:${contactEmail}`);
        await expect(
          page.locator('#contacto').getByRole('link', { name: directPhoneNormalized }),
        ).toHaveAttribute('href', `tel:${directPhoneNormalized}`);

        // SC-002: el borrador no aparece por ninguna vía.
        await expect(page.getByText(draftService)).toHaveCount(0);
      } finally {
        await visitorContext.close();
      }

      // ── FR-010/C3: primera visita en español, cambio, persistencia ────────
      const cleanContext = await browser.newContext();
      const langPage = await cleanContext.newPage();
      try {
        await langPage.goto('/');
        await expect(langPage.locator('html')).toHaveAttribute('lang', 'es');
        await expect(
          langPage.locator('header').getByRole('radio', { name: 'Español' }),
        ).toBeChecked();
        await expect(langPage.getByRole('heading', { name: 'Quiénes somos' })).toBeVisible();

        // Cambio a inglés: interfaz + contenidos (el servidor resuelve el
        // fallback; `about`/identidad no tienen inglés → siguen en español).
        await langPage.locator('header').getByRole('radio', { name: 'Inglés' }).check();
        await expect(langPage.locator('html')).toHaveAttribute('lang', 'en');
        await expect(
          langPage.getByRole('navigation', { name: 'Page sections' }).getByRole('link', {
            name: 'Who we are',
          }),
        ).toBeVisible();
        await expect(langPage.getByRole('heading', { level: 1, name: identityName })).toBeVisible();
        await expect(langPage.getByRole('heading', { name: 'Our mission' })).toBeVisible();
        await expect(langPage.locator('#quienes-somos').getByText(aboutText)).toBeVisible();
        // Nunca un hueco: el h1 y el texto conservan contenido.
        await expect(langPage.getByRole('heading', { level: 1 })).not.toHaveText('');

        // Persiste al navegar (recarga) y en `localStorage`.
        await langPage.reload();
        await expect(langPage.locator('html')).toHaveAttribute('lang', 'en');
        expect(await langPage.evaluate(() => window.localStorage.getItem('ss.lang'))).toBe('en');

        // Re-visita con la preferencia guardada: respeta el idioma (C3).
        const storage = await cleanContext.storageState();
        const revisitContext = await browser.newContext({ storageState: storage });
        const revisitPage = await revisitContext.newPage();
        try {
          await revisitPage.goto('/');
          await expect(revisitPage.locator('html')).toHaveAttribute('lang', 'en');
          await expect(
            revisitPage.getByRole('navigation', { name: 'Page sections' }).getByRole('link', {
              name: 'Who we are',
            }),
          ).toBeVisible();
        } finally {
          await revisitContext.close();
        }
      } finally {
        await cleanContext.close();
      }

      // Sin auto-detección del navegador: contexto en inglés, sin preferencia
      // guardada → español (FR-010).
      const autoContext = await browser.newContext({ locale: 'en-US' });
      const autoPage = await autoContext.newPage();
      try {
        await autoPage.goto('/');
        await expect(autoPage.locator('html')).toHaveAttribute('lang', 'es');
      } finally {
        await autoContext.close();
      }

      // ── SC-012: una sección sin publicados desaparece ─────────────────────
      const state = (await (
        await apiSend(adminPage, 'GET', '/api/v1/admin/portada')
      ).json()) as AdminPortada;
      const publishedChannels = state.whatsapp.items.filter(
        (channel) => channel.publicationState === 'published',
      );
      expect(publishedChannels.length).toBeGreaterThan(0);
      for (const channel of publishedChannels) {
        expect(
          (
            await apiSend(adminPage, 'PATCH', `/api/v1/admin/portada/whatsapp/${channel.id}`, {
              publicationState: 'draft',
            })
          ).ok(),
        ).toBe(true);
      }

      const hiddenContext = await browser.newContext();
      const hiddenPage = await hiddenContext.newPage();
      try {
        await hiddenPage.goto('/');
        await expect(hiddenPage.locator('#whatsapp')).toHaveCount(0);
        await expect(
          hiddenPage
            .getByRole('navigation', { name: 'Secciones de la portada' })
            .getByRole('link', { name: 'WhatsApp' }),
        ).toHaveCount(0);
      } finally {
        await hiddenContext.close();
      }

      // Restaurar el estado para el resto de la prueba (y para el panel).
      for (const channel of publishedChannels) {
        expect(
          (
            await apiSend(adminPage, 'PATCH', `/api/v1/admin/portada/whatsapp/${channel.id}`, {
              publicationState: 'published',
            })
          ).ok(),
        ).toBe(true);
      }

      // ── C4: la imagen de una identidad retirada deja de servirse ──────────
      expect((await request.get(`${API_URL}/api/v1/media/${logoFileName}`)).status()).toBe(200);
      expect(
        (
          await apiSend(adminPage, 'PUT', '/api/v1/admin/portada/identidad', {
            nameEs: identityName,
            logoFile: logoFileName,
            logoAltEs: `Logotipo de ${identityName}`,
            publicationState: 'draft',
          })
        ).ok(),
      ).toBe(true);
      expect((await request.get(`${API_URL}/api/v1/media/${logoFileName}`)).status()).toBe(404);
      expect(
        (
          await apiSend(adminPage, 'PUT', '/api/v1/admin/portada/identidad', {
            nameEs: identityName,
            logoFile: logoFileName,
            logoAltEs: `Logotipo de ${identityName}`,
            publicationState: 'published',
          })
        ).ok(),
      ).toBe(true);
      expect((await request.get(`${API_URL}/api/v1/media/${logoFileName}`)).status()).toBe(200);

      // ── SC-007/SC-008: responsiva sin scroll horizontal y teclado ─────────
      const responsiveContext = await browser.newContext();
      const responsivePage = await responsiveContext.newPage();
      try {
        for (const width of WIDTHS) {
          await responsivePage.setViewportSize({ width, height: 900 });
          await responsivePage.goto('/');
          await expect(
            responsivePage.getByRole('heading', { level: 1, name: identityName }),
          ).toBeVisible();
          const overflow = await responsivePage.evaluate(
            () => document.documentElement.scrollWidth - window.innerWidth,
          );
          expect(overflow, `sin scroll horizontal a ${width}px`).toBeLessThanOrEqual(1);
        }

        // Recorrido por teclado: el primer `Tab` enfoca el enlace de salto.
        await responsivePage.setViewportSize({ width: 1280, height: 900 });
        await responsivePage.goto('/');
        await responsivePage.keyboard.press('Tab');
        await expect(
          responsivePage.getByRole('link', { name: 'Saltar al contenido principal' }),
        ).toBeFocused();
        await responsivePage.keyboard.press('Enter');
        await expect(responsivePage).toHaveURL(/#contenido$/);
      } finally {
        await responsiveContext.close();
      }
    } finally {
      await adminContext.close();
    }
  });
});
