import type { ContactPublic } from '../../../api/portada';
import { Card } from '../../../components/publico/Card';
import { SectionHeading } from '../../../components/publico/SectionHeading';
import { useLanguage } from '../i18n/useLanguage';

export interface ContactSectionProps {
  contact: ContactPublic;
}

/** Teléfono listo para `tel:` (sin espacios ni separadores). */
function phoneHref(phone: string): string {
  return `tel:${phone.replace(/[^\d+]/g, '')}`;
}

/** Fila-pastilla de contacto: microinteracción solo con `transform`. */
const ROW =
  'flex flex-col gap-1 rounded-2xl border border-line bg-surface-raised p-4 ' +
  'transition duration-[var(--motion-duration-micro)] ease-[var(--motion-ease-standard)] ' +
  'hover:translate-x-1.5';

const LINK =
  'inline-flex min-h-11 items-center text-ink underline underline-offset-4 ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-focus';

/**
 * Contacto con el sistema visual de la propuesta B (F10 · T1023): filas-pastilla
 * (dirección como **texto plano**, correo y teléfono como enlaces nativos
 * `mailto:`/`tel:`) con desplazamiento de 6 px en hover (**solo `transform`**:
 * nunca desplaza el maquetado, FR-008). Datos y contratos de F3 intactos.
 */
export function ContactSection({ contact }: ContactSectionProps) {
  const { t } = useLanguage();

  return (
    <section id="contacto" aria-labelledby="contacto-title" className="py-section">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <Card variant="glass" className="flex flex-col gap-block">
          <SectionHeading id="contacto-title" kicker="04">
            {t('nav.sections.contact')}
          </SectionHeading>

          <dl className="grid gap-fluid sm:grid-cols-3">
            <div className={ROW}>
              <dt className="font-medium text-muted">{t('contact.address')}</dt>
              <dd className="text-ink">{contact.address}</dd>
            </div>
            <div className={ROW}>
              <dt className="font-medium text-muted">{t('contact.email')}</dt>
              <dd className="mt-1">
                <a href={`mailto:${contact.email}`} className={LINK}>
                  {contact.email}
                </a>
              </dd>
            </div>
            <div className={ROW}>
              <dt className="font-medium text-muted">{t('contact.phone')}</dt>
              <dd className="mt-1">
                <a href={phoneHref(contact.phone)} className={LINK}>
                  {contact.phone}
                </a>
              </dd>
            </div>
          </dl>
        </Card>
      </div>
    </section>
  );
}
