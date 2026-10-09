import { Button } from './Button';

export interface PaginationProps {
  /** Página actual, 1-based. */
  page: number;
  totalPages: number;
  onChange: (page: number) => void;
  /** Resumen visible y anunciado, p. ej. "Mostrando 1–20 de 45 registros". */
  summaryText: string;
}

/**
 * Paginación de `limit`/`offset`: "Anterior" y "Siguiente" deshabilitados en
 * los extremos. El resumen se anuncia con `role="status"` tras cada cambio
 * (ux.md §3.10/§6). Conservar los filtros es responsabilidad de la feature.
 */
export function Pagination({ page, totalPages, onChange, summaryText }: PaginationProps) {
  const safeTotal = Math.max(totalPages, 1);
  return (
    <nav aria-label="Paginación" className="flex flex-wrap items-center justify-between gap-3 pt-4">
      <p role="status" aria-live="polite" className="text-slate-700">
        {summaryText}
      </p>
      <div className="flex gap-2">
        <Button variant="secondary" onClick={() => onChange(page - 1)} disabled={page <= 1}>
          Anterior
        </Button>
        <Button variant="secondary" onClick={() => onChange(page + 1)} disabled={page >= safeTotal}>
          Siguiente
        </Button>
      </div>
    </nav>
  );
}
