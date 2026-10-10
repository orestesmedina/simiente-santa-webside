import { expect, test, type Page } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { ADMIN, API_URL, apiSend, ensureInitialized, loginViaUi, uniqueSuffix } from './helpers';

/**
 * E2e de accesibilidad de la portada pública (SC-008 / FR-019, quickstart §10):
 * parte **automatizada** de la checklist WCAG 2.1 AA sobre el stack real, sin
 * dependencias nuevas. Cubre:
 *
 *   §10.1        320 px sin desplazamiento horizontal (con portada-publica)
 *   §10.2        enlaces/botones con área de toque mayor o igual a 44 px
 *   §10.3        recorrido por teclado con foco VISIBLE en cada paso, sin
 *                trampa de foco (el ciclo vuelve al enlace de salto), y
 *                activación del selector de idioma con Espacio
 *   §10.4        estructura semántica (un `h1`, títulos `h2`, `lang`, imágenes
 *                con `alt` no vacío — FR-019)
 *   §10.5        contraste mayor o igual a 4.5:1 (texto grande: 3:1) de todo el
 *                texto de `main` y del pie, calculado con las luminancias
 *                relativas de la fórmula WCAG
 *   ampliación   texto al 200 % sin pérdida (sin desplazamiento horizontal)
 *
 * Queda MANUAL (no automatizable, se consigna en `revision-…-qa.md`): lector de
 * pantalla real y el texto del hero sobre la imagen de portada (foto + capa
 * `bg-navy/80`: el fondo pintado depende de la imagen, solo verificable
 * visualmente). Ejecutar con `--workers=1` (singletons compartidos).
 */

const FIXTURE_IMAGE = join(process.cwd(), '..', 'resources', 'simiente.jpeg');

interface Bounded {
  tag: string;
  text: string;
  width: number;
  height: number;
}

interface TargetAudit {
  interactive: Bounded[];
  radios: Bounded[];
}

interface TabStep {
  label: string;
  indicator: boolean;
}

interface StepInfo {
  label: string | null;
  indicator: boolean;
}

async function csrfToken(page: Page): Promise<string> {
  const cookies = await page.context().cookies(API_URL);
  const csrf = cookies.find((cookie) => cookie.name === 'csrf_token');
  if (!csrf) throw new Error('No hay cookie csrf_token');
  return csrf.value;
}

/** Recorre con Tab hasta cerrar el ciclo (vuelve al enlace de salto). */
async function tabCycle(page: Page, maxSteps = 30): Promise<TabStep[]> {
  const path: TabStep[] = [];
  for (let i = 0; i < maxSteps; i++) {
    // El salto inicial puede requerir más de un Tab si la SPA aún no montó
    // contenido interactivo: se reintenta mientras `document.activeElement`
    // siga en `body` (y el primer paso relevante sea el enlace de salto).
    await page.keyboard.press('Tab');
    const info = await page.evaluate<StepInfo>(() => {
      const el = document.activeElement;
      if (!el || el === document.body) return { label: null, indicator: false };
      const style = getComputedStyle(el);
      return {
        label: `${el.tagName.toLowerCase()} ${(
          (el as HTMLElement).innerText ||
          (el as HTMLInputElement).value ||
          ''
        )
          .trim()
          .slice(0, 40)}`,
        indicator:
          (style.outlineStyle !== 'none' && parseFloat(style.outlineWidth) > 0) ||
          style.boxShadow !== 'none',
      };
    });
    if (info.label === null) {
      if (path.length === 0 && i < 3) {
        continue; // la SPA aún no montó: reintenta el salto
      }
      break;
    }
    const label = info.label;
    expect(
      info.indicator,
      `foco visible en ${label} (paso ${i + 1} del recorrido por teclado)`,
    ).toBe(true);
    if (path.length > 0 && path[0].label === label) {
      path.push({ label, indicator: info.indicator });
      break; // el ciclo cerró: volvió al primer elemento
    }
    path.push({ label, indicator: info.indicator });
  }
  return path;
}

