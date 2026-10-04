/**
 * Estados de la vista de estado del sistema (ux.md §4). La respuesta de
 * `/healthz` se mapea SIEMPRE a uno de estos, mediante `toEstadoSistema`.
 * El `motivo` de `inaccesible` es interno: guía pruebas y diagnóstico.
 */
export type EstadoSistema =
  | { kind: 'conectado' }
  | { kind: 'bd-no-conectada' }
  | { kind: 'inaccesible'; motivo: 'sin-respuesta' | 'respuesta-inesperada' };

/**
 * Palabra de estado de una fila del detalle de verificación (ux.md §2 y §6).
 * `servidor-ok` → "en marcha"; `bd-ok` → "conectada"; `bd-error` → "no conectada".
 */
export type VerificationValue =
  'servidor-ok' | 'bd-ok' | 'bd-error' | 'sin-respuesta' | 'no-comprobable';
