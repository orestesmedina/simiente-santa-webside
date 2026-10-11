import { render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { SiteLink } from '../hooks/useSiteLinks';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { PublicNav } from './PublicNav';

const links: SiteLink[] = [
  { id: 'who', anchor: 'quienes-somos', messageKey: 'nav.sections.who', href: '#quienes-somos' },
  { id: 'schedule', anchor: 'horario', messageKey: 'nav.sections.schedule', href: '#horario' },
  { id: 'contact', anchor: 'contacto', messageKey: 'nav.sections.contact', href: '#contacto' },
];

function renderNav(extra: SiteLink[] = []) {
  return render(
    <LanguageProvider>
      <PublicNav links={[...links, ...extra]} />
    </LanguageProvider>,
  );
}

describe('PublicNav · anclas de sección (T1012/FR-011/FR-012)', () => {
  it('pinta cada sección una sola vez dentro del landmark nombrado', () => {
    renderNav();

    const nav = screen.getByRole('navigation', { name: 'Secciones de la portada' });
    expect(within(nav).getAllByRole('link', { name: 'Quiénes somos' })).toHaveLength(1);
    expect(within(nav).getAllByRole('link', { name: 'Horario de servicios' })).toHaveLength(1);
    expect(within(nav).getAllByRole('link', { name: 'Contacto' })).toHaveLength(1);
  });

  it('mantiene los enlaces en el DOM sin atributos que los saquen de jsdom (D14)', () => {
    renderNav();

    for (const name of ['Quiénes somos', 'Horario de servicios', 'Contacto']) {
      const link = screen.getByRole('link', { name });
      expect(link).not.toHaveAttribute('hidden');
      expect(link).not.toHaveAttribute('aria-hidden');
      expect(link).not.toHaveAttribute('inert');
      expect(link.parentElement).not.toHaveAttribute('hidden');
    }
  });

  it('la lista usa la clase de colapso (display:none a ≤ 860 px) y no hace scroll lateral', () => {
    const { container } = renderNav();

    const list = container.querySelector('ul.public-nav-list');
    expect(list).not.toBeNull();
    expect(list?.className).not.toContain('overflow-x-auto');
    // Sin el menú abierto, la lista queda colapsada en móvil (D16).
    expect(list).not.toHaveAttribute('data-open', 'true');
  });

  it('los enlaces tienen objetivo táctil ≥ 44 px', () => {
    renderNav();

    expect(screen.getByRole('link', { name: 'Contacto' }).className).toContain('min-h-11');
  });

  it('escala a F4–F9: una lista larga sigue íntegra en filas flex-wrap', () => {
    const extra: SiteLink[] = Array.from({ length: 5 }, (_, index) => ({
      id: 'who',
      anchor: `futura-${index}`,
      messageKey: 'nav.sections.who',
      href: `#futura-${index}`,
    }));
    const { container } = renderNav(extra);

    const list = container.querySelector('ul.public-nav-list');
    expect(list?.querySelectorAll('li')).toHaveLength(links.length + extra.length);
    expect(list?.className).toContain('public-nav-list');
  });
});
