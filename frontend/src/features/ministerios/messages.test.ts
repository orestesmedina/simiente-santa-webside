import { describe, expect, it } from 'vitest';
import * as messages from './messages';

/**
 * Catálogo de textos del panel de ministerios (ux.md §7.1, FR-018): el panel va
 * **siempre en español** y no pasa por el diccionario del sitio público. Esta
 * prueba fija los textos y garantiza que ninguna constante sea una clave i18n.
 */

const strings = messages;

/** Claves tipo i18n (`dominio.subclave…`), que el panel nunca debe usar. */
const I18N_KEY_PATTERN = /^[a-z][a-zA-Z0-9]*(\.[a-zA-Z0-9]+)+$/;

describe('messages del panel (FR-018)', () => {
  it('ninguna constante exportada es una clave i18n (el panel no se traduce)', () => {
    for (const [name, value] of Object.entries(strings)) {
      if (typeof value !== 'string') {
        continue;
      }
      expect(value, `${name} no debe ser una clave i18n`).not.toMatch(I18N_KEY_PATTERN);
    }
  });

  it('los textos de guardado y publicación son los de ux.md §7.1', () => {
    expect(messages.SAVED).toBe('Cambios guardados.');
    expect(messages.SAVED_PUBLISHED).toBe('Cambios guardados: ya visibles en la sección.');
    expect(messages.CREATED).toBe('Ministerio creado.');
    expect(messages.DRAFT_NOTICE).toBe(
      'Quedó en borrador: no lo verá el público hasta que lo publiques.',
    );
    expect(messages.PUBLISHED).toBe('Publicado en "Ministerios".');
    expect(messages.UNPUBLISHED).toBe(
      'Retirado de la sección. Estará disponible como borrador para publicarlo cuando quieras.',
    );
    expect(messages.SAVED_LEADER_HIDDEN).toBe(
      'Cambios guardados. La ficha del encargado quedó oculta: el público no verá su nombre ni su contacto.',
    );
    expect(messages.SAVED_LEADER_VISIBLE).toBe(
      'Cambios guardados. La ficha del encargado está publicada.',
    );
    expect(messages.DELETED).toBe('Eliminado.');
  });

  it('las confirmaciones interpolan el nombre', () => {
    expect(messages.retireConfirmDescription('Alabanza')).toBe(
      '"Alabanza" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.',
    );
    expect(messages.deleteConfirmDescription('Alabanza')).toBe(
      '"Alabanza" se eliminará para siempre. Esto no se puede deshacer.',
    );
  });

  it('los textos del interruptor y las leyendas de privacidad son los del §4.5/§7.1', () => {
    expect(messages.SWITCH_LABEL).toBe('Mostrar la ficha del encargado en el sitio público');
    expect(messages.LEGEND_LEADER_VISIBLE).toBe(
      'Se publicará: el nombre y (si existe) el contacto del encargado aparecerán en la página de Ministerios.',
    );
    expect(messages.LEGEND_LEADER_HIDDEN).toBe(
      'La ficha del encargado quedará oculta: el público no verá ni su nombre ni su contacto, en ningún lugar del sitio. Tus datos quedan guardados aquí; puedes activar el interruptor otra vez cuando quieras publicarlos.',
    );
    expect(messages.SWITCH_HELP).toBe(
      'La ficha del encargado está oculta hasta que actives este interruptor. Empieza apagado para que seas tú quien decida cuándo hacer pública su información.',
    );
    expect(messages.LEGEND_NEEDS_NAME).toBe(
      'Para publicar la ficha, escribe primero el nombre del encargado.',
    );
  });

  it('la sección sin publicados explica la redirección de `/ministerios` (D-14)', () => {
    expect(messages.NOTHING_PUBLISHED).toBe(
      'Nada está publicado ahora: el público no ve esta sección ni el bloque de la portada, y `/ministerios` redirige a la portada, hasta que publiques al menos un ministerio.',
    );
  });

  it('los límites son espejo de las constantes del contrato (FR-019)', () => {
    expect(messages.NAME_MAX).toBe(160);
    expect(messages.DESCRIPTION_MAX).toBe(2000);
    expect(messages.SCHEDULE_MAX).toBe(200);
    expect(messages.PLACE_MAX).toBe(200);
    expect(messages.LEADER_NAME_MAX).toBe(160);
    expect(messages.PHONE_MAX).toBe(32);
    expect(messages.EMAIL_MAX).toBe(254);
    expect(messages.WHATSAPP_URL_MAX).toBe(500);
    expect(messages.IMAGE_ALT_MAX).toBe(300);
  });

  it('los helpers de validación componen el texto correcto', () => {
    expect(messages.textTooLong(160)).toBe('Ese texto es demasiado largo (máximo 160).');
    expect(messages.descriptionLimitMessage(2005)).toBe(
      '2005 de 2000 caracteres. Recorta 5 para poder guardar.',
    );
    expect(messages.counter(12, 160)).toBe('12/160');
    expect(messages.spanishRequiredFor('la descripción')).toBe(
      'El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de la descripción.',
    );
    expect(messages.emptyModuleMessage()).toBe(
      'Todavía no hay ministerios. Agrega el primero para que el público lo vea.',
    );
  });
});
