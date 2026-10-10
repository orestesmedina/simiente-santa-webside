import { useId } from 'react';
import { useLanguage } from './useLanguage';

const OPTIONS = ['es', 'en'] as const;

/**
 * Selector de idioma del sitio público (ux.md §6.2/D-1): dos **radios nativos**
 * con los dos idiomas escritos completos, sin gestos ni auto-detección. Cambia
 * el `LanguageContext` y persiste la elección.
 *
 * El `name` del grupo es único por instancia (`useId`): la portada monta el
 * selector en la cabecera y en el pie, y compartir un mismo `name` haría que el
 * navegador los tratara como **un solo** grupo nativo, dejando sin marcar los
 * radios del que se montó primero.
 */
export function LanguageSwitcher({ className = '' }: { className?: string }) {
  const { lang, setLang, t } = useLanguage();
  const groupName = useId();

  return (
    <fieldset className={['flex items-center gap-3', className].filter(Boolean).join(' ')}>
      <legend className="sr-only">{t('lang.label')}</legend>
      {OPTIONS.map((code) => (
        <label key={code} className="inline-flex items-center gap-1.5 text-sm">
          <input
            type="radio"
            name={groupName}
            value={code}
            checked={lang === code}
            onChange={() => setLang(code)}
          />
          <span>{code === 'es' ? t('lang.es') : t('lang.en')}</span>
        </label>
      ))}
    </fieldset>
  );
}
