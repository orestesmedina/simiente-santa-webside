import { render, screen, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import type { IdentityPublic } from '../../../api/portada';
import type { SiteLink } from '../hooks/useSiteLinks';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { PublicFooter } from './PublicFooter';

const identity: IdentityPublic = {
  name: 'Iglesia Simiente Santa',
  logoUrl: '/api/v1/media/img_logo.png',
  logoAlt: 'Logotipo de la iglesia',
};

const links: SiteLink[] = [
  { id: 'who', anchor: 'quienes-somos', messageKey: 'nav.sections.who', href: '#quienes-somos' },
  { id: 'contact', anchor: 'contacto', messageKey: 'nav.sections.contact', href: '#contacto' },
];

const contact = {
  address: 'San José, Costa Rica',
  email: 'hola@ejemplo.com',
  phone: '+506 8888 8888',
};

const socials = [
  { id: 'so-1', network: 'facebook' as const, url: 'https://facebook.com/simiente' },
];

function renderFooter() {
  return render(
    <MemoryRouter>
      <LanguageProvider>
        <PublicFooter identity={identity} links={links} contact={contact} socials={socials} />
      </LanguageProvider>
    </MemoryRouter>,
  );
}

describe('PublicFooter · propuesta B (T1016/FR-027)', () => {
  it('es un único landmark contentinfo que contiene el mapa del sitio', () => {
    const { container } = renderFooter();

    expect(container.querySelectorAll('footer')).toHaveLength(1);
    const footer = screen.getByRole('contentinfo');
    expect(within(footer).getByRole('navigation', { name: 'Mapa del sitio' })).toBeInTheDocument();
    expect(within(footer).getByRole('link', { name: 'Quiénes somos' })).toBeInTheDocument();
  });

  it('la franja inferior contiene el lema y el selector de idioma (F3 intacto)', () => {
    renderFooter();

    expect(screen.getByText('Esta siempre será tu casa.')).toBeInTheDocument();
    expect(screen.getByRole('group', { name: 'Idioma' })).toBeInTheDocument();
    expect(screen.getAllByRole('radio', { name: 'Español' })).toHaveLength(1);
  });

  it('la marca del pie es texto, no un enlace (D20/C3)', () => {
    renderFooter();

    const brand = screen.getByText('Iglesia Simiente Santa');
    expect(brand.tagName).toBe('SPAN');
    expect(screen.queryByRole('link', { name: /Iglesia Simiente Santa/ })).toBeNull();
  });

  it('usa tokens semánticos de tema y espaciado de sección', () => {
    const { container } = renderFooter();

    const footer = container.querySelector('footer') as HTMLElement;
    expect(footer.className).toContain('bg-surface');
    expect(footer.className).toContain('text-ink');
    expect(container.querySelector('.py-section-tight')).not.toBeNull();
  });
});
