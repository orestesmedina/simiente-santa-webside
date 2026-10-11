import type { ContactPublic } from '../../../api/portada';
import { useLanguage } from '../i18n/useLanguage';

export interface ContactSectionProps {
  contact: ContactPublic;
}

/** Teléfono listo para `tel:` (sin espacios ni separadores). */
function phoneHref(phone: string): string {
  return `tel:${phone.replace(/[^\d+]/g, '')}`;
}

/**
 * Contacto (ux.md §4.1): dirección como texto plano, correo y teléfono como
 * enlaces nativos (`mailto:` / `tel:`) cómodos en el teléfono.
 */
export function ContactSection({ contact }: ContactSectionProps) {
  const { t } = useLanguage();

  return (
    <section id="contacto" aria-labelledby="contacto-title" className="bg-cream py-12 sm:py-20">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <h2
          id="contacto-title"
          className="font-display text-[clamp(1.75rem,4.5vw,2.5rem)] leading-tight text-navy"
        >
          {t('nav.sections.contact')}
        </h2>
        <dl className="mt-6 grid gap-4 sm:grid-cols-3">
          <div className="rounded-2xl border border-navy-soft bg-white p-4">
            <dt className="font-medium text-navy">{t('contact.address')}</dt>
            <dd className="mt-1 text-navy">{contact.address}</dd>
          </div>
          <div className="rounded-2xl border border-navy-soft bg-white p-4">
            <dt className="font-medium text-navy">{t('contact.email')}</dt>
            <dd className="mt-1">
              <a
                href={`mailto:${contact.email}`}
                className="inline-flex min-h-11 items-center text-navy underline underline-offset-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-navy"
              >
                {contact.email}
              </a>
            </dd>
          </div>
          <div className="rounded-2xl border border-navy-soft bg-white p-4">
            <dt className="font-medium text-navy">{t('contact.phone')}</dt>
            <dd className="mt-1">
              <a
                href={phoneHref(contact.phone)}
                className="inline-flex min-h-11 items-center text-navy underline underline-offset-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-navy"
              >
                {contact.phone}
              </a>
            </dd>
          </div>
        </dl>
      </div>
    </section>
  );
}
