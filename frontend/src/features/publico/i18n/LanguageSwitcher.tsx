import { useId } from 'react';
import { useLanguage } from './useLanguage';

const OPTIONS = ['es', 'en'] as const;

/**
 * Selector de idioma del sitio público (ux.md §0.6/§6.2, F10 · T1027): dos
 * **radios nativos** con los dos idiomas escritos completos («Español» /
 * «English»), sin gestos ni auto-detección.
 *
 * F10 solo **compacta el formato** (Q5/FR-015/FR-017): una pastilla fina con
 * borde translúcido, espaciado más ajustado y el idioma activo resaltado con el
 * acento del sistema. **Sin** píldoras de siglas, **sin** menú desplegable y
 * **sin** banderas. El **comportamiento de F3 queda intacto** (FR-016): radios
 * nativos, `useId` para el grupo, `localStorage['ss.lang']`, cambio inmediato,
 * memoria entre visitas y primera visita en español.
 *
 * El cuerpo de texto se mantiene en 0.875 rem (14 px, umbral de legibilidad M4)
 * y cada opción conserva objetivo táctil ≥ 44 px. El `name` del grupo es único
 * por instancia (`useId`): la portada monta el selector en la cabecera y en el
 * pie, y compartir un mismo `name` haría que el navegador los tratara como **un
 * solo** grupo nativo, dejando sin marcar los radios del que se montó primero.
 */
export function LanguageSwitcher({ className = '' }: { className?: string }) {
  const { lang, setLang, t } = useLanguage();
  const groupName = useId();

  return (
    <fieldset
      className={[
        'inline-flex items-center gap-1 rounded-full border border-line px-1.5 py-0.5',
        className,
      ]
        .filter(Boolean)
        .join(' ')}
    >
      <legend className="sr-only">{t('lang.label')}</legend>
      {OPTIONS.map((code) => {
        const active = lang === code;
        return (
          <label
            key={code}
            className={[
              'inline-flex min-h-11 cursor-pointer items-center gap-1 rounded-full px-2 text-sm',
              active ? 'bg-accent/15 font-bold text-accent' : 'font-medium text-muted',
            ].join(' ')}
          >
            <input
              type="radio"
              name={groupName}
              value={code}
              checked={active}
              onChange={() => setLang(code)}
            />
            <span>{code === 'es' ? t('lang.es') : t('lang.en')}</span>
          </label>
        );
      })}
    </fieldset>
  );
}
