import { createContext, useContext } from 'react';

/** Tema efectivo aplicado al sitio público. */
export type Theme = 'light' | 'dark';

/** Elección guardada del visitante; `null` = seguir la preferencia del sistema. */
export type ThemePreference = Theme | null;

export interface ThemeContextValue {
  /** Tema efectivo en pantalla (`preference` si la hay, si no el del sistema). */
  theme: Theme;
  /** Elección manual guardada (`null` si nunca se eligió). */
  preference: ThemePreference;
  /** Aplica y persiste una elección manual. */
  setTheme: (theme: Theme) => void;
}

/** Clave de `localStorage` que recuerda la elección entre visitas (P10-3). */
export const THEME_STORAGE_KEY = 'ss.theme';

export const ThemeContext = createContext<ThemeContextValue | null>(null);

/**
 * Acceso al tema efectivo, a la preferencia guardada y a `setTheme` (FR-023).
 * Es la única fuente de verdad del tema del sitio público.
 */
export function useTheme(): ThemeContextValue {
  const value = useContext(ThemeContext);
  if (!value) {
    throw new Error('useTheme debe usarse dentro de <ThemeProvider>');
  }
  return value;
}
