import { render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { IdentityPublic } from '../../../api/portada';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { IdentityHero } from './IdentityHero';

const identity: IdentityPublic = {
  name: 'Iglesia Simiente Santa',
  tagline: 'Un espacio para encontrarse con Dios',
  mission: 'Nuestra misión real',
  vision: 'Nuestra visión real',
  logoUrl: '/api/v1/media/img_logo.png',
  logoAlt: 'Logotipo de la iglesia',
  coverImageUrl: '/api/v1/media/img_cover.png',
  coverImageAlt: 'Imagen de la portada',
};

function renderHero(options: { whatsappHref?: string | null; scheduleHref?: string | null } = {}) {
  const { whatsappHref = '#whatsapp', scheduleHref = '#horario' } = options;
  return render(
    <LanguageProvider>
      <IdentityHero
        identity={identity}
        whatsappHref={whatsappHref ?? undefined}
        scheduleHref={scheduleHref ?? undefined}
      />
    </LanguageProvider>,
  );
}

describe('IdentityHero · propuesta B (T1021/FR-001/FR-002)', () => {
  it('conserva el contrato congelado de F3 (region, h1, imágenes)', () => {
    renderHero();

    const hero = screen.getByRole('region', { name: 'Iglesia Simiente Santa' });
    expect(
      within(hero).getByRole('heading', { level: 1, name: 'Iglesia Simiente Santa' }),
    ).toHaveAttribute('id', 'portada-nombre');
    expect(within(hero).getByRole('img', { name: 'Logotipo de la iglesia' })).toBeInTheDocument();
    expect(within(hero).getByRole('img', { name: 'Imagen de la portada' })).toBeInTheDocument();
    expect(within(hero).getByText('Un espacio para encontrarse con Dios')).toBeInTheDocument();
  });

  it('los CTAs usan PublicButton y no colisionan con los canales de WhatsApp (D19)', () => {
    renderHero();

    const hero = screen.getByRole('region', { name: 'Iglesia Simiente Santa' });
    expect(within(hero).getByRole('link', { name: 'Ver horario' })).toHaveAttribute(
      'href',
      '#horario',
    );
    expect(within(hero).getByRole('link', { name: 'Escríbenos por WhatsApp' })).toHaveAttribute(
      'href',
      '#whatsapp',
    );

    const names = within(hero)
      .getAllByRole('link')
      .map((link) => link.textContent ?? '');
    expect(names.some((name) => /Escribir por WhatsApp/.test(name))).toBe(false);
    expect(names.some((name) => /Entrar al grupo/.test(name))).toBe(false);
  });

  it('sin canal de WhatsApp publicado no pinta el CTA (0 enlaces rotos)', () => {
    renderHero({ whatsappHref: null });

    expect(screen.queryByRole('link', { name: /WhatsApp/ })).toBeNull();
  });

  it('sin horario publicado no pinta el CTA «Ver horario» (0 enlaces rotos, I1)', () => {
    renderHero({ scheduleHref: null });

    const hero = screen.getByRole('region', { name: 'Iglesia Simiente Santa' });
    expect(within(hero).queryByRole('link', { name: 'Ver horario' })).toBeNull();
    expect(within(hero).queryByRole('link', { name: '#horario' })).toBeNull();
    // El resto del hero y el CTA de WhatsApp siguen ahí.
    expect(within(hero).getByRole('link', { name: 'Escríbenos por WhatsApp' })).toBeInTheDocument();
  });

  it('usa tokens semánticos y la entrada del hero (no colores explícitos de F3)', () => {
    const { container } = renderHero();

    const section = container.querySelector('section') as HTMLElement;
    expect(section.className).toContain('hero-enter');
    expect(section.className).not.toContain('bg-white');
    expect(section.className).not.toContain('text-navy');
    expect(screen.getByRole('heading', { level: 1 }).className).toContain('text-ink');
  });

  it('el logotipo no recibe clases de animación ni de tema', () => {
    renderHero();

    const logo = screen.getByRole('img', { name: 'Logotipo de la iglesia' });
    expect(logo.className).not.toMatch(/animate|hero-enter|reveal|rotate|invert/);
    expect(logo.className).toContain('bg-logo-surface');
  });
});
