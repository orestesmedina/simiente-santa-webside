import { render, screen, within } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { ItemMinisterioPortada } from '../../../api/ministerios';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { MinistryItem } from './MinistryItem';

/**
 * Tarjeta compacta del bloque «Ministerios» de la portada (US6/D-11/D-12, ux.md
 * §4.9): imagen opcional, chip neutro «Ministerio», **título que es el enlace**
 * a `/ministerios` y líneas de horario/lugar solo con dato. **Sin encargado y
 * sin botón de unirse** (el CTA del bloque es el único).
 */

function makeItem(overrides: Partial<ItemMinisterioPortada> = {}): ItemMinisterioPortada {
  return { id: 'min-1', name: 'Alabanza', ...overrides };
}

function renderItem(item: ItemMinisterioPortada) {
  return render(
    <LanguageProvider>
      <MemoryRouter initialEntries={['/']}>
        <ul>
          <MinistryItem item={item} />
        </ul>
      </MemoryRouter>
    </LanguageProvider>,
  );
}

describe('MinistryItem · tarjeta compacta (D-11/D-12)', () => {
  it('imagen con alt, chip «Ministerio», título-enlace y líneas de horario y lugar', () => {
    renderItem(
      makeItem({
        schedule: 'Todos los viernes, 7:00 p. m.',
        place: 'Templo central',
        image: { url: '/api/v1/media/img_alabanza.png', alt: 'Coro cantando' },
      }),
    );

    const item = screen.getByRole('listitem');
    expect(within(item).getByRole('img', { name: 'Coro cantando' })).toHaveAttribute(
      'src',
      `${API_BASE_URL}/api/v1/media/img_alabanza.png`,
    );
    expect(within(item).getByText('Ministerio')).toBeVisible();

    const enlace = within(item).getByRole('link', { name: 'Ministerio: Alabanza' });
    expect(enlace).toHaveAttribute('href', '/ministerios');
    expect(within(item).getByRole('heading', { level: 3 })).toContainElement(enlace);

    // Horario y lugar como texto truncado con el valor completo en `title`.
    expect(within(item).getByText('Todos los viernes, 7:00 p. m.')).toHaveAttribute(
      'title',
      'Todos los viernes, 7:00 p. m.',
    );
    expect(within(item).getByText('Templo central')).toBeVisible();
  });

  it('sin encargado y sin botón de unirse: el CTA del bloque es el único (D-12)', () => {
    const { container } = renderItem(
      makeItem({ schedule: 'Domingos, 9:00 a. m.', place: 'Salón 2' }),
    );

    // El único enlace es el título; no hay botón de WhatsApp ni línea de encargado.
    expect(screen.getAllByRole('link')).toHaveLength(1);
    expect(container.querySelectorAll('a')).toHaveLength(1);
    expect(screen.queryByText(/Encargado|Unirse/)).toBeNull();
    // Sin `<time>`: el horario es texto libre (FR-024/SC-015).
    expect(container.querySelector('time')).toBeNull();
  });

  it('solo nombre: sin imagen, sin líneas vacías y sin huecos', () => {
    const { container } = renderItem(makeItem({ name: 'Matrimonios' }));

    const item = screen.getByRole('listitem');
    expect(within(item).getByRole('link', { name: 'Ministerio: Matrimonios' })).toBeVisible();
    expect(container.querySelector('img')).toBeNull();
    expect(container.querySelectorAll('p')).toHaveLength(0);
    expect(container.textContent).not.toContain('—');
  });

  it('en inglés el chip y el nombre accesible se traducen (SC-006)', () => {
    window.localStorage.setItem('ss.lang', 'en');
    renderItem(makeItem({ name: 'Worship' }));

    expect(screen.getByText('Ministry')).toBeVisible();
    expect(screen.getByRole('link', { name: 'Ministry: Worship' })).toBeVisible();
    window.localStorage.clear();
  });
});
