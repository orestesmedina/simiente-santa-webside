import { createContext, useContext } from 'react';
import type { Language, PublicMessageKey } from './messages';

/** Valores que `t()` puede interpolar en las plantillas `{nombre}`. */
export type MessageVars = Record<string, string | number>;

export interface LanguageContextValue {
  /** Idioma activo de la interfaz y parámetro `?lang=` de la API (R3-2). */
  lang: Language;
  /** Cambia el idioma, lo persiste en `localStorage` y actualiza `<html lang>`. */
  setLang: (lang: Language) => void;
  /** Resuelve una cadena de interfaz del diccionario activo. */
  t: (key: PublicMessageKey, vars?: MessageVars) => string;
}

/** Contexto de idioma del sitio público (lo provee `LanguageProvider`). */
export const LanguageContext = createContext<LanguageContextValue | null>(null);

/** Acceso al idioma de la interfaz. Debe usarse bajo `<LanguageProvider>`. */
export function useLanguage(): LanguageContextValue {
  const context = useContext(LanguageContext);
  if (!context) {
    throw new Error('useLanguage debe usarse dentro de <LanguageProvider>');
  }
  return context;
}
