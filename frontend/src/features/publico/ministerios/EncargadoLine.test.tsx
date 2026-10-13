import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { LiderPublico } from '../../../api/ministerios';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { EncargadoLine } from './EncargadoLine';

/**
 * Línea del encargado (FR-026/SC-017, ux.md §4.1/§6.3): nombre siempre, y el
 * **único** medio de contacto solo si existe. Sin lógica de visibilidad: el
 * servidor ya no manda el dato cuando la ficha no está publicada.
 */

function renderLine(leader: LiderPublico) {
  return render(
    <LanguageProvider>
      <EncargadoLine leader={leader} />
    </LanguageProvider>,
  );
}

describe('EncargadoLine · ficha del encargado (FR-026)', () => {
  it('muestra el nombre y el celular cuando existe', () => {
    renderLine({ name: 'Josué Ramírez', phone: '+504 9999-9999' });

    expect(screen.getByText('Josué Ramírez')).toBeVisible();
    expect(screen.getByText('+504 9999-9999')).toBeVisible();
    expect(screen.getByText('Encargado:')).toBeInTheDocument();
  });

  it('con ficha publicada y sin contacto muestra solo el nombre (no es error, Q8)', () => {
    const { container } = renderLine({ name: 'Ana Rosales' });

    expect(screen.getByText('Ana Rosales')).toBeVisible();
    // Ninguna línea de contacto vacía ni huérfana.
    expect(container.querySelectorAll('p')).toHaveLength(1);
    expect(screen.queryByText('—')).toBeNull();
  });

  it('el correo es el medio alternativo y nunca aparece junto al celular', () => {
    renderLine({ name: 'María Pérez', email: 'maria@ejemplo.com' });

    expect(screen.getByText('maria@ejemplo.com')).toBeVisible();
    expect(screen.queryByText('+504 9999-9999')).toBeNull();
  });
});
