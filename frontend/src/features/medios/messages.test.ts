import { describe, expect, it } from 'vitest';
import * as mensajes from './messages';
import {
  DESCRIPTION_MAX,
  FILTER_EMPTY,
  HELP,
  MARK_CLOSED,
  SPANISH_REQUIRED,
  SUMMARY_BOTH_FEEDS_FAILED,
  SUMMARY_NO_NEWS,
  counter,
  descriptionLimitMessage,
  feedLabel,
  isAvailabilityMark,
  markExplainer,
  markLabel,
  providerLabel,
  showAgainConfirm,
  summaryAvailabilityChecked,
  summaryFeedFailed,
  summaryImported,
  summaryImportedWithMarks,
} from './messages';

/**
 * Catálogo del panel de F9 (ux.md §7.1, FR-024): el panel va **siempre en
 * español** —ninguna de estas cadenas pasa por el diccionario del sitio
 * público—, los seis casos del resumen son literales del catálogo y los
 * segmentos con número 0 **no** se mencionan (⟲ D-19).
 */

/** Forma de una clave de i18n («media.page.title»): prohibida en el panel. */
const CLAVE_I18N = /^[a-z][a-z0-9-]*(\.[a-z0-9-]+)+$/;

describe('catálogo del panel de medios (FR-024)', () => {
  it('todas las cadenas son textos en español y no claves de i18n', () => {
    const cadenas = Object.entries(mensajes).filter(
      (entrada) => typeof entrada[1] === 'string',
    ) as Array<[string, string]>;
    expect(cadenas.length).toBeGreaterThan(30);
    for (const [nombre, texto] of cadenas) {
      expect(texto.trim().length, `${nombre} no vacía`).toBeGreaterThan(0);
      expect(CLAVE_I18N.test(texto), `${nombre} no es una clave de i18n`).toBe(false);
    }
  });

  it('la banda de ayuda es la de ux.md §7.1', () => {
    expect(HELP).toBe(
      'Aquí miras y mantienes las grabaciones que el sitio trae desde el canal de YouTube y el feed del podcast. Pulsa “Buscar episodios nuevos” cuando se publique algo nuevo y revisa con calma lo que quede marcado; solo lo visible lo verá el público.',
    );
  });

  it('los servicios se nombran como los llama la iglesia (FR-018/D-21)', () => {
    expect(providerLabel('youtube')).toBe('YouTube');
    expect(providerLabel('podcast')).toBe('el podcast');
    expect(feedLabel('youtube')).toBe('feed de YouTube');
    expect(feedLabel('podcast')).toBe('feed del podcast');
    for (const texto of [HELP, SUMMARY_NO_NEWS, summaryFeedFailed('youtube')]) {
      for (const tecnicismo of ['Atom', 'RSS', 'Anchor', 'feed Atom', 'idempotente']) {
        expect(texto, `«${tecnicismo}» no aparece`).not.toContain(tecnicismo);
      }
    }
  });
});

