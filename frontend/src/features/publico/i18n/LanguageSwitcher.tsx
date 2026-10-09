import { useLanguage } from './useLanguage';

const OPTIONS = ['es', 'en'] as const;

/**
 * Selector de idioma del sitio público (ux.md §6.2/D-1): dos **radios nativos**
 * con los dos idiomas escritos completos, sin gestos ni auto-detección. Cambia
 * el `LanguageContext` y persiste la elección.
 */
export function LanguageSwitcher({ className = '' }: { className?: string }) {
  const { lang, setLang, t } = useLanguage();

  return (
    <fieldset className={['flex items-center gap-3', className].filter(Boolean).join(' ')}>
      <legend className="sr-only">{t('lang.label')}</legend>
      {OPTIONS.map((code) => (
        <label key={code} className="inline-flex items-center gap-1.5 text-sm">
          <input
            type="radio"
            name="ss-lang"
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
