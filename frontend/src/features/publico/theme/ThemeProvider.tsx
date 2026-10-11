import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';
import {
  THEME_STORAGE_KEY,
  ThemeContext,
  type Theme,
  type ThemeContextValue,
  type ThemePreference,
} from './useTheme';

const DARK_QUERY = '(prefers-color-scheme: dark)';

function isTheme(value: unknown): value is Theme {
  return value === 'light' || value === 'dark';
}

/**
 * Lee la elección guardada (FR-023). Sin elección → `null` (manda el sistema).
 * Si el almacenamiento no está disponible (navegación privada), se sigue la
 * preferencia del sistema sin avisos, igual que el `LanguageProvider` de F3.
 */
function readStoredTheme(): ThemePreference {
  try {
    const stored = window.localStorage.getItem(THEME_STORAGE_KEY);
    return isTheme(stored) ? stored : null;
  } catch {
    return null;
  }
}

/** Preferencia del sistema; degrada a claro si `matchMedia` no está. */
function readSystemTheme(): Theme {
  const media = window.matchMedia?.(DARK_QUERY);
  return media?.matches ? 'dark' : 'light';
}

function persistTheme(theme: Theme): void {
  try {
    window.localStorage.setItem(THEME_STORAGE_KEY, theme);
  } catch {
    // Si falla, la elección no persiste entre visitas; sin avisos.
  }
}

/**
 * Tema claro/oscuro del sitio público (P10-3/R10-3/FR-023). Sigue el patrón del
 * `LanguageProvider` de F3: estado del **tema efectivo** y de la **preferencia
 * guardada**; aplica `data-theme` y `color-scheme` en `<html>` (lo que ya fijó el
 * script anti-FOUC de `index.html` antes del primer pintado) y respeta el cambio
 * del sistema **en caliente** mientras no haya elección manual. El panel no se
 * envuelve (P10-11): el atributo es global pero el efecto es solo del sitio
 * público, cuyos componentes usan los tokens semánticos.
 */
export function ThemeProvider({ children }: { children: ReactNode }) {
  const [preference, setPreference] = useState<ThemePreference>(readStoredTheme);
  const [systemTheme, setSystemTheme] = useState<Theme>(readSystemTheme);

  const theme: Theme = preference ?? systemTheme;

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    document.documentElement.style.setProperty('color-scheme', theme);
  }, [theme]);

  useEffect(() => {
    if (preference) {
      return;
    }
    const media = window.matchMedia?.(DARK_QUERY);
    if (!media?.addEventListener) {
      return;
    }
    const onChange = (event: MediaQueryListEvent) => {
      setSystemTheme(event.matches ? 'dark' : 'light');
    };
    media.addEventListener('change', onChange);
    return () => media.removeEventListener('change', onChange);
  }, [preference]);

  const setTheme = useCallback((next: Theme) => {
    setPreference(next);
    persistTheme(next);
  }, []);

  const value = useMemo<ThemeContextValue>(
    () => ({ theme, preference, setTheme }),
    [theme, preference, setTheme],
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}
