import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { ContactPublic } from '../../../api/portada';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { ContactSection } from './ContactSection';

const contact: ContactPublic = {
  address: 'San José, Costa Rica',
  email: 'hola@ejemplo.com',
  phone: '+506 8888 8888',
};

function renderSection() {
  return render(
    <LanguageProvider>
      <ContactSection contact={contact} />
    </LanguageProvider>,
  );
}

describe('ContactSection · propuesta B (T1023/FR-003)', () => {
  it('el contacto ofrece enlaces nativos de correo y teléfono (contrato de F3)', () => {
    renderSection();

    expect(screen.getByRole('link', { name: 'hola@ejemplo.com' })).toHaveAttribute(
      'href',
      'mailto:hola@ejemplo.com',
    );
    expect(screen.getByRole('link', { name: '+506 8888 8888' })).toHaveAttribute(
      'href',
      'tel:+50688888888',
    );
  });

  it('la dirección sigue siendo texto, no un enlace', () => {
    renderSection();

    expect(screen.getByText('San José, Costa Rica')).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: /San José/ })).toBeNull();
  });

  it('las filas-pastilla se desplazan en hover solo con transform', () => {
    const { container } = renderSection();

    const rows = container.querySelectorAll('#contacto dl > div');
    expect(rows).toHaveLength(3);
    for (const row of rows) {
      expect(row.className).toContain('hover:translate-x-1.5');
      expect(row.className).toContain('--motion-duration-micro');
      // Nunca mueve el maquetado con márgenes/padding en el hover.
      expect(row.className).not.toMatch(/hover:m-|hover:p-/);
    }
  });

  it('los enlaces tienen objetivo táctil ≥ 44 px y usan tokens', () => {
    const { container } = renderSection();

    expect(screen.getByRole('link', { name: 'hola@ejemplo.com' }).className).toContain('min-h-11');
    expect((container.querySelector('section') as HTMLElement).className).toContain('py-section');
    expect(container.querySelector('.bg-surface-raised')).not.toBeNull();
  });
});
