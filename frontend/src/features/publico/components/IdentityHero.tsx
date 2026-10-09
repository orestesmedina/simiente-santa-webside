import { useState } from 'react';
import type { IdentityPublic } from '../../../api/portada';
import { useLanguage } from '../i18n/useLanguage';
import { BrandLogo } from './BrandLogo';

export interface IdentityHeroProps {
  identity: IdentityPublic;
}

/**
 * Hero de la portada (ux.md §4.1): imagen de portada de fondo con capa `navy`,
 * logotipo `object-contain` (nunca se deforma, D-9), H1 con el nombre oficial y
 * los bloques de misión/visión/lema que existan. Si no hay imagen o falla su
 * carga, el hero conserva el fondo `navy` y todo el texto: nunca un hueco.
 */
export function IdentityHero({ identity }: IdentityHeroProps) {
  const { t } = useLanguage();
  const [coverFailed, setCoverFailed] = useState(false);
  const showCover = Boolean(identity.coverImageUrl) && !coverFailed;

  return (
    <section aria-labelledby="portada-nombre" className="relative bg-navy text-white">
      {showCover && (
        <img
          src={identity.coverImageUrl}
          alt={identity.coverImageAlt ?? t('a11y.hero')}
          className="absolute inset-0 h-full w-full object-cover"
          onError={() => setCoverFailed(true)}
        />
      )}
      <div aria-hidden="true" className="absolute inset-0 bg-navy/80" />
      <div className="relative mx-auto flex w-full max-w-6xl flex-col items-start gap-6 px-4 py-16 sm:px-6 sm:py-24">
        <BrandLogo src={identity.logoUrl} alt={identity.logoAlt ?? t('a11y.logo')} size={72} />
        <h1
          id="portada-nombre"
          className="font-display text-[clamp(2.5rem,7vw,4.75rem)] leading-none"
        >
          {identity.name}
        </h1>

        {(identity.mission || identity.vision) && (
          <div className="grid w-full gap-6 sm:grid-cols-2">
            {identity.mission && (
              <div>
                <h2 className="font-display text-2xl text-teal">{t('identity.mission')}</h2>
                <p className="mt-1 text-lg leading-relaxed">{identity.mission}</p>
              </div>
            )}
            {identity.vision && (
              <div>
                <h2 className="font-display text-2xl text-teal">{t('identity.vision')}</h2>
                <p className="mt-1 text-lg leading-relaxed">{identity.vision}</p>
              </div>
            )}
          </div>
        )}

        {identity.tagline && (
          <p className="font-emotiva text-[clamp(1.25rem,3vw,1.75rem)] italic">
            <span className="sr-only">{t('identity.lema')}: </span>
            {identity.tagline}
          </p>
        )}
      </div>
    </section>
  );
}
