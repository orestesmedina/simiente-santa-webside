const LOCALE = 'es-ES';

const DATE_FORMAT = new Intl.DateTimeFormat(LOCALE, {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
});

const TIME_FORMAT = new Intl.DateTimeFormat(LOCALE, {
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
});

function toDate(value: string | Date): Date {
  return value instanceof Date ? value : new Date(value);
}

/** Fecha en formato español: "03/10/2026". Zona horaria del navegador. */
export function formatDate(value: string | Date): string {
  return DATE_FORMAT.format(toDate(value));
}

/** Hora en formato español de 24 h: "17:42". */
export function formatTime(value: string | Date): string {
  return TIME_FORMAT.format(toDate(value));
}

/** Fecha y hora: "03/10/2026, 17:42" (historiales de auditoría). */
export function formatDateTime(value: string | Date): string {
  return `${formatDate(value)}, ${formatTime(value)}`;
}

/**
 * Último acceso exitoso de una cuenta (FR-021, ux.md §3.4): "03/10/2026, a las
 * 17:42, desde 189.2.4.15". Si la cuenta nunca ha entrado, devuelve el texto
 * acordado sin inventar ninguna fecha (US8 esc. 6).
 */
export function formatLastLogin(at: string | null | undefined, ip?: string | null): string {
  if (!at) {
    return 'Nunca ha entrado al panel.';
  }
  const base = `${formatDate(at)}, a las ${formatTime(at)}`;
  return ip ? `${base}, desde ${ip}` : base;
}
