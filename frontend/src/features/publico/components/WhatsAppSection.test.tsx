import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { WhatsappChannelPublic } from '../../../api/portada';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { WhatsAppSection } from './WhatsAppSection';

const items: WhatsappChannelPublic[] = [
  { id: 'wa-1', name: 'Escríbenos', kind: 'direct', url: 'https://wa.me/50688888888' },
  { id: 'wa-2', name: 'Grupo de la iglesia', kind: 'group', url: 'https://chat.whatsapp.com/ABC' },
];

function renderSection() {
  return render(
    <LanguageProvider>
      <WhatsAppSection items={items} />
    </LanguageProvider>,
  );
}

describe('WhatsAppSection · propuesta B (T1023/FR-003)', () => {
  it('los enlaces apuntan al destino y abren fuera (contrato de F3 intacto)', () => {
    renderSection();

    const dm = screen.getByRole('link', { name: /Escribir por WhatsApp/ });
    expect(dm).toHaveAttribute('href', 'https://wa.me/50688888888');
    expect(dm).toHaveAttribute('target', '_blank');
    expect(dm).toHaveAttribute('rel', 'noopener noreferrer');

    expect(screen.getByRole('link', { name: /Entrar al grupo/ })).toHaveAttribute(
      'href',
      'https://chat.whatsapp.com/ABC',
    );
  });

  it('usa el CTA verde del sistema con objetivo ≥ 44 px', () => {
    renderSection();

    const dm = screen.getByRole('link', { name: /Escribir por WhatsApp/ });
    expect(dm.className).toContain('bg-leaf');
    expect(dm.className).toContain('min-h-11');
    expect(dm.className).toContain('--motion-duration-micro');
  });

  it('conserva el ancla, el h2 y el nombre del canal', () => {
    const { container } = renderSection();

    expect(container.querySelector('#whatsapp')).not.toBeNull();
    expect(screen.getByRole('heading', { level: 2, name: 'WhatsApp' })).toHaveAttribute(
      'id',
      'whatsapp-title',
    );
    expect(screen.getByText('Escríbenos')).toBeInTheDocument();
  });

  it('usa tokens y el espaciado de sección', () => {
    const { container } = renderSection();

    expect((container.querySelector('section') as HTMLElement).className).toContain('py-section');
    expect(container.querySelector('.bg-surface-raised')).not.toBeNull();
  });
});
