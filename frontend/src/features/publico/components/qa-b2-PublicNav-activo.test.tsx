import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import type { SiteLink } from '../hooks/useSiteLinks';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { PublicNav } from './PublicNav';

/**
 * QA · Bug 2 (⟲ 004-eventos-actividades): el menú no tiene estados de hover ni
 * marca la opción actual. Causa raíz: `PublicNav.tsx` renderiza los `<a>` sin
 * clases `hover:` ni calcula el enlace correspondiente a la ruta activa
 * (`aria-current="page"`). Lot de este test:
 *
 * 1. el enlace de la ruta actual (`/eventos` → «Eventos y actividades») lleva
 *    `aria-current="page"` y los demás no llevan `aria-current`;
 * 2. cada enlace tiene clases de hover (`hover:`), que es lo que da la
 *    retroalimentación visual al pasar el cursor.
 */

const links: SiteLink[] = [
  { id: 'who', anchor: 'quienes-somos', messageKey: 'nav.sections.who', href: '/#quienes-somos' },
  { id: 'whatsapp', anchor: 'whatsapp', messageKey: 'nav.sections.whatsapp', href: '/#whatsapp' },
  { id: 'events', messageKey: 'nav.sections.events', href: '/eventos' },
];

function renderNavEn(ruta: string) {
  return render(
    <LanguageProvider>
      <MemoryRouter initialEntries={[ruta]}>
        <PublicNav links={links} />
      </MemoryRouter>
    </LanguageProvider>,
  );
}

describe('PublicNav · QA Bug 2: estado activo y hover', () => {
  it('el enlace de la ruta actual lleva aria-current="page" (en /eventos)', () => {
    renderNavEn('/eventos');

    const eventos = screen.getByRole('link', { name: 'Eventos y actividades' });
    expect(eventos).toHaveAttribute('aria-current', 'page');

    expect(screen.getByRole('link', { name: 'Quiénes somos' })).not.toHaveAttribute('aria-current');
    expect(screen.getByRole('link', { name: 'WhatsApp' })).not.toHaveAttribute('aria-current');
  });

  it('la portada no marca ninguna opción de ruta como actual', () => {
    renderNavEn('/');

    expect(screen.getByRole('link', { name: 'Eventos y actividades' })).not.toHaveAttribute(
      'aria-current',
    );
  });

  it('los enlaces de la barra tienen clases de hover', () => {
    renderNavEn('/eventos');

    for (const name of ['Quiénes somos', 'WhatsApp', 'Eventos y actividades']) {
      expect(
        screen.getByRole('link', { name }).className,
        `el enlace «${name}» debe tener estado :hover`,
      ).toMatch(/hover:/);
    }
  });
});
