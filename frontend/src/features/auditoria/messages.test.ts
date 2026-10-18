import { describe, expect, it } from 'vitest';
import { ACTION_LABELS, actionLabel } from './messages';

/** Los 15 códigos de auditoría de F4 (migración `000008`, R4-10). */
const F4_ACTIONS = [
  'event.create',
  'event.update',
  'event.delete',
  'event.publish',
  'event.unpublish',
  'event.image.upload',
  'activity.create',
  'activity.update',
  'activity.delete',
  'activity.publish',
  'activity.unpublish',
  'activity.image.upload',
  'category.create',
  'category.update',
  'category.delete',
] as const;

describe('ACTION_LABELS · etiquetas humanas de la auditoría (ux.md §4.9)', () => {
  it('cubre los 15 códigos de F4 sin mostrar códigos crudos', () => {
    for (const action of F4_ACTIONS) {
      const label = ACTION_LABELS[action];
      expect(label, `etiqueta de ${action}`).toBeTruthy();
      expect(label).not.toBe(action);
      expect(actionLabel(action)).toBe(label);
    }
  });

  it('distingue el concepto y el verbo de cada código', () => {
    expect(ACTION_LABELS['event.create']).toBe('Creó un evento');
    expect(ACTION_LABELS['event.update']).toBe('Modificó un evento');
    expect(ACTION_LABELS['event.publish']).toBe('Publicó un evento');
    expect(ACTION_LABELS['event.unpublish']).toBe('Retiró un evento');
    expect(ACTION_LABELS['event.delete']).toBe('Eliminó un evento');
    expect(ACTION_LABELS['event.image.upload']).toBe('Subió la imagen de un evento');
    expect(ACTION_LABELS['activity.create']).toBe('Creó una actividad');
    expect(ACTION_LABELS['activity.publish']).toBe('Publicó una actividad');
    expect(ACTION_LABELS['category.create']).toBe('Creó una categoría');
    expect(ACTION_LABELS['category.delete']).toBe('Eliminó una categoría');
  });

  it('conserva intactas las etiquetas de F2', () => {
    expect(ACTION_LABELS['user.create']).toBe('Creó una cuenta');
    expect(ACTION_LABELS['role.delete']).toBe('Eliminó un rol');
  });

  it('cubre los 8 códigos `ministry.*` de F6 (migración `000011`, R6-12)', () => {
    const F6_ACTIONS = [
      'ministry.create',
      'ministry.update',
      'ministry.delete',
      'ministry.publish',
      'ministry.unpublish',
      'ministry.leader.show',
      'ministry.leader.hide',
      'ministry.image.upload',
    ] as const;

    for (const action of F6_ACTIONS) {
      const label = ACTION_LABELS[action];
      expect(label, `etiqueta de ${action}`).toBeTruthy();
      expect(label).not.toBe(action);
      expect(actionLabel(action)).toBe(label);
    }
  });

  it('distingue el concepto y el verbo de cada código de ministerios (ux.md §4.8/§7.1)', () => {
    expect(ACTION_LABELS['ministry.create']).toBe('Creó el ministerio');
    expect(ACTION_LABELS['ministry.update']).toBe('Modificó el ministerio');
    expect(ACTION_LABELS['ministry.publish']).toBe('Publicó el ministerio');
    expect(ACTION_LABELS['ministry.unpublish']).toBe('Retiró el ministerio');
    expect(ACTION_LABELS['ministry.delete']).toBe('Eliminó el ministerio');
    expect(ACTION_LABELS['ministry.leader.show']).toBe(
      'Publicó la ficha del encargado del ministerio',
    );
    expect(ACTION_LABELS['ministry.leader.hide']).toBe(
      'Ocultó la ficha del encargado del ministerio',
    );
    expect(ACTION_LABELS['ministry.image.upload']).toBe('Subió la imagen de un ministerio');
  });

  it('cubre los 6 códigos `media.*` de F9 (migración `000015`, P9-14)', () => {
    const F9_ACTIONS = [
      'media.sync',
      'media.description.update',
      'media.hide',
      'media.show',
      'media.merge',
      'media.mark.resolve',
    ] as const;

    for (const action of F9_ACTIONS) {
      const label = ACTION_LABELS[action];
      expect(label, `etiqueta de ${action}`).toBeTruthy();
      expect(label).not.toBe(action);
      expect(actionLabel(action)).toBe(label);
    }

    expect(ACTION_LABELS['media.sync']).toBe('Buscó episodios nuevos');
    expect(ACTION_LABELS['media.description.update']).toBe(
      'Modificó la descripción de un episodio',
    );
    expect(ACTION_LABELS['media.hide']).toBe('Ocultó un episodio');
    expect(ACTION_LABELS['media.show']).toBe('Mostró un episodio');
    expect(ACTION_LABELS['media.merge']).toBe('Unió dos entradas en una sola ficha');
    expect(ACTION_LABELS['media.mark.resolve']).toBe('Resolvió la marca de un episodio');
  });

  it('un código desconocido se muestra tal cual (nunca en blanco)', () => {
    expect(actionLabel('home.update')).toBe('home.update');
  });

  it('cubre los 8 códigos `donation.*` de F7 (migración `000013`, R7-10)', () => {
    const F7_ACTIONS = [
      'donation.create',
      'donation.update',
      'donation.delete',
      'donation.publish',
      'donation.unpublish',
      'donation.text.update',
      'donation.text.publish',
      'donation.text.unpublish',
    ] as const;

    for (const action of F7_ACTIONS) {
      const label = ACTION_LABELS[action];
      expect(label, `etiqueta de ${action}`).toBeTruthy();
      expect(label).not.toBe(action);
      expect(actionLabel(action)).toBe(label);
    }

    expect(ACTION_LABELS['donation.create']).toBe('Creó una cuenta o número de donaciones');
    expect(ACTION_LABELS['donation.publish']).toBe('Publicó una cuenta o número de donaciones');
    expect(ACTION_LABELS['donation.text.update']).toBe(
      'Modificó el texto introductorio o el de uso',
    );
    expect(ACTION_LABELS['donation.text.unpublish']).toBe(
      'Retiró el texto introductorio o el de uso',
    );
  });
});
