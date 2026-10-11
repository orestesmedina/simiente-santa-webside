import { Button } from '../../../components/Button';
import { Notice } from '../../../components/Notice';
import { usePublicHomeData } from '../hooks/usePublicHomeData';
import { useLanguage } from '../i18n/useLanguage';
import { Reveal } from '../motion/Reveal';
import { RevealGroup } from '../motion/RevealGroup';
import { ContactSection } from '../components/ContactSection';
import { IdentityHero } from '../components/IdentityHero';
import { ScheduleSection } from '../components/ScheduleSection';
import { SectionSkeleton } from '../components/SectionSkeleton';
import { SocialSection } from '../components/SocialSection';
import { WhatsAppSection } from '../components/WhatsAppSection';
import { WhoWeAreSection } from '../components/WhoWeAreSection';

/**
 * Portada pública en `/` (ux.md §4.1, US1/US4/US5). Compone **solo** las
 * secciones presentes en la respuesta (SC-012: cero secciones vacías) y
 * renderiza los contenidos **tal como los resolvió el servidor** (analyze I2:
 * sin fallback de idioma en el cliente). Estados: esqueletos al cargar, error
 * con «Volver a intentar» sin contenido previo y aviso discreto si falla una
 * actualización con datos ya en pantalla.
 *
 * Movimiento (F10 · T1026/FR-005/FR-008): cada sección se envuelve en `Reveal`
 * (entrada lateral alternada en escritorio; vertical en móvil) y el conjunto se
 * escalona con `RevealGroup` (tope acotado). Las **claves React estables por
 * ancla** garantizan que cambiar de idioma o de tema re-renderiza **sin
 * re-montar**: las secciones ya vistas no se re-animan y el hero solo entra en
 * la primera carga real (SC-014).
 */
export function HomePage() {
  const { t } = useLanguage();
  const query = usePublicHomeData();

  if (query.isPending) {
    return (
      <div>
        <SectionSkeleton />
        <SectionSkeleton />
      </div>
    );
  }

  if (query.isError && !query.data) {
    return (
      <section className="bg-surface px-4 py-section-tight text-ink sm:px-6">
        <div className="mx-auto w-full max-w-3xl">
          <h1 className="font-display text-[clamp(2.5rem,7vw,4.75rem)] leading-none">
            {t('brand.name')}
          </h1>
          <p className="mt-4 text-lg leading-relaxed">{t('error.load')}</p>
          <Button variant="accent" className="mt-6" onClick={() => void query.refetch()}>
            {t('error.retry')}
          </Button>
        </div>
      </section>
    );
  }

  if (!query.data) {
    return null;
  }

  const { identity, about, schedule, whatsapp, socials, contact } = query.data;

  return (
    <>
      {query.isError && (
        <div className="mx-auto w-full max-w-6xl px-4 pt-4 sm:px-6">
          <Notice variant="warning">{t('error.partial')}</Notice>
        </div>
      )}

      {identity && (
        <IdentityHero
          identity={identity}
          whatsappHref={whatsapp && whatsapp.length > 0 ? '#whatsapp' : undefined}
        />
      )}

      <RevealGroup>
        {about && (
          <Reveal key="quienes-somos" from="left">
            <WhoWeAreSection about={about} />
          </Reveal>
        )}
        {schedule && schedule.length > 0 && (
          <Reveal key="horario" from="right">
            <ScheduleSection items={schedule} />
          </Reveal>
        )}
        {whatsapp && whatsapp.length > 0 && (
          <Reveal key="whatsapp" from="left">
            <WhatsAppSection items={whatsapp} />
          </Reveal>
        )}
        {contact && (
          <Reveal key="contacto" from="right">
            <ContactSection contact={contact} />
          </Reveal>
        )}
        {socials && socials.length > 0 && (
          <Reveal key="redes" from="left">
            <SocialSection items={socials} />
          </Reveal>
        )}
      </RevealGroup>
    </>
  );
}
