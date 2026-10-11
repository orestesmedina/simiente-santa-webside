import type { WhatsappChannelPublic } from '../../../api/portada';
import { useLanguage } from '../i18n/useLanguage';

export interface WhatsAppSectionProps {
  items: WhatsappChannelPublic[];
}

/** Icono de WhatsApp decorativo (el texto del CTA va siempre al lado). */
function WhatsAppIcon() {
  return (
    <svg
      aria-hidden="true"
      focusable="false"
      width={24}
      height={24}
      viewBox="0 0 24 24"
      fill="currentColor"
    >
      <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51l-.57-.01c-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.872.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 0 1-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 0 1-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 0 1 2.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0 0 12.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 0 0 5.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 0 0-3.48-8.413z" />
    </svg>
  );
}

/**
 * Canales de WhatsApp (ux.md §4.1): una tarjeta por canal con su nombre como
 * título y el CTA correcto según el tipo (`direct` → «Escribir por WhatsApp»,
 * `group` → «Entrar al grupo»). El enlace ya viene resuelto por el servidor
 * (`wa.me` o URL del grupo) y se abre con un clic (SC-009).
 */
export function WhatsAppSection({ items }: WhatsAppSectionProps) {
  const { t } = useLanguage();

  return (
    <section id="whatsapp" aria-labelledby="whatsapp-title" className="bg-cream py-12 sm:py-20">
      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <h2
          id="whatsapp-title"
          className="font-display text-[clamp(1.75rem,4.5vw,2.5rem)] leading-tight text-navy"
        >
          {t('nav.sections.whatsapp')}
        </h2>
        <ul className="mt-6 grid gap-4 sm:grid-cols-2">
          {items.map((channel) => (
            <li
              key={channel.id}
              className="rounded-2xl border border-navy-soft bg-white p-6 shadow-brand"
            >
              <h3 className="font-sans text-xl font-semibold text-navy">{channel.name}</h3>
              <a
                href={channel.url}
                target="_blank"
                rel="noopener noreferrer"
                className="mt-4 inline-flex min-h-12 items-center gap-2 rounded-lg bg-teal px-4 py-2 font-medium text-navy hover:bg-teal-strong hover:text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-navy"
              >
                <WhatsAppIcon />
                {channel.kind === 'direct' ? t('whatsapp.action.dm') : t('whatsapp.action.group')}
              </a>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
