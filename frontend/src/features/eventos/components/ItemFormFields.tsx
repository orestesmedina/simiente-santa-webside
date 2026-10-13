import type { ReactNode } from 'react';
import { Field } from '../../../components/Field';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { TextArea } from '../../informacion/components/TextArea';
import {
  DETAIL_MAX,
  ENGLISH_HELP,
  MAP_HELP,
  ORGANIZER_HELP,
  PLACE_MAX,
  TITLE_MAX,
  counter,
} from '../messages';
import type { ItemFields } from '../validation';

const LANGUAGE_TABS: TabItem[] = [
  { id: 'es', label: 'Español' },
  { id: 'en', label: 'English (opcional)' },
];

export interface ItemFormFieldsProps {
  fields: ItemFields;
  onChange: (patch: Partial<ItemFields>) => void;
  errors: Record<string, string>;
  langTab: 'es' | 'en';
  onLangTabChange: (tab: 'es' | 'en') => void;
  /** Campo extra de la pestaña español (la frecuencia, solo en eventos). */
  spanishExtra?: ReactNode;
  /** Campo extra de la pestaña English (la frecuencia, solo en eventos). */
  englishExtra?: ReactNode;
  /** Campos no lingüísticos (fechas, categoría, imagen) tras las pestañas. */
  children?: ReactNode;
}

/**
 * Campos compartidos por el alta/edición de eventos y actividades (ux.md §4.4/
 * §4.5): pestañas **Español** (obligatorio) / **English (opcional)**, título,
 * lugar, enlace de mapa, detalle con contador vivo y organizador. El español es
 * el idioma base: dejar el inglés vacío es válido y la ayuda lo explica
 * (Decisión 8). Los campos no lingüísticos llegan como `children`.
 */
export function ItemFormFields({
  fields,
  onChange,
  errors,
  langTab,
  onLangTabChange,
  spanishExtra,
  englishExtra,
  children,
}: ItemFormFieldsProps) {
  return (
    <>
      <Tabs
        tabs={LANGUAGE_TABS}
        active={langTab}
        onChange={(id) => onLangTabChange(id as 'es' | 'en')}
        label="Idioma del contenido"
      />

      {langTab === 'es' ? (
        <>
          <Field
            id="agenda-title-es"
            label="Título (Español)"
            required
            value={fields.titleEs}
            help={counter(fields.titleEs.length, TITLE_MAX)}
            error={errors.titleEs}
            onChange={(event) => onChange({ titleEs: event.target.value })}
          />
          {spanishExtra}
          <Field
            id="agenda-place-es"
            label="Lugar (Español)"
            required
            value={fields.placeEs}
            help={counter(fields.placeEs.length, PLACE_MAX)}
            error={errors.placeEs}
            onChange={(event) => onChange({ placeEs: event.target.value })}
          />
          <Field
            id="agenda-map-url"
            label="Enlace de mapa (opcional)"
            value={fields.mapUrl}
            help={MAP_HELP}
            error={errors.mapUrl}
            onChange={(event) => onChange({ mapUrl: event.target.value })}
          />
          <TextArea
            id="agenda-detail-es"
            label="Detalle (Español)"
            required
            value={fields.detailEs}
            help={counter(fields.detailEs.length, DETAIL_MAX)}
            error={errors.detailEs}
            onChange={(event) => onChange({ detailEs: event.target.value })}
          />
          <Field
            label="Organizador o encargado (opcional)"
            value={fields.organizerEs}
            help={ORGANIZER_HELP}
            error={errors.organizerEs}
            onChange={(event) => onChange({ organizerEs: event.target.value })}
          />
        </>
      ) : (
        <>
          <p className="text-sm text-slate-600">{ENGLISH_HELP}</p>
          <Field
            id="agenda-title-en"
            label="Título (English)"
            value={fields.titleEn}
            help={counter(fields.titleEn.length, TITLE_MAX)}
            error={errors.titleEn}
            onChange={(event) => onChange({ titleEn: event.target.value })}
          />
          {englishExtra}
          <Field
            id="agenda-place-en"
            label="Lugar (English)"
            value={fields.placeEn}
            error={errors.placeEn}
            onChange={(event) => onChange({ placeEn: event.target.value })}
          />
          <TextArea
            id="agenda-detail-en"
            label="Detalle (English)"
            value={fields.detailEn}
            error={errors.detailEn}
            onChange={(event) => onChange({ detailEn: event.target.value })}
          />
          <Field
            label="Organizador o encargado (English)"
            value={fields.organizerEn}
            error={errors.organizerEn}
            onChange={(event) => onChange({ organizerEn: event.target.value })}
          />
        </>
      )}

      {children}
    </>
  );
}
