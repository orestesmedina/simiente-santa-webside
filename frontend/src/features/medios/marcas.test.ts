import { describe, expect, it } from 'vitest';
import type { MedioAdmin } from '../../api/medios';
import {
  contarPendientes,
  enlaceDisponible,
  marcaDisponibilidadPendiente,
  marcasPendientes,
  parDuplicadoAbierto,
  proveedorDelEpisodio,
} from './marcas';

/**
 * Reglas de marcas del panel (ux.md §4.7, P9-19): los dos ejes son
 * **independientes** y **conviven**, «pendiente» es «sin resolver» en su eje y
 * el explicador de «ya no aparece en el feed» nombra el servicio de la ficha.
 */

function medio(overrides: Partial<MedioAdmin> = {}): MedioAdmin {
  return {
    id: 'med-1',
    title: 'Potencial sin Resultados La Lección de la Higuera',
    descriptionEs: 'La higuera estéril.',
    descriptionEn: null,
    descriptionEdited: false,
    publishedAt: '2026-05-25T18:43:42Z',
    origin: 'both',
    visibility: 'visible',
    sources: [],
    createdAt: '2026-10-18T20:34:17Z',
    updatedAt: '2026-10-18T20:34:17Z',
    ...overrides,
  };
}

describe('marcas pendientes de una ficha (FR-009/FR-014/FR-017)', () => {
  it('un par de duplicados está pendiente mientras no tenga fecha de resolución', () => {
    expect(parDuplicadoAbierto(medio())).toBe(false);
    expect(parDuplicadoAbierto(medio({ duplicateOfId: 'otro' }))).toBe(true);
    expect(
      parDuplicadoAbierto(
        medio({ duplicateOfId: 'otro', duplicateMarkResolvedAt: '2026-10-18T21:00:00Z' }),
      ),
    ).toBe(false);
  });

  it('una marca de disponibilidad está pendiente mientras no tenga fecha de resolución', () => {
    expect(marcaDisponibilidadPendiente(medio())).toBeUndefined();
    expect(marcaDisponibilidadPendiente(medio({ availabilityMark: 'missing_from_feed' }))).toBe(
      'missing_from_feed',
    );
    expect(
      marcaDisponibilidadPendiente(
        medio({
          availabilityMark: 'unavailable',
          availabilityMarkResolvedAt: '2026-10-18T21:00:00Z',
        }),
      ),
    ).toBeUndefined();
  });

  it('los dos ejes conviven: una fila puede tener dos marcas pendientes a la vez (P9-19)', () => {
    const conDos = medio({
      duplicateOfId: 'otro',
      availabilityMark: 'missing_from_feed',
    });
    expect(marcasPendientes(conDos)).toEqual(['possible_duplicate', 'missing_from_feed']);

    // Resolver un eje no cierra el otro.
    expect(
      marcasPendientes({ ...conDos, duplicateMarkResolvedAt: '2026-10-18T21:00:00Z' }),
    ).toEqual(['missing_from_feed']);
  });

  it('cuenta los pendientes por clase para las pastillas del filtro (⟲ D-10)', () => {
    const items = [
      medio({ id: 'a', duplicateOfId: 'b' }),
      medio({ id: 'b', duplicateOfId: 'a' }),
      medio({ id: 'c', availabilityMark: 'unavailable' }),
      medio({ id: 'd', availabilityMark: 'missing_from_feed' }),
      medio({ id: 'e' }),
    ];

    expect(contarPendientes(items)).toEqual({
      possible_duplicate: 2,
      unavailable: 1,
      missing_from_feed: 1,
    });
  });

  it('el explicador nombra el servicio de la primera fuente de la ficha (FR-017)', () => {
    expect(
      proveedorDelEpisodio(
        medio({
          sources: [
            {
              provider: 'podcast',
              externalId: 'ep-1',
              feedTitle: 'Fe como fruto del Espíritu',
              url: 'https://podcasters.spotify.com/pod/show/x/episodes/y',
            },
          ],
        }),
      ),
    ).toBe('podcast');
    // Sin fuentes manda el formato; y una comprobación con `available` cambia la línea.
    expect(proveedorDelEpisodio(medio({ origin: 'audio' }))).toBe('podcast');
    expect(proveedorDelEpisodio(medio({ origin: 'video' }))).toBe('youtube');
    expect(enlaceDisponible(medio({ availabilityResult: 'available' }))).toBe(true);
    expect(enlaceDisponible(medio({ availabilityResult: 'inconclusive' }))).toBe(false);
  });
});
