/** Criterio telefónico de F2: separadores habituales, `+` opcional, ≥7 dígitos. */
export function isValidPhone(value: string): boolean {
  if (!/^\+?[0-9\s().-]+$/.test(value)) {
    return false;
  }
  const digits = value.replace(/\D/g, '');
  return digits.length >= 7 && value.length <= 32;
}

/** Formato de correo simple (la autoridad final es el servidor). */
export function isValidEmail(value: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
}

/** URL `https` con host no vacío (espejo de la etiqueta `url` del servidor). */
export function isHttpsUrl(value: string): boolean {
  try {
    const url = new URL(value);
    return url.protocol === 'https:' && url.host !== '';
  } catch {
    return false;
  }
}

/** Enlace de grupo de WhatsApp: `https` de `chat.whatsapp.com` o `wa.me`. */
export function isWhatsappGroupUrl(value: string): boolean {
  if (!isHttpsUrl(value)) {
    return false;
  }
  const host = new URL(value).host;
  return host === 'chat.whatsapp.com' || host === 'wa.me';
}
