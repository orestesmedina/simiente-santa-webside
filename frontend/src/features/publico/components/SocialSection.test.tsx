import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { SocialLinkPublic } from '../../../api/portada';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { SocialSection } from './SocialSection';

const items: SocialLinkPublic[] = [
  { id: 'so-1', network: 'facebook', url: 'https://facebook.com/simiente' },
  { id: 'so-2', network: 'instagram', url: 'https://instagram.com/simiente' },
];

function renderSection() {
  return render(
    <LanguageProvider>
      <SocialSection items={items} />
    </LanguageProvider>,
  );
}

describe('SocialSection · propuesta B (T1024/FR-003)', () => {
  it('conserva los enlaces y nombres accesibles de F3', () => {
    renderSection();

    const facebook = screen.getByRole('link', { name: /Ver en Facebook/ });
    expect(facebook).toHaveAttribute('href', 'https://facebook.com/simiente');
    expect(facebook).toHaveAttribute('target', '_blank');
    expect(facebook).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('los chips tienen objetivo ≥ 44 px y elevación solo con transform', () => {
    renderSection();

    const chip = screen.getByRole('link', { name: /Ver en Facebook/ });
    expect(chip.className).toContain('min-h-11');
    expect(chip.className).toContain('hover:-translate-y-1');
    expect(chip.className).toContain('--motion-duration-micro');
    expect(chip.className).not.toMatch(/hover:m-|hover:p-/);
  });

  it('los iconos de red siguen siendo decorativos', () => {
    const { container } = renderSection();

    expect(container.querySelectorAll('svg[aria-hidden="true"]')).toHaveLength(2);
  });

  it('conserva el ancla, el h2 y usa tokens con espaciado de sección', () => {
    const { container } = renderSection();

    expect(container.querySelector('#redes')).not.toBeNull();
    expect(screen.getByRole('heading', { level: 2, name: 'Redes sociales' })).toHaveAttribute(
      'id',
      'redes-title',
    );
    expect((container.querySelector('section') as HTMLElement).className).toContain('py-section');
    expect(container.querySelector('.bg-surface-raised')).not.toBeNull();
  });
});
