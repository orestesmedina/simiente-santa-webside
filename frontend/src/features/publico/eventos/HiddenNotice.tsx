import { Card } from '../../../components/publico/Card';
import { useLanguage } from '../i18n/useLanguage';

/**
 * Aviso de sección sin contenido (ux.md §4.1, ⟲ D-13/FR-013): la página `/eventos`
 * existe, pero **sin redirigir** a la portada, y muestra solo el aviso de
 * interfaz —ningún dato real, ningún borrador—. El ítem de navegación tampoco
 * existe en barra, menú móvil ni pie (FR-022, `useSiteLinks`).
 */
export function HiddenNotice() {
  const { t } = useLanguage();

  return (
    <section data-testid="hidden-notice" className="py-section-tight">
      <div className="mx-auto w-full max-w-3xl px-4 sm:px-6">
        <Card variant="glass" className="flex flex-col gap-block">
          <p role="status" className="text-lg leading-relaxed text-ink">
            {t('events.empty.hidden')}
          </p>
        </Card>
      </div>
    </section>
  );
}
