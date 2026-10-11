import { act, render, screen, within } from '@testing-library/react';
import type { ComponentProps } from 'react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import type { IdentityPublic } from '../../../api/portada';
import type { SiteLink } from '../hooks/useSiteLinks';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { SiteHeader } from './SiteHeader';

const identity: IdentityPublic = {
  name: 'Iglesia Simiente Santa',
  logoUrl: '/api/v1/media/img_logo.png',
  logoAlt: 'Logotipo de la iglesia',
};

const links: SiteLink[] = [
  { id: 'who', anchor: 'quienes-somos', messageKey: 'nav.sections.who', href: '#quienes-somos' },
  { id: 'schedule', anchor: 'horario', messageKey: 'nav.sections.schedule', href: '#horario' },
];

function renderHeader(props: Partial<ComponentProps<typeof SiteHeader>> = {}) {
  return render(
    <MemoryRouter>
      <LanguageProvider>
        <SiteHeader identity={identity} links={links} {...props} />
      </LanguageProvider>
    </MemoryRouter>,
  );
}

describe('SiteHeader · cabecera fija del sitio público (T1011)', () => {
  it('expone el <nav> con el nombre accesible congelado de F3', () => {
    renderHeader();

    expect(screen.getByRole('navigation', { name: 'Secciones de la portada' })).toBeInTheDocument();
  });

  it('pinta las anclas de las secciones presentes con sus nombres', () => {
    renderHeader();

    const nav = screen.getByRole('navigation', { name: 'Secciones de la portada' });
    expect(within(nav).getByRole('link', { name: 'Quiénes somos' })).toHaveAttribute(
      'href',
      '#quienes-somos',
    );
    expect(within(nav).getByRole('link', { name: 'Horario de servicios' })).toBeInTheDocument();
    // Las secciones ausentes no dejan ancla (hereda el criterio de F3).
    expect(within(nav).queryByRole('link', { name: 'Contacto' })).toBeNull();
  });

  it('el enlace de marca conserva el nombre completo con «Simiente» (D20/C3)', () => {
    renderHeader();

    const brand = screen.getByRole('link', { name: /Simiente/ });
    expect(brand).toHaveAttribute('href', '/');
    expect(brand.textContent).toContain('Iglesia Simiente Santa');

    const word = screen.getByText('Iglesia Simiente Santa');
    // Solo se oculta visualmente a < 480 px: nunca `display:none` (el innerText
    // lo exige el `tabCycle` congelado a 320 px).
    expect(word.className).toContain('sr-only');
    expect(word.className).not.toContain('display-none');
  });

  it('ordena los bloques marca → nav → controles', () => {
    const { container } = renderHeader();

    const header = container.querySelector('header');
    expect(header).not.toBeNull();
    const brand = screen.getByRole('link', { name: /Simiente/ });
    const nav = screen.getByRole('navigation', { name: 'Secciones de la portada' });
    const lang = screen.getByRole('group', { name: 'Idioma' });

    // brand aparece antes que nav y este antes que el clúster de controles.
    expect(brand.compareDocumentPosition(nav) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(nav.compareDocumentPosition(lang) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });

  it('renderiza el slot de acciones (control de tema) tras el selector', () => {
    renderHeader({ actions: <button type="button">tema</button> });

    expect(screen.getByRole('button', { name: 'tema' })).toBeInTheDocument();
  });

  it('usa tokens semánticos de superficie (funciona en ambos temas)', () => {
    const { container } = renderHeader();

    const header = container.querySelector('header') as HTMLElement;
    expect(header.className).toContain('bg-surface');
    expect(header.className).toContain('text-ink');
    expect(header.className).toContain('sticky');
  });

  it('publica --header-offset con la altura medida y la re-mide al cambiar (T1014)', () => {
    class ResizeObserverStub {
      static latest: ResizeObserverStub | null = null;
      callback: ResizeObserverCallback;
      constructor(callback: ResizeObserverCallback) {
        this.callback = callback;
        ResizeObserverStub.latest = this;
      }
      observe() {}
      unobserve() {}
      disconnect() {}
      trigger() {
        this.callback([], this as unknown as ResizeObserver);
      }
    }
    (globalThis as { ResizeObserver?: unknown }).ResizeObserver = ResizeObserverStub;

    let height = 88;
    const rectSpy = vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(
      () =>
        ({
          height,
          width: 0,
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          x: 0,
          y: 0,
          toJSON: () => ({}),
        }) as DOMRect,
    );

    try {
      renderHeader();
      expect(document.documentElement.style.getPropertyValue('--header-offset')).toBe('88px');

      height = 120;
      act(() => ResizeObserverStub.latest?.trigger());
      expect(document.documentElement.style.getPropertyValue('--header-offset')).toBe('120px');
    } finally {
      rectSpy.mockRestore();
      delete (globalThis as { ResizeObserver?: unknown }).ResizeObserver;
      document.documentElement.style.removeProperty('--header-offset');
    }
  });
});
