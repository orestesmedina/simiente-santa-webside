import { useLanguage } from '../i18n/useLanguage';

export interface CategoryChipProps {
  /** Nombre de la categoría ya resuelto al idioma activo (Q6). */
  name: string;
}

/**
 * Etiqueta de categoría del elemento (ux.md §4.1, ⟲ D-2): píldora pequeña con
 * borde del sistema y texto Poppins. **Sin color semántico**: la categoría
 * orienta sobre de qué se trata, no es un estado (lo que distingue un evento de
 * una actividad son los dos bloques con título, no el color).
 *
 * El nombre accesible es «Categoría: {nombre}» (`events.category.aria`), de
 * modo que un lector de pantalla no oye un texto huérfano.
 */
export function CategoryChip({ name }: CategoryChipProps) {
  const { t } = useLanguage();

  return (
    <span
      aria-label={t('events.category.aria', { name })}
      className="inline-flex w-fit items-center rounded-full border border-line px-3 py-1 font-sans text-sm text-muted"
    >
      {name}
    </span>
  );
}
