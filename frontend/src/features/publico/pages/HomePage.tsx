import { Button } from '../../../components/Button';
import { Notice } from '../../../components/Notice';
import { usePublicHomeData } from '../hooks/usePublicHomeData';
import { useLanguage } from '../i18n/useLanguage';
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
 */
export function HomePage() {
  const { t } = useLanguage();
  const query = usePublicHomeData();

  if (query.isPending) {
    return (
      <div className="py-4">
        <SectionSkeleton />
        <SectionSkeleton />
      </div>
    );
  }

  if (query.isError && !query.data) {
    return (
      <section className="bg-navy px-4 py-16 text-white sm:px-6">
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

      {identity && <IdentityHero identity={identity} />}
      {about && <WhoWeAreSection about={about} />}
      {schedule && schedule.length > 0 && <ScheduleSection items={schedule} />}
      {whatsapp && whatsapp.length > 0 && <WhatsAppSection items={whatsapp} />}
      {contact && <ContactSection contact={contact} />}
      {socials && socials.length > 0 && <SocialSection items={socials} />}
    </>
  );
}
