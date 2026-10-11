import { useState } from 'react';
import { Field } from '../../../components/Field';
import { ENGLISH_HELP } from '../messages';
import { validateImage } from '../imageValidation';

export interface ImageUploaderProps {
  label: string;
  help?: string;
  /** URL de la imagen ya referenciada por la identidad (previsualización). */
  currentUrl?: string;
  /** Texto alternativo en español (obligatorio cuando hay imagen, FR-019). */
  altEs: string;
  onAltEsChange: (value: string) => void;
  /** Texto alternativo en inglés (opcional). */
  altEn: string;
  onAltEnChange: (value: string) => void;
  /** Recibe el archivo ya validado en el cliente (tipo y tamaño). */
  onFileSelected: (file: File) => void;
  /** Subida en curso: deshabilita el campo. */
  uploading?: boolean;
  /** Error del servidor (tipo/tamaño/texto alternativo) junto al bloque. */
  error?: string;
  /** Error de validación del `alt` español. */
  altError?: string;
}

/**
 * Subida de imagen del panel (ux.md §4.4/§10): campo de archivo, previsualización
 * inmediata con `FileReader` y validación de tipo/tamaño **antes** de enviar;
 * textos alternativos es (obligatorio) / en (opcional). La previsualización es
 * local y no pasa por `/api/v1/media` (analyze C4).
 */
export function ImageUploader({
  label,
  help,
  currentUrl,
  altEs,
  onAltEsChange,
  altEn,
  onAltEnChange,
  onFileSelected,
  uploading = false,
  error,
  altError,
}: ImageUploaderProps) {
  const [preview, setPreview] = useState<string | undefined>(currentUrl);
  const [localError, setLocalError] = useState<string>();

  const handleFile = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }
    const invalid = validateImage(file);
    if (invalid) {
      setLocalError(invalid);
      return;
    }
    setLocalError(undefined);
    const reader = new FileReader();
    reader.onload = () => {
      if (typeof reader.result === 'string') {
        setPreview(reader.result);
      }
    };
    reader.readAsDataURL(file);
    onFileSelected(file);
  };

  return (
    <fieldset className="space-y-3 rounded border border-slate-200 p-4">
      <legend className="px-1 font-medium text-slate-900">{label}</legend>

      {preview && (
        <img
          src={preview}
          alt={altEs || label}
          className="max-h-40 w-auto rounded border border-slate-200 object-contain"
        />
      )}

      <Field
        label="Archivo"
        type="file"
        accept="image/jpeg,image/png,image/webp"
        disabled={uploading}
        help={help}
        error={localError ?? error}
        onChange={handleFile}
      />

      <Field
        label={`Texto alternativo (Español)`}
        value={altEs}
        required
        error={altError}
        onChange={(event) => onAltEsChange(event.target.value)}
      />
      <Field
        label="Texto alternativo (English, opcional)"
        value={altEn}
        help={ENGLISH_HELP}
        onChange={(event) => onAltEnChange(event.target.value)}
      />
    </fieldset>
  );
}
