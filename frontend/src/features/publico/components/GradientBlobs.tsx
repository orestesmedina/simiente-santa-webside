/**
 * Capa decorativa de la propuesta B «Marea Turquesa» (F10 · T1025/FR-001/FR-007):
 * dos blobs de gradiente en deriva continua al fondo.
 *
 * Son **decorativos** (`aria-hidden`) y no interceptan eventos
 * (`pointer-events: none`); solo animan `transform` y quedan **estáticos** con
 * `prefers-reduced-motion`. Nunca van detrás de texto esencial con contraste
 * insuficiente (el fondo del contenido es siempre `--surface` opaco) ni aplican
 * efectos sobre el logotipo de marca. El `blur` es moderado y acotado por el
 * contenedor con `overflow-hidden` (0 desbordes horizontales); si en el
 * dispositivo de referencia modesto degradara la fluidez, se reduce o se retira
 * antes que entorpecer la lectura (RG10-9/`ux.md` §9.4).
 */
export function GradientBlobs() {
  return (
    <div aria-hidden="true" className="pointer-events-none fixed inset-0 -z-10 overflow-hidden">
      <div className="public-blob public-blob-1" />
      <div className="public-blob public-blob-2" />
    </div>
  );
}
