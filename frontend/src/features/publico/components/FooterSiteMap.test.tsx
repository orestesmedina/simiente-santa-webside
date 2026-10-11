import { render, screen, within } from '@testing-library/react';
import type { ComponentProps } from 'react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import type { SiteLink } from '../hooks/useSiteLinks';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { FooterSiteMap } from './FooterSiteMap';

const links: SiteLink[] = [
  { id: 'who', anchor: 'quienes-somos', messageKey: 'nav.sections.who', href: '#quienes-somos' },
  { id: 'schedule', anchor: 'horario', messageKey: 'nav.sections.schedule', href: '#horario' },
  { id: 'whatsapp', anchor: 'whatsapp', messageKey: 'nav.sections.whatsapp', href: '#whatsapp' },
  { id: 'contact', anchor: 'contacto', messageKey: 'nav.sections.contact', href: '#contacto' },
  { id: 'social', anchor: 'redes', messageKey: 'nav.sections.social', href: '#redes' },
];

const contact = {
  address: 'San José, Costa Rica',
  email: 'hola@ejemplo.com',
  phone: '+506 8888 8888',
};

const whatsapp = [
  { id: 'wa-1', name: 'Escríbenos', kind: 'direct' as const, url: 'https://wa.me/50688888888' },
  { id: 'wa-2', name: 'Grupo', kind: 'group' as const, url: 'https://chat.whatsapp.com/ABC' },
];

const socials = [
  { id: 'so-1', network: 'facebook' as const, url: 'https://facebook.com/simiente' },
  { id: 'so-2', network: 'instagram' as const, url: 'https://instagram.com/simiente' },
];

function renderMap(props: Partial<ComponentProps<typeof FooterSiteMap>> = {}) {
  return render(
    <MemoryRouter>
      <LanguageProvider>
        <FooterSiteMap
          links={links}
          contact={contact}
          socials={socials}
          whatsapp={whatsapp}
          {...props}
        />
      </LanguageProvider>
    </MemoryRouter>,
  );
}

describe('FooterSiteMap · mapa del sitio (T1015/FR-027)', () => {
  it('usa un landmark con nombre accesible propio, distinto del de la cabecera', () => {
    renderMap();

    expect(screen.getByRole('navigation', { name: 'Mapa del sitio' })).toBeInTheDocument();
  });

  it('lista las secciones presentes con los mismos nombres y anclas que la cabecera', () => {
    renderMap();

    const nav = screen.getByRole('navigation', { name: 'Mapa del sitio' });
    expect(within(nav).getByRole('link', { name: 'Inicio' })).toHaveAttribute('href', '/');

    const esperados: [string, string][] = [
      ['Quiénes somos', '#quienes-somos'],
      ['Horario de servicios', '#horario'],
      ['WhatsApp', '#whatsapp'],
      ['Contacto', '#contacto'],
      ['Redes sociales', '#redes'],
    ];
    for (const [name, href] of esperados) {
      expect(within(nav).getByRole('link', { name })).toHaveAttribute('href', href);
    }
  });

  it('una sección en borrador (no incluida en links) no deja enlace', () => {
    renderMap({ links: links.filter((link) => link.id !== 'whatsapp') });

    const nav = screen.getByRole('navigation', { name: 'Mapa del sitio' });
    expect(within(nav).queryByRole('link', { name: 'WhatsApp' })).toBeNull();
  });

  it('los rótulos de grupo son Explorar/Contacto/Síguenos y NO tienen rol heading', () => {
    renderMap();

    for (const label of ['Explorar', 'Contacto', 'Síguenos']) {
      // Rótulo de grupo = etiqueta no-heading (`<p>`), nunca un encabezado.
      expect(screen.getByText(label, { selector: 'p' })).toBeInTheDocument();
      expect(screen.queryByRole('heading', { name: label })).toBeNull();
    }
  });

  it('con datos vacíos no dibuja el mapa', () => {
    const { container } = renderMap({
      links: [],
      contact: undefined,
      socials: [],
      whatsapp: [],
    });

    expect(screen.queryByRole('navigation', { name: 'Mapa del sitio' })).toBeNull();
    expect(container.textContent).toBe('');
  });

  it('sin datos de contacto el grupo Contacto no aparece (pero Explorar y Síguenos sí)', () => {
    renderMap({
      contact: undefined,
      whatsapp: [],
      links: links.filter((link) => link.id !== 'contact'),
    });

    expect(screen.queryByText('Contacto')).toBeNull();
    expect(screen.getByText('Explorar')).toBeInTheDocument();
    expect(screen.getByText('Síguenos')).toBeInTheDocument();
  });

  it('los nombres accesibles del pie son únicos y no duplican los de las secciones (D19/C2)', () => {
    renderMap();

    const nav = screen.getByRole('navigation', { name: 'Mapa del sitio' });
    const names = within(nav)
      .getAllByRole('link')
      .map((link) => link.getAttribute('aria-label') ?? link.textContent ?? '');

    expect(new Set(names).size).toBe(names.length);
    expect(names).not.toContain('hola@ejemplo.com');
    expect(names).not.toContain('+506 8888 8888');
    expect(names.some((name) => /Ver en /.test(name))).toBe(false);
    expect(names.some((name) => /Escribir por WhatsApp|Entrar al grupo/.test(name))).toBe(false);
    expect(names).toContain('Contacto: Correo');
    expect(names).toContain('Contacto: Teléfono');
    expect(names).toContain('Contacto: WhatsApp');
    expect(names).toContain('Facebook');
    expect(names).toContain('Instagram');
  });

  it('la dirección del grupo Contacto es texto, no un enlace', () => {
    renderMap();

    expect(screen.getByText(/San José, Costa Rica/)).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /San José/ })).toBeNull();
  });

  it('«Contacto: WhatsApp» es un solo enlace (el primer canal publicado)', () => {
    renderMap();

    const whatsappLinks = screen.getAllByRole('link', { name: 'Contacto: WhatsApp' });
    expect(whatsappLinks).toHaveLength(1);
    expect(whatsappLinks[0]).toHaveAttribute('href', 'https://wa.me/50688888888');
  });
});
