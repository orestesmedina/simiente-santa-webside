import { useLanguage } from '../i18n/useLanguage';
import { useTheme } from './useTheme';

export interface ThemeToggleProps {
  className?: string;
}

/**
 * Interruptor claro/oscuro de la cabecera del sitio público (F10 · T1028/FR-023):
 * botón de 44 px con icono decorativo 🌙/☀️, **estado anunciado** (`aria-pressed`)
 * y **nombre accesible dinámico** («Cambiar a tema oscuro»/«Cambiar a tema
 * claro»). Cambia el tema **al instante sin recargar** vía `setTheme` (T1005) y
 * **no pierde el foco** al cambiar. Con elección guardada manda la manual; sin
 * ella, sigue la preferencia del sistema (D7).
 */
export function ThemeToggle({ className = '' }: ThemeToggleProps) {
  const { theme, setTheme } = useTheme();
  const { t } = useLanguage();

  const isDark = theme === 'dark';
  const next = isDark ? 'light' : 'dark';

  return (
    <button
      type="button"
      aria-pressed={isDark}
      aria-label={isDark ? t('theme.toLight') : t('theme.toDark')}
      onClick={() => setTheme(next)}
      className={[
        'inline-flex min-h-11 min-w-11 items-center justify-center rounded-full border border-line bg-surface-raised text-ink',
        'transition duration-[var(--motion-duration-micro)] ease-[var(--motion-ease-standard)] hover:bg-accent hover:text-navy',
        'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus',
        className,
      ]
        .filter(Boolean)
        .join(' ')}
    >
      <span aria-hidden="true" className="text-lg leading-none">
        {isDark ? '☀️' : '🌙'}
      </span>
    </button>
  );
}
