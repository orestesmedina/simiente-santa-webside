import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { en } from './en';
import { es } from './es';
import { LanguageProvider } from './LanguageProvider';
import { LanguageSwitcher } from './LanguageSwitcher';
import { LANGUAGE_STORAGE_KEY, publicMessageKeys } from './messages';
import { useLanguage } from './useLanguage';

function Consumer() {
  const { lang, t } = useLanguage();
  return (
    <div>
      <p data-testid="lang">{lang}</p>
      <p data-testid="who">{t('nav.sections.who')}</p>
      <p data-testid="social">{t('social.action', { network: 'Facebook' })}</p>
    </div>
  );
}

function renderApp() {
  return render(
    <LanguageProvider>
      <LanguageSwitcher />
      <Consumer />
    </LanguageProvider>,
  );
}

describe('i18n del sitio público', () => {
  beforeEach(() => {
    window.localStorage.clear();
    document.documentElement.lang = 'es';
  });

  afterEach(() => {
    window.localStorage.clear();
  });

  it('los dos diccionarios tienen exactamente las mismas claves del catálogo', () => {
    expect(Object.keys(es).sort()).toEqual([...publicMessageKeys].sort());
    expect(Object.keys(en).sort()).toEqual([...publicMessageKeys].sort());
  });

  it('incluye las claves de marca de ux.md §7.2 (brand.name, footer.welcome, schedule.day.0…6)', () => {
    for (const key of [
      'brand.name',
      'footer.welcome',
      'schedule.day.0',
      'schedule.day.6',
    ] as const) {
      expect(es[key]).toBeTruthy();
      expect(en[key]).toBeTruthy();
    }
  });

  it('la primera visita sin preferencia guardada entra en español', () => {
    renderApp();
    expect(screen.getByTestId('lang')).toHaveTextContent('es');
    expect(screen.getByTestId('who')).toHaveTextContent('Quiénes somos');
  });

  it('respeta la preferencia guardada en una re-visita (localStorage)', () => {
    window.localStorage.setItem(LANGUAGE_STORAGE_KEY, 'en');
    renderApp();
    expect(screen.getByTestId('lang')).toHaveTextContent('en');
    expect(screen.getByTestId('who')).toHaveTextContent('Who we are');
  });

  it('el selector cambia la interfaz al instante, persiste y actualiza <html lang>', async () => {
    renderApp();

    // En español los radios muestran «Español» / «Inglés» (catálogo ux.md §7.2).
    await userEvent.click(screen.getByRole('radio', { name: 'Inglés' }));

    expect(screen.getByTestId('lang')).toHaveTextContent('en');
    expect(screen.getByTestId('who')).toHaveTextContent('Who we are');
    expect(window.localStorage.getItem(LANGUAGE_STORAGE_KEY)).toBe('en');
    expect(document.documentElement.lang).toBe('en');

    // Ya en inglés, los radios muestran «Spanish» / «English».
    await userEvent.click(screen.getByRole('radio', { name: 'Spanish' }));
    expect(screen.getByTestId('who')).toHaveTextContent('Quiénes somos');
    expect(window.localStorage.getItem(LANGUAGE_STORAGE_KEY)).toBe('es');
    expect(document.documentElement.lang).toBe('es');
  });

  it('interpola las variables de las plantillas traducidas', () => {
    renderApp();
    expect(screen.getByTestId('social')).toHaveTextContent('Ver en Facebook');
  });

  it('dos selectores en la misma página no comparten grupo nativo (cabecera y pie)', () => {
    // La portada monta el selector dos veces; con el mismo `name` el navegador
    // los uniría en un solo grupo y dejaría sin marcar uno de los dos.
    render(
      <LanguageProvider>
        <LanguageSwitcher />
        <LanguageSwitcher />
      </LanguageProvider>,
    );

    const spanish = screen.getAllByRole('radio', { name: 'Español' });
    expect(spanish).toHaveLength(2);
    expect(spanish[0].getAttribute('name')).not.toBe(spanish[1].getAttribute('name'));
    expect(spanish[0]).toBeChecked();
    expect(spanish[1]).toBeChecked();
  });

  it('expone el idioma base sin auto-detectar el navegador', () => {
    // Aunque el navegador (jsdom) esté en otro idioma, sin preferencia guardada
    // el sitio entra en español (FR-010: sin `navigator.language`).
    window.localStorage.clear();
    renderApp();
    expect(screen.getByTestId('lang')).toHaveTextContent('es');
  });
});

describe('claves i18n nuevas de F10 (FR-019/SC-010)', () => {
  const CLAVES_NUEVAS = [
    'nav.menu.open',
    'nav.menu.close',
    'nav.menu.label',
    'footer.sitemap.label',
    'footer.sitemap.explore',
    'footer.sitemap.contact',
    'footer.sitemap.social',
    'footer.sitemap.home',
    'theme.label',
    'theme.toDark',
    'theme.toLight',
  ] as const;

  it('existen en es y en, y ninguna queda vacía', () => {
    for (const key of CLAVES_NUEVAS) {
      expect(es[key], `es.${key}`).toBeTruthy();
      expect(en[key], `en.${key}`).toBeTruthy();
      expect(es[key].trim().length, `es.${key} no vacía`).toBeGreaterThan(0);
      expect(en[key].trim().length, `en.${key} no vacía`).toBeGreaterThan(0);
    }
  });

  it('las claves congeladas de F3 conservan su texto (contrato R10-9)', () => {
    expect(es['nav.label']).toBe('Secciones de la portada');
    expect(en['nav.label']).toBe('Page sections');
    expect(es['a11y.skip']).toBe('Saltar al contenido principal');
    expect(en['a11y.skip']).toBe('Skip to main content');
    expect(es['lang.label']).toBe('Idioma');
    expect(en['lang.label']).toBe('Language');
  });

  it('los textos nuevos son los del catálogo de ux.md §7', () => {
    expect(es['nav.menu.open']).toBe('Abrir menú de secciones');
    expect(en['nav.menu.open']).toBe('Open sections menu');
    expect(es['nav.menu.close']).toBe('Cerrar menú de secciones');
    expect(en['nav.menu.close']).toBe('Close sections menu');
    expect(es['nav.menu.label']).toBe('Menú de secciones');
    expect(en['nav.menu.label']).toBe('Sections menu');
    expect(es['footer.sitemap.label']).toBe('Mapa del sitio');
    expect(en['footer.sitemap.label']).toBe('Site map');
    expect(es['footer.sitemap.explore']).toBe('Explorar');
    expect(en['footer.sitemap.explore']).toBe('Explore');
    expect(es['footer.sitemap.contact']).toBe('Contacto');
    expect(en['footer.sitemap.contact']).toBe('Contact');
    expect(es['footer.sitemap.social']).toBe('Síguenos');
    expect(en['footer.sitemap.social']).toBe('Follow us');
    expect(es['footer.sitemap.home']).toBe('Inicio');
    expect(en['footer.sitemap.home']).toBe('Home');
    expect(es['theme.label']).toBe('Tema');
    expect(en['theme.label']).toBe('Theme');
    expect(es['theme.toDark']).toBe('Cambiar a tema oscuro');
    expect(en['theme.toDark']).toBe('Switch to dark theme');
    expect(es['theme.toLight']).toBe('Cambiar a tema claro');
    expect(en['theme.toLight']).toBe('Switch to light theme');
  });
});
