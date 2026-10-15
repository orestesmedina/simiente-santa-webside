import { Button } from '../../../components/Button';
import { Notice } from '../../../components/Notice';
import { usePublicHomeData } from '../hooks/usePublicHomeData';
import { useLanguage } from '../i18n/useLanguage';
import { Reveal } from '../motion/Reveal';
import { RevealGroup } from '../motion/RevealGroup';
import { ContactSection } from '../components/ContactSection';
import { IdentityHero } from '../components/IdentityHero';
import { ProximamenteBlock } from '../eventos/ProximamenteBlock';
import { MinisteriosBlock } from '../ministerios/MinisteriosBlock';
import { OfrendasBlock } from '../ofrendas/OfrendasBlock';
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
          scheduleHref={schedule && schedule.length > 0 ? '#horario' : undefined}
        />
      )}

      <RevealGroup>
        {/* Bloque «Lo que viene» (F4/US7, ux.md §4.10): **primera sección** de la
            portada, inmediatamente después del hero y antes de «quiénes somos»
            (ajuste del 2026-10-12). Es la **única excepción** a «ocultar
            secciones vacías»: se pinta siempre, con su mensaje cuando no hay
            ítems. `from="right"` conserva la alternancia y deja intactas las
            direcciones de entrada de las secciones de F3. */}
        <Reveal key="proximamente" from="right">
          <ProximamenteBlock />
        </Reveal>
        {/* Bloque «Ministerios» (F6/US6, ux.md §4.9): **una sola inserción**
            inmediatamente después del bloque de eventos y antes de «quiénes
            somos» (FR-027/FR-029). Se oculta por completo sin publicados —sin
            mensaje, a diferencia del bloque de F4—. `from="left"` es la
            dirección propuesta por `ux.md` §3 (decisión menor de implementación;
            las demás secciones no se tocan). */}
        <Reveal key="ministerios" from="left">
          <MinisteriosBlock />
        </Reveal>
        {/* Bloque «Ofrendas y diezmos» (F7/US6, ux.md §4.9): **una sola inserción**
            inmediatamente después del bloque de ministerios y antes de «quiénes
            somos» (Q7/FR-022). Se oculta por completo sin métodos publicados —sin
            mensaje, como el de ministerios—. `from="right"` conserva la
            alternancia con el bloque de F6 (⟲ D-15). */}
        <Reveal key="ofrendas" from="right">
          <OfrendasBlock />
        </Reveal>
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
