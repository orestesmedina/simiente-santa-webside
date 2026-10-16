export interface IntroTextProps {
  /** Texto introductorio ya resuelto a `lang` por el servidor (FR-008/FR-021). */
  text: string;
}

/**
 * Texto introductorio de `/ofrendas` (FR-002/FR-021, ux.md §4.1/§6.3): columna
 * de lectura cómoda (~65ch) con los saltos de línea del equipo **preservados**
 * (`whitespace-pre-line`) y ningún rótulo propio —arranca la lectura—.
 *
 * **No se monta si el texto no viene** (en borrador o sin escribir): la sección
 * queda coherente, sin hueco ni marcador (US3 esc. 7). Esa decisión la toma la
 * página.
 */
export function IntroText({ text }: IntroTextProps) {
  return (
    <p className="max-w-[65ch] whitespace-pre-line text-lg leading-relaxed text-ink">{text}</p>
  );
}
