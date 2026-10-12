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

  it('un código desconocido se muestra tal cual (nunca en blanco)', () => {
    expect(actionLabel('home.update')).toBe('home.update');
  });
});
