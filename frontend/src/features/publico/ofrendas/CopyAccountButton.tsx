import { useEffect, useState } from 'react';
import { PublicButton } from '../../../components/publico/PublicButton';
import { useLanguage } from '../i18n/useLanguage';

export interface CopyAccountButtonProps {
  /** Número exacto y **normalizado** que se copia al portapapeles (FR-003/SC-009). */
  number: string;
  /** Entidad financiera: completa el nombre accesible del botón (FR-025). */
  institution: string;
}

/** Tiempo que se muestra la confirmación de copiado antes de volver a la ayuda. */
const FEEDBACK_MS = 2000;

/**
 * Botón «Copiar número» de una ficha pública (FR-003, ⟲ D-5): icono **con
 * texto**, objetivo táctil ≥ 44 px (a todo el ancho en móvil) y nombre accesible
 * completo («Copiar número de {entidad}»), de modo que varias fichas no
 * compartan nombre.
 *
 * Al pulsarlo copia al portapapeles el número **tal como se guardó** —normalizado,
 * sin los espacios de presentación— y muestra la confirmación «Número copiado.»
 * en una región `role="status"` durante 2 s. La ayuda de transcripción por
 * bloques es **siempre visible** (US5, personas mayores): si el portapapeles no
 * está disponible, esa misma ayuda queda en pantalla y nunca hay silencio.
 */
export function CopyAccountButton({ number, institution }: CopyAccountButtonProps) {
  const { t } = useLanguage();
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!copied) {
      return;
    }
    const timer = window.setTimeout(() => setCopied(false), FEEDBACK_MS);
    return () => window.clearTimeout(timer);
  }, [copied]);

  async function handleCopy() {
    try {
      if (!navigator.clipboard?.writeText) {
        throw new Error('portapapeles no disponible');
      }
      await navigator.clipboard.writeText(number);
      setCopied(true);
    } catch {
      // Sin portapapeles nunca hay silencio: la ayuda de transcripción sigue
      // visible bajo el botón (⟲ D-5).
      setCopied(false);
    }
  }

  return (
    <div className="flex flex-col items-start gap-2 max-sm:w-full">
      <PublicButton
        variant="secondary"
        className="max-sm:w-full"
        ariaLabel={t('a11y.offerings.copy', { entidad: institution })}
        onClick={() => void handleCopy()}
      >
        <span aria-hidden="true">⧉</span>
        {t('offerings.copy')}
      </PublicButton>

      <p
        role="status"
        className={copied ? 'text-sm font-semibold text-accent-strong' : 'text-sm text-muted'}
      >
        {copied ? t('offerings.copied') : t('offerings.copy.hint')}
      </p>
    </div>
  );
}
