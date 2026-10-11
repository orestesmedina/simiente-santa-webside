import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { SectionSkeleton } from './SectionSkeleton';

function renderSkeleton(lines?: number) {
  return render(
    <LanguageProvider>
      <SectionSkeleton lines={lines} />
    </LanguageProvider>,
  );
}

describe('SectionSkeleton · propuesta B (T1024/FR-009)', () => {
  it('anuncia la carga (role=status) y conserva el espaciado de sección', () => {
    const { container } = renderSkeleton();

    expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
    expect(screen.getByText('Cargando…')).toBeInTheDocument();
    const status = container.querySelector('[role="status"]') as HTMLElement;
    expect(status.className).toContain('py-section');
    expect(container.querySelector('.bg-surface-raised')).not.toBeNull();
  });

  it('el pulso es sutil y solo con movimiento (motion-safe), sobre fantasmas decorativos', () => {
    const { container } = renderSkeleton(2);

    const ghosts = container.querySelectorAll('[aria-hidden="true"]');
    expect(ghosts).toHaveLength(3);
    for (const ghost of ghosts) {
      expect(ghost.className).toContain('motion-safe:animate-pulse');
    }
  });
});
