import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';
import { en } from './en';
import { es } from './es';
import {
  LANGUAGE_STORAGE_KEY,
  type Language,
  type PublicMessageKey,
  type PublicMessages,
} from './messages';
import { LanguageContext, type LanguageContextValue, type MessageVars } from './useLanguage';

const DICTIONARIES: Record<Language, PublicMessages> = { es, en };

function isLanguage(value: unknown): value is Language {
  return value === 'es' || value === 'en';
}

/**
 * Lee la preferencia guardada. Solo la primera visita sin preferencia entra en
 * español; una re-visita la respeta (FR-010/analyze C3). Nunca se auto-detecta
 * el idioma del navegador (`navigator.language`). Si el almacenamiento no está
 * disponible (p. ej. navegación privada) se sigue en español sin avisos.
 */
function readStoredLanguage(): Language {
  try {
    const stored = window.localStorage.getItem(LANGUAGE_STORAGE_KEY);
    if (isLanguage(stored)) {
      return stored;
    }
  } catch {
    // Almacenamiento no disponible: se usa el idioma base.
  }
  return 'es';
}

function persistLanguage(lang: Language): void {
  try {
    window.localStorage.setItem(LANGUAGE_STORAGE_KEY, lang);
  } catch {
    // Si falla, el idioma simplemente no persiste entre visitas; sin avisos.
  }
}

function interpolate(template: string, vars?: MessageVars): string {
  if (!vars) {
    return template;
  }
  return template.replace(/\{(\w+)\}/g, (match, name: string) =>
    name in vars ? String(vars[name]) : match,
  );
}

/**
 * Contexto de idioma del sitio público (R3-9/P3-10). Sin dependencias externas:
 * diccionarios tipados en compilación y memoria en `localStorage['ss.lang']`.
 * Lo reutilizan F4–F9.
 */
export function LanguageProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Language>(readStoredLanguage);

  useEffect(() => {
    document.documentElement.lang = lang;
  }, [lang]);

  const setLang = useCallback((next: Language) => {
    setLangState(next);
    persistLanguage(next);
  }, []);

  const t = useCallback(
    (key: PublicMessageKey, vars?: MessageVars) => interpolate(DICTIONARIES[lang][key], vars),
    [lang],
  );

  const value = useMemo<LanguageContextValue>(() => ({ lang, setLang, t }), [lang, setLang, t]);

  return <LanguageContext.Provider value={value}>{children}</LanguageContext.Provider>;
}
