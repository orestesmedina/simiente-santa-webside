import { useState } from 'react';
import { mediaUrl } from '../../../api/client';
import type { IdentityPublic } from '../../../api/portada';
import { Card } from '../../../components/publico/Card';
import { PublicButton } from '../../../components/publico/PublicButton';
import { useLanguage } from '../i18n/useLanguage';
import { BrandLogo } from './BrandLogo';

export interface IdentityHeroProps {
  identity: IdentityPublic;
  /**
   * Ancla del bloque de WhatsApp. Si falta (no hay canal publicado), el CTA de
   * WhatsApp **no** se muestra: nunca se pinta un enlace a un destino ausente.
   */
  whatsappHref?: string;
  /**
   * Ancla del bloque de horario. Si falta (no hay servicios publicados), el CTA
   * «Ver horario» **no** se muestra (FR-027/SC-015: 0 enlaces rotos).
   */
  scheduleHref?: string;
}

/**
 * Hero de identidad con la propuesta B «Marea Turquesa» (F10 · T1021).
 *
 * Composición: texto a la izquierda (logotipo intacto sobre `--logo-surface`,
 * H1 en Bebas con el **nombre oficial** —contrato congelado, `portada-nombre`—,
 * misión/visión, lema en Playfair itálica y CTAs) e imagen de portada en un
 * **marco de vidrio** (`Card glass`) a la derecha; en móvil, columna.
 *
 * El CTA de WhatsApp usa un nombre accesible **propio** («Escríbenos por
 * WhatsApp», D19/C2) que no contiene los de `WhatsAppSection` («Escribir por
 * WhatsApp»/«Entrar al grupo»), para no colisionar con las consultas congeladas
 * de coincidencia única. La entrada del hero usa `--motion-duration-hero` y solo
 * ocurre en la primera carga real (claves React estables, FR-008); con
 * movimiento reducido o sin JS el hero se ve completo al instante.
 */
export function IdentityHero({ identity, whatsappHref, scheduleHref }: IdentityHeroProps) {
  const { t } = useLanguage();
  const [coverFailed, setCoverFailed] = useState(false);
  // La API devuelve URLs relativas (`/api/v1/media/…`): se resuelven contra la
  // base de la API para que el navegador no las pida al origen de la SPA (bug F3).
  const coverSrc = mediaUrl(identity.coverImageUrl);
  const showCover = Boolean(coverSrc) && !coverFailed;

  return (
    <section
      aria-labelledby="portada-nombre"
      className="hero-enter mx-auto flex w-full max-w-6xl flex-col items-start gap-block px-4 py-section-tight sm:px-6 min-[861px]:grid min-[861px]:grid-cols-[1.1fr_0.9fr] min-[861px]:items-center"
    >
      <div className="flex w-full flex-col items-start gap-block">
        <BrandLogo
          src={mediaUrl(identity.logoUrl)}
          alt={identity.logoAlt ?? t('a11y.logo')}
          size={72}
          className="rounded-full bg-logo-surface p-1"
        />
        <h1
          id="portada-nombre"
          className="font-display text-[clamp(2.5rem,7vw,4.75rem)] leading-none text-ink"
        >
          {identity.name}
        </h1>

        {(identity.mission || identity.vision) && (
          <div className="grid w-full gap-block sm:grid-cols-2">
            {identity.mission && (
              <div>
                <h2 className="font-display text-2xl text-accent">{t('identity.mission')}</h2>
                <p className="mt-1 text-lg leading-relaxed text-ink">{identity.mission}</p>
              </div>
            )}
            {identity.vision && (
              <div>
                <h2 className="font-display text-2xl text-accent">{t('identity.vision')}</h2>
                <p className="mt-1 text-lg leading-relaxed text-ink">{identity.vision}</p>
              </div>
            )}
          </div>
        )}

        {identity.tagline && (
          <p className="font-emotiva text-[clamp(1.25rem,3vw,1.75rem)] italic text-muted">
            <span className="sr-only">{t('identity.lema')}: </span>
            {identity.tagline}
          </p>
        )}

        <div className="flex flex-wrap gap-fluid">
          {scheduleHref && (
            <PublicButton href={scheduleHref}>{t('hero.cta.schedule')}</PublicButton>
          )}
          {whatsappHref && (
            <PublicButton href={whatsappHref} variant="whatsapp">
              {t('hero.cta.whatsapp')}
            </PublicButton>
          )}
        </div>
      </div>

      <Card variant="glass" padded={false} className="w-full overflow-hidden">
        {showCover ? (
          <img
            src={coverSrc}
            alt={identity.coverImageAlt ?? t('a11y.hero')}
            className="aspect-[4/3] h-full w-full object-cover"
            onError={() => setCoverFailed(true)}
          />
        ) : (
          <div
            aria-hidden="true"
            className="aspect-[4/3] w-full bg-gradient-to-br from-accent/25 to-transparent"
          />
        )}
      </Card>
    </section>
  );
}
