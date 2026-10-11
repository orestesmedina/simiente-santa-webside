import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { WhoWeAreSection } from './WhoWeAreSection';

function renderSection(text = 'Somos una familia') {
  return render(
    <LanguageProvider>
      <WhoWeAreSection about={{ text }} />
    </LanguageProvider>,
  );
}

describe('WhoWeAreSection · propuesta B (T1022/FR-003)', () => {
  it('conserva el ancla, el h2 y el texto (contrato congelado)', () => {
    const { container } = renderSection();

    expect(container.querySelector('#quienes-somos')).not.toBeNull();
    expect(screen.getByRole('heading', { level: 2, name: 'Quiénes somos' })).toHaveAttribute(
      'id',
      'quienes-somos-title',
    );
    expect(screen.getByText('Somos una familia')).toBeInTheDocument();
  });

  it('usa tokens semánticos y el espaciado de sección (sin padding mágico)', () => {
    const { container } = renderSection();

    const section = container.querySelector('section') as HTMLElement;
    expect(section.className).toContain('py-section');
    expect(section.className).not.toMatch(/\bpy-(8|12|16|20|24|32)\b/);
    expect(section.querySelector('.bg-surface-raised')).not.toBeNull();
    expect(screen.getByText('Somos una familia').className).toContain('text-ink');
  });

  it('un texto largo se renderiza sin romper la composición', () => {
    renderSection('Línea uno\nLínea dos '.repeat(30));

    expect(screen.getByText(/Línea uno/)).toBeInTheDocument();
  });
});