describe('los seis resúmenes de «Buscar episodios nuevos» (ux.md §4.4/§7.1)', () => {
  it('nuevos sin marcas: literal del catálogo', () => {
    expect(summaryImported(10, 8)).toBe(
      'Se agregaron 10 episodios nuevos; 8 se unieron con su video y su audio. Ya están visibles en “Medios”.',
    );
  });

  it('nuevos con marcas: literal del catálogo', () => {
    expect(summaryImportedWithMarks(10, 8, 2)).toBe(
      'Se agregaron 10 episodios nuevos; 8 se unieron con su video y su audio; 2 quedaron marcados como posible duplicado. Revísalos abajo.',
    );
  });

  it('un segmento con número 0 no se menciona (⟲ D-19)', () => {
    expect(summaryImported(3, 0)).toBe(
      'Se agregaron 3 episodios nuevos. Ya están visibles en “Medios”.',
    );
    expect(summaryImportedWithMarks(3, 0, 0)).toBe(
      'Se agregaron 3 episodios nuevos. Revísalos abajo.',
    );
    expect(summaryImportedWithMarks(3, 0, 1)).not.toContain('se unieron');
  });

  it('sin novedades: mensaje neutro con expectativa (FR-012/⟲ D-20)', () => {
    expect(SUMMARY_NO_NEWS).toBe(
      'No hay episodios nuevos: nada cambió. Cuando la iglesia publique algo más, vuelve a pulsar el botón.',
    );
  });

  it('un feed falló: dice qué feed, que nada se ocultó ni se marcó y que se reintente (FR-016)', () => {
    expect(summaryFeedFailed('youtube')).toBe(
      'El feed de YouTube no respondió. Se agregó lo que trajo el podcast; nada se ocultó ni se marcó por esa causa. Vuelve a pulsar la búsqueda en un rato para recuperar lo que falte.',
    );
    expect(summaryFeedFailed('podcast')).toBe(
      'El feed del podcast no respondió. Se agregó lo que trajo YouTube; nada se ocultó ni se marcó por esa causa. Vuelve a pulsar la búsqueda en un rato para recuperar lo que falte.',
    );
  });

  it('ambos feeds fallaron: todo lo que había está a salvo', () => {
    expect(SUMMARY_BOTH_FEEDS_FAILED).toBe(
      'Ninguno de los dos feeds respondió. No se agregó nada y no se cambió nada: todo lo que tenías está a salvo. Prueba más tarde.',
    );
  });

  it('la comprobación de disponibilidad invita a mirar las marcas (FR-017)', () => {
    expect(summaryAvailabilityChecked(4)).toBe(
      'La búsqueda revisó 4 episodios para ver si sus enlaces siguen funcionando: mira las marcas.',
    );
  });
});

describe('marcas de revisión y sus resoluciones (ux.md §4.7)', () => {
  it('cada marca tiene palabra (nunca solo color, FR-030)', () => {
    expect(markLabel('possible_duplicate')).toBe('Posible duplicado');
    expect(markLabel('unavailable')).toBe('No disponible');
    expect(markLabel('missing_from_feed')).toBe('Ya no aparece en el feed');
  });

  it('cada marca tiene su explicador de una línea', () => {
    expect(markExplainer('possible_duplicate', 'youtube')).toBe(
      'El sistema no pudo estar seguro de si estas dos entradas son el mismo sermón.',
    );
    expect(markExplainer('unavailable', 'youtube')).toBe(
      'Se comprobó: el enlace de esta grabación ya no funciona (borrada o privada). El público no la ve.',
    );
    expect(markExplainer('missing_from_feed', 'youtube', true)).toBe(
      'Salió de la ventana del feed de YouTube (así es: la ventana es corta). Su enlace sigue funcionando y sigue visible para el público.',
    );
    expect(markExplainer('missing_from_feed', 'podcast', false)).toBe(
      'Salió de la ventana del feed del podcast. No se pudo comprobar si su enlace sigue funcionando; mientras tanto, sigue visible para el público.',
    );
  });

  it('los dos ejes de marca se distinguen (P9-19)', () => {
    expect(isAvailabilityMark('unavailable')).toBe(true);
    expect(isAvailabilityMark('missing_from_feed')).toBe(true);
    expect(isAvailabilityMark('possible_duplicate')).toBe(false);
  });

  it('al resolver se dice que la marca no se reabre por la misma causa', () => {
    expect(MARK_CLOSED).toBe('La marca se cerró y no se vuelve a abrir por esta misma causa.');
  });

  it('«Mostrar de nuevo» advierte la consecuencia (⟲ D-28)', () => {
    expect(showAgainConfirm('Potencial sin Resultados')).toBe(
      'Si el enlace de “Potencial sin Resultados” sigue sin funcionar, volverá a quedar oculto la próxima vez que se revise.',
    );
  });

  it('el filtro sin pendientes es un logro, no un hueco (ux.md §4.8)', () => {
    expect(FILTER_EMPTY).toBe('No hay nada pendiente de revisión: todo está al día.');
  });
});

describe('validación de la descripción (FR-023)', () => {
  it('el límite es de 300 caracteres por idioma', () => {
    expect(DESCRIPTION_MAX).toBe(300);
  });

  it('el contador excedido dice cuánto recortar', () => {
    expect(descriptionLimitMessage(312)).toBe(
      '312 de 300 caracteres. Recorta 12 para poder guardar.',
    );
    expect(counter(12)).toBe('12/300');
  });

  it('el inglés sin español no permite guardar', () => {
    expect(SPANISH_REQUIRED).toBe(
      'El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de la descripción corta.',
    );
  });
});
