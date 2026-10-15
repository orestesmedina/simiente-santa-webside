import { describe, expect, it } from 'vitest';
import * as messages from './messages';

/**
 * Catálogo de textos del panel de donaciones (ux.md §7.1, FR-018): el panel va
 * **siempre en español** y no pasa por el diccionario del sitio público. Esta
 * prueba fija los textos y garantiza que ninguna constante sea una clave i18n.
 */

/** Claves tipo i18n (`dominio.subclave…`), que el panel nunca debe usar. */
const I18N_KEY_PATTERN = /^[a-z][a-zA-Z0-9]*(\.[a-zA-Z0-9]+)+$/;

describe('messages del panel de donaciones (FR-018)', () => {
  it('ninguna constante exportada es una clave i18n (el panel no se traduce)', () => {
    for (const [name, value] of Object.entries(messages)) {
      if (typeof value !== 'string') {
        continue;
      }
      expect(value, `${name} no debe ser una clave i18n`).not.toMatch(I18N_KEY_PATTERN);
    }
  });

  it('el encabezado y la banda de ayuda son los de ux.md §4.3/§7.1', () => {
    expect(messages.TITLE).toBe('Gestión de donaciones');
    expect(messages.HELP).toBe(
      'Aquí administras las cuentas y números donde la iglesia recibe ofrendas, y los textos que explican cómo donar y en qué se usan. Solo lo publicado es visible al público.',
    );
  });

  it('los textos de guardado y publicación son los de ux.md §7.1', () => {
    expect(messages.CREATED_ACCOUNT).toBe('Cuenta creada.');
    expect(messages.CREATED_SINPE).toBe('Número de SINPE creado.');
    expect(messages.DRAFT_NOTICE_ACCOUNT).toBe(
      'Quedó en borrador: no la verá el público hasta que la publiques.',
    );
    expect(messages.DRAFT_NOTICE_SINPE).toBe(
      'Quedó en borrador: no lo verá el público hasta que lo publiques.',
    );
    expect(messages.SAVED).toBe('Cambios guardados.');
    expect(messages.SAVED_PUBLISHED).toBe('Cambios guardados: ya visibles en la sección.');
    expect(messages.PUBLISHED).toBe('Publicado en "Ofrendas y diezmos".');
    expect(messages.UNPUBLISHED).toBe(
      'Retirado de la sección. Estará disponible como borrador para publicarlo cuando quieras.',
    );
    expect(messages.DELETED).toBe('Eliminado.');
  });

  it('la sección sin publicados explica la redirección de `/ofrendas` (Q9/P7-13)', () => {
    expect(messages.NOTHING_PUBLISHED).toBe(
      'Nada está publicado ahora: el público no ve esta sección ni el bloque de la portada, y `/ofrendas` redirige a la portada, hasta que publiques al menos un método (cuenta o número) o un texto.',
    );
  });

  it('las confirmaciones interpolan el dato del método (ux.md §4.3/§7.1)', () => {
    expect(messages.retireConfirmDescription('Banco Nacional de Costa Rica')).toBe(
      '"Banco Nacional de Costa Rica" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.',
    );
    expect(
      messages.deleteConfirmDescription('Cuenta IBAN', 'Banco Nacional', 'CR05015202001234567890'),
    ).toBe(
      'Cuenta IBAN Banco Nacional CR05015202001234567890 se eliminará para siempre. Esto no se puede deshacer.',
    );
  });

  it('los límites son espejo de las constantes del contrato (FR-014/FR-019)', () => {
    expect(messages.INSTITUTION_MAX).toBe(160);
    expect(messages.HOLDER_MAX).toBe(160);
    expect(messages.TEXT_MAX).toBe(2000);
    expect(messages.NUMBER_INPUT_MAX).toBe(60);
  });

  it('los helpers de validación componen el texto correcto', () => {
    expect(messages.textTooLong(160)).toBe('Ese texto es demasiado largo (máximo 160).');
    expect(messages.textLimitMessage(2005)).toBe(
      '2005 de 2000 caracteres. Recorta 5 para poder guardar.',
    );
    expect(messages.counter(12, 2000)).toBe('12/2000');
    expect(messages.spanishRequiredFor('el titular')).toBe(
      'El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de el titular.',
    );
    expect(messages.emptyGroupMessage('iban')).toBe(
      'Todavía no hay cuentas IBAN. Agrega la primera para que el público la vea.',
    );
    expect(messages.emptyGroupMessage('sinpe')).toBe(
      'Todavía no hay números de SINPE Móvil. Agrega el primero para que el público lo vea.',
    );
  });

  it('los textos de error y validación de ux.md §2.5/§7.1 están presentes', () => {
    expect(messages.REQUIRED_INSTITUTION).toBe('Escribe el nombre del banco.');
    expect(messages.REQUIRED_NUMBER).toBe(
      'Escribe el número de la cuenta (o el número de celular para SINPE Móvil).',
    );
    expect(messages.IBAN_INVALID).toBe(
      'Ese número no parece una cuenta IBAN válida. Revisa la cantidad de letras y dígitos: vale tanto el IBAN de Costa Rica como uno internacional.',
    );
    expect(messages.SINPE_INVALID).toBe(
      'Ese número no parece un celular de Costa Rica: son 8 dígitos, con o sin el prefijo +506.',
    );
    expect(messages.REQUIRED_CURRENCY).toBe('Elige la moneda de la cuenta: colones o dólares.');
    expect(messages.DUPLICATE).toBe(
      'Ya existe esa misma cuenta o número (mismo tipo, entidad y número). Edita el que ya tienes o cambia algún dato.',
    );
    expect(messages.TEXT_NOT_WRITTEN).toBe('Todavía sin escribir.');
  });
});
