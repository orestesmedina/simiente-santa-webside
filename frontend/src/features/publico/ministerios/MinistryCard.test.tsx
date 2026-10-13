import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { API_BASE_URL } from '../../../api/client';
import type { MinisterioPublico } from '../../../api/ministerios';
import { LanguageProvider } from '../i18n/LanguageProvider';
import { MinistryCard } from './MinistryCard';

/**
 * Ficha pública (FR-002/FR-024/FR-026, ux.md §4.1): cada elemento existe **solo
 * si tiene dato** —cero huecos, cero líneas vacías, cero botones rotos
 * (SC-017)— y el horario/lugar son **texto libre**, nunca una fecha (FR-024).
 */

function makeMinistry(overrides: Partial<MinisterioPublico> = {}): MinisterioPublico {
  return {
    id: 'min-1',
    name: 'Alabanza',
    description: 'Llevamos la adoración en los cultos.',
    ...overrides,
  };
}

function renderCard(ministry: MinisterioPublico) {
  return render(
    <LanguageProvider>
      <ul>
        <MinistryCard ministry={ministry} />
      </ul>
    </LanguageProvider>,
  );
}

describe('MinistryCard · ficha completa (FR-002)', () => {
  it('pinta imagen con alt, nombre, descripción completa, horario, lugar, encargado y botón', () => {
    renderCard(
      makeMinistry({
        name: 'Alabanza',
        description: 'Línea uno\nLínea dos',
        schedule: 'Todos los viernes, 7:00 p. m.',
        place: 'Templo central',
        leader: { name: 'Josué Ramírez', phone: '+504 9999-9999' },
        whatsappUrl: 'https://chat.whatsapp.com/ABC123',
        image: { url: '/api/v1/media/img_alabanza.png', alt: 'Coro cantando' },
      }),
    );

    const card = screen.getByRole('listitem');
    const image = within(card).getByRole('img', { name: 'Coro cantando' });
    // Las URLs relativas se resuelven contra la base de la API (patrón F3/F4).
    expect(image).toHaveAttribute('src', `${API_BASE_URL}/api/v1/media/img_alabanza.png`);

    expect(within(card).getByRole('heading', { level: 3, name: 'Alabanza' })).toBeVisible();
    // Descripción completa: sin «Ver más» y con los saltos preservados.
    expect(within(card).getByText(/Línea uno\s+Línea dos/)).toBeVisible();
    expect(within(card).queryByText(/Ver más/)).toBeNull();

    expect(within(card).getByText('Todos los viernes, 7:00 p. m.')).toBeVisible();
    expect(within(card).getByText('Templo central')).toBeVisible();
    expect(within(card).getByText('Josué Ramírez')).toBeVisible();
    expect(within(card).getByText('+504 9999-9999')).toBeVisible();
    expect(
      within(card).getByRole('link', { name: 'Unirse al grupo de WhatsApp de Alabanza' }),
    ).toHaveAttribute('href', 'https://chat.whatsapp.com/ABC123');
  });

  it('el horario es texto libre: sin `<time>`, sin fecha calculada y con el valor tal cual', () => {
    const { container } = renderCard(
      makeMinistry({ schedule: 'Una vez al mes, fecha variable', place: 'Salón 2' }),
    );

    expect(screen.getByText('Una vez al mes, fecha variable')).toBeVisible();
    expect(container.querySelector('time')).toBeNull();
  });
});

describe('MinistryCard · casos «sin dato» (SC-017, 0 huecos)', () => {
  it('solo nombre y descripción: no hay imagen, ni líneas, ni botón, ni «—»', () => {
    const { container } = renderCard(
      makeMinistry({ name: 'Matrimonios', description: 'Acompañamos a las parejas.' }),
    );

    const card = screen.getByRole('listitem');
    expect(within(card).getByRole('heading', { level: 3, name: 'Matrimonios' })).toBeVisible();
    expect(within(card).getByText('Acompañamos a las parejas.')).toBeVisible();
    expect(container.querySelector('img')).toBeNull();
    expect(container.querySelector('a')).toBeNull();
    expect(screen.queryByText('—')).toBeNull();
    // Los dos únicos elementos de texto son el nombre y la descripción.
    expect(container.querySelectorAll('p')).toHaveLength(1);
  });

  it('sin enlace no hay botón (ni vacío ni deshabilitado)', () => {
    renderCard(makeMinistry({ schedule: 'Domingos, 9:00 a. m.' }));

    expect(screen.queryByRole('link')).toBeNull();
    expect(screen.queryByRole('button')).toBeNull();
  });

  it('sin `leader` en el DTO no se dibuja ninguna línea de encargado (SC-014)', () => {
    const { container } = renderCard(makeMinistry());

    // El servidor omite la clave con la ficha oculta: aquí solo hay nombre y
    // descripción, sin línea de encargado ni hueco reservado.
    expect(container.querySelectorAll('p')).toHaveLength(1);
    expect(screen.queryByText(/Encargado/)).toBeNull();
  });

  it('imagen rota: se retira sin quebrar la ficha (FR-020)', () => {
    const { container } = renderCard(
      makeMinistry({ image: { url: '/api/v1/media/img_rota.png', alt: 'Imagen del ministerio' } }),
    );

    const image = container.querySelector('img');
    expect(image).not.toBeNull();
    fireEvent.error(image as HTMLImageElement);

    expect(container.querySelector('img')).toBeNull();
    expect(screen.getByRole('heading', { level: 3, name: 'Alabanza' })).toBeVisible();
  });
});