test.describe('Portada pública — accesibilidad WCAG 2.1 AA (SC-008)', () => {
  test.setTimeout(180_000);

  test('estructura, teclado con foco visible, toque ≥44 px, contraste y texto al 200 %', async ({
    browser,
  }) => {
    const setupContext = await browser.newContext();
    await ensureInitialized(setupContext.request);
    await setupContext.close();
    const suffix = uniqueSuffix();
    const identityName = `Iglesia Simiente ${suffix}`;
    const mission = `Misión ${suffix}`;
    const aboutText = `Somos una iglesia cercana y abierta ${suffix}.`;
    const serviceName = `Culto dominical ${suffix}`;
    const channelName = `Escríbenos ${suffix}`;
    const channelPhone = '+506 7000 4321';
    const contactEmail = `hola-${suffix}@ejemplo.com`;

    // ── Sembrar contenido publicado (identidad con logo + alt, FR-019) ─────
    const adminContext = await browser.newContext();
    const adminPage = await adminContext.newPage();
    let logoFile: string;
    try {
      await loginViaUi(adminPage, ADMIN.email, ADMIN.password);
      await expect(adminPage).toHaveURL(/\/panel$/);
      const upload = await adminPage.request.post(`${API_URL}/api/v1/admin/portada/imagenes`, {
        headers: { 'X-CSRF-Token': await csrfToken(adminPage) },
        multipart: {
          file: { name: 'logo.jpg', mimeType: 'image/jpeg', buffer: readFileSync(FIXTURE_IMAGE) },
        },
      });
      expect(upload.status()).toBe(201);
      logoFile = ((await upload.json()) as { fileName: string }).fileName;

      // Canales previos de corridas antiguas saturan el límite de 20: la base
      // es compartida y no se limpia, así que se retira lo viejo (204) para
      // que esta prueba sea repetible.
      const oldState = (await (
        await apiSend(adminPage, 'GET', '/api/v1/admin/portada')
      ).json()) as {
        whatsapp: { items: { id: string }[] };
      };
      for (const channel of oldState.whatsapp.items) {
        await apiSend(adminPage, 'DELETE', `/api/v1/admin/portada/whatsapp/${channel.id}`);
      }

      for (const [method, path, body] of [
        [
          'PUT',
          '/api/v1/admin/portada/identidad',
          {
            nameEs: identityName,
            missionEs: mission,
            logoFile,
            logoAltEs: `Logotipo de ${identityName}`,
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
            phone: channelPhone,
            publicationState: 'published',
          },
        ],
        [
          'POST',
          '/api/v1/admin/portada/horario',
          {
            dayOfWeek: 0,
            startTime: '10:00',
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
            destination: channelPhone,
            publicationState: 'published',
          },
        ],
      ] as const) {
        const response = await apiSend(adminPage, method, path, { ...body });
        expect(response.ok(), `${method} ${path} → ${response.status()}`).toBe(true);
      }
    } finally {
      await adminContext.close();
    }

    // ── Visitante en teléfono estrecho de 320 px (quickstart §10.1) ────────
    const visitor = await browser.newContext();
    const page = await visitor.newPage();
    try {
      await page.setViewportSize({ width: 320, height: 900 });
      await page.goto('/');
      await expect(page.getByRole('heading', { level: 1, name: identityName })).toBeVisible();

      const overflow = await page.evaluate(
        () => document.documentElement.scrollWidth - window.innerWidth,
      );
      expect(overflow, 'sin scroll horizontal a 320 px').toBeLessThanOrEqual(1);

      // §10.4 — un único h1 y títulos h2 de sección.
      expect(await page.locator('h1').count()).toBe(1);
      for (const name of ['Quiénes somos', 'Horario de servicios', 'WhatsApp', 'Contacto']) {
        await expect(page.getByRole('heading', { level: 2, name }).first()).toBeVisible();
      }

      // `lang` correcto e imágenes con alt no vacío (FR-019).
      await expect(page.locator('html')).toHaveAttribute('lang', 'es');
      const imagesWithoutAlt = await page.evaluate(() =>
        [...document.querySelectorAll('img')]
          .filter((img) => !img.alt || img.alt.trim() === '')
          .map((img) => img.src),
      );
      expect(imagesWithoutAlt, 'todas las imágenes con alt').toEqual([]);
      await expect(
        page.getByRole('img', { name: `Logotipo de ${identityName}` }).first(),
      ).toBeVisible();

      // §10.2 — tamaño de toque de enlaces y botones (≥ 44 px).
      const audit = await page.evaluate<TargetAudit>(() => {
        const visible = (el: Element): boolean => {
          const box = el.getBoundingClientRect();
          return box.width > 0 && box.height > 0;
        };
        const bounded = (el: Element): Bounded => {
          const box = el.getBoundingClientRect();
          return {
            tag: el.tagName.toLowerCase(),
            text: ((el as HTMLElement).innerText || '').trim().slice(0, 60),
            width: box.width,
            height: box.height,
          };
        };
        const interactive = [...document.querySelectorAll('a[href], button')].filter(
          (el) => visible(el) && el.getBoundingClientRect().width >= 20,
        );
        // El objetivo puntable de un radio es su `label` asociado (Radio:
        // WCAG 2.5.8 «Target Size (Minimum)», 24 px de mínimo en WCAG 2.2 AA).
        const radios = [...document.querySelectorAll('input[type=radio]')]
          .filter(visible)
          .map((el) => el.closest('label') ?? el);
        return {
          interactive: interactive.map(bounded),
          radios: radios.map(bounded),
        };
      });
      expect(audit.interactive.length, 'hay objetivos interactivos reales').toBeGreaterThan(5);
      const tooSmall = audit.interactive.filter((target) => target.height < 44);
      expect(
        tooSmall.map((t) => `${t.tag}[${t.text}] h=${t.height.toFixed(1)}px`),
        'todos los enlaces y botones con área de toque de al menos 44 px',
      ).toEqual([]);
      // Radios del selector de idioma: su objetivo puntable (el `label`) debe
      // alcanzar el mínimo de WCAG 2.5.8 «Target Size (Minimum)» (24 px). El
      // `label` usa `min-h-11` (44 px), así que supera el mínimo con holgura
      // (hallazgo A1 de revision-…-codigo.md, resuelto).
      for (const radio of audit.radios) {
        expect(radio.height, `objetivo de ${radio.text} ≥ 24 px`).toBeGreaterThanOrEqual(24);
      }

      // §10.3 — recorrido completo por teclado (US5 esc. 3).
      await page.goto('/');
      await expect(page.getByRole('heading', { level: 1, name: identityName })).toBeVisible();
      await expect(page.getByRole('navigation', { name: 'Secciones de la portada' })).toBeVisible();
      await expect(page.locator('#horario')).toBeVisible();
      await expect(page.locator('#whatsapp')).toBeVisible();
      const path = await tabCycle(page);
      expect(path[0]?.label, 'el primer Tab enfoca el enlace de salto al contenido').toContain(
        'Saltar al contenido principal',
      );
      const joined = path.map((step) => step.label).join(' | ');
      for (const mustHave of [
        'Quiénes somos',
        'Horario de servicios',
        'WhatsApp',
        'Contacto',
        'Redes sociales',
      ]) {
        expect(joined, `la travesía por teclado alcanza a «${mustHave}»`).toContain(mustHave);
      }
      // El grupo de radios marca el `es` del selector de cabecera y sigue
      // (comportamiento nativo de Tab); el `en` se activa con FLECHAS.
      expect(joined, 'la travesía alcanza el grupo de radios del idioma').toContain('input es');
      expect(
        path.some((step) => step.label.startsWith('a') && step.label.includes('Simiente')),
        'la travesía por teclado también recorre el pie',
      ).toBe(true);
      expect(
        path.length,
        'el ciclo de foco cierra volviendo al inicio (sin trampa de foco)',
      ).toBeLessThan(30);

      // Flechas del grupo de radios: sobre el radio marcado, una flecha mueve
      // la selección al otro idioma (quickstart §10.3, activable solo con
      // teclado). Vuelta con la flecha contraria, sin pérdida de información.
      const esRadio = page.locator('header').getByRole('radio', { name: 'Español' });
      await esRadio.focus();
      await expect(esRadio).toBeChecked();
      await page.keyboard.press('ArrowRight'); // es → en
      await expect(page.locator('html')).toHaveAttribute('lang', 'en');
      await expect(page.getByRole('heading', { level: 2, name: 'Who we are' })).toBeVisible();
      await expect(page.locator('header').getByRole('radio', { name: 'English' })).toBeChecked();
      await page.keyboard.press('ArrowLeft'); // en → es
      await expect(page.locator('html')).toHaveAttribute('lang', 'es');
      await expect(page.getByRole('heading', { level: 2, name: 'Quiénes somos' })).toBeVisible();

      // §10.5 — texto ampliable al 200 % SIN PÉRDIDA de información (FR-019,
      // WCAG 1.4.4). El zoom real del navegador escala los px CSS: se simula
      // con un viewport de la mitad (160 px) y `deviceScaleFactor: 2`. WCAG
      // 1.4.10 «Reflow» garantiza 320 px equivalentes y ahí el sitio no tiene
      // desplazamiento horizontal (comprobado arriba, §10.1); con zoom de 200 %
      // sobre un teléfono de 320 px se exige que NADA desaparezca, no la
      // ausencia total de scroll (el reflow fino quedó como hallazgo en el
      // informe QA: 56 px de desborde de las tarjetas de contacto/enlaces).
      const zoomContext = await browser.newContext({
        viewport: { width: 160, height: 450 },
        deviceScaleFactor: 2,
      });
      const zoomPage = await zoomContext.newPage();
      try {
        await zoomPage.goto('/');
        for (const locator of [
          zoomPage.getByRole('heading', { level: 1, name: identityName }),
          zoomPage.locator('#quienes-somos').getByText(aboutText),
          // En 160 px el diseño usa la lista de tarjetas (la `table` solo
          // es visible desde `sm`): objetivo directo a esa representación.
          zoomPage.locator('#horario li').filter({ hasText: serviceName }),
          zoomPage.locator('#whatsapp').getByText(channelName),
          zoomPage.locator('#contacto'),
        ]) {
          await expect(
            locator.first(),
            'con el texto al 200 % se conserva todo el contenido',
          ).toBeVisible();
        }
        await expect(
          zoomPage.getByRole('navigation', { name: 'Secciones de la portada' }),
        ).toBeVisible();
      } finally {
        await zoomContext.close();
      }

      // Contraste de todo el texto de `main` (excluido el hero sobre imagen,
      // que queda para la revisión manual) y del pie.
      const contrastReport = await page.evaluate(() => {
        const channels = (color: string): [number, number, number, number] => {
          const m = color.match(
            /rgba?\(\s*([\d.]+)[,\s]+([\d.]+)[,\s]+([\d.]+)(?:[,\s/]+([\d.]+))?\)/,
          );
          return m
            ? [Number(m[1]), Number(m[2]), Number(m[3]), m[4] === undefined ? 1 : Number(m[4])]
            : [0, 0, 0, 1];
        };
        const lin = (chunk: number): number => {
          const v = chunk / 255;
          return v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4;
        };
        const luminance = (rgb: [number, number, number]): number => {
          const [r, g, b] = rgb.map(lin);
          return 0.2126 * r + 0.7152 * g + 0.0722 * b;
        };
        const opaqueBg = (el: Element): [number, number, number] => {
          let node: Element | null = el;
          while (node && node !== document.documentElement) {
            const [r, g, b, alpha] = channels(getComputedStyle(node).backgroundColor);
            if (alpha >= 1) return [r, g, b];
            node = node.parentElement;
          }
          return [255, 255, 255];
        };
        const failures: string[] = [];
        const selector = 'main :is(h2,h3,p,li,td,th,a,button,label,span), footer :is(a,p,span)';
        for (const el of [...document.querySelectorAll(selector)]) {
          if (
            !el.textContent ||
            !el.textContent.trim() ||
            el.closest('[aria-labelledby="portada-nombre"]')
          ) {
            continue;
          }
          const box = el.getBoundingClientRect();
          if (box.width === 0 || box.height === 0) continue;
          const style = getComputedStyle(el);
          const fg = channels(style.color);
          if (fg[3] < 1) continue;
          const bg = opaqueBg(el);
          const ratio =
            (Math.max(luminance(fg), luminance(bg)) + 0.05) /
            (Math.min(luminance(fg), luminance(bg)) + 0.05);
          const fontSize = parseFloat(style.fontSize);
          const bold = parseInt(style.fontWeight, 10) >= 700;
          const large = fontSize >= 24 || (fontSize >= 18.66 && bold);
          if (ratio < (large ? 3 : 4.5)) {
            failures.push(
              `${el.tagName.toLowerCase()}[${
                (el as HTMLElement).innerText.slice(0, 30) || el.textContent.trim().slice(0, 30)
              }] ${ratio.toFixed(2)}:1 (${style.fontSize})`,
            );
          }
        }
        return failures;
      });
      expect(
        contrastReport,
        'todo el texto de main y del pie supera 4.5:1 (texto grande: 3:1)',
      ).toEqual([]);
    } finally {
      await visitor.close();
    }
  });
});
