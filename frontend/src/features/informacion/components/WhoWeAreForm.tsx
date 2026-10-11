import { useState } from 'react';
import { z } from 'zod';
import type { AboutAdmin, AboutInput, PublicationState } from '../../../api/portada';
import { Button } from '../../../components/Button';
import { Notice } from '../../../components/Notice';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { generalErrorMessage } from '../errors';
import { useSectionPublication } from '../hooks/useSectionPublication';
import { useSyncedFields } from '../hooks/useSyncedFields';
import { useUpdateWhoWeAre } from '../hooks/useUpdateWhoWeAre';
import { useWhoWeAre } from '../hooks/useWhoWeAre';
import {
  ABOUT_HELP,
  DRAFT_NOTICE,
  ENGLISH_HELP,
  REQUIRED_ABOUT,
  aboutLimitMessage,
} from '../messages';
import { PublishControls } from './PublishControls';
import { TextArea } from './TextArea';

const MAX_ABOUT = 1000;

const LANGUAGE_TABS: TabItem[] = [
  { id: 'es', label: 'Español' },
  { id: 'en', label: 'English (opcional)' },
];

const SERVER_FIELDS = ['textEs', 'textEn'] as const;

interface AboutFields {
  textEs: string;
  textEn: string;
}

function fieldsFromAbout(about: AboutAdmin | null): AboutFields {
  return { textEs: about?.textEs ?? '', textEn: about?.textEn ?? '' };
}

const aboutSchema = z.object({ textEs: z.string() }).superRefine((values, ctx) => {
  if (values.textEs.trim() === '') {
    ctx.addIssue({ code: 'custom', path: ['textEs'], message: REQUIRED_ABOUT });
  } else if (values.textEs.length > MAX_ABOUT) {
    ctx.addIssue({
      code: 'custom',
      path: ['textEs'],
      message: aboutLimitMessage(values.textEs.length),
    });
  }
});

/**
 * «Quiénes somos» (ux.md §4.5, FR-003/FR-013/FR-015): texto plano con contador
 * de 1.000 caracteres, pestañas es/en (el inglés es opcional) y publicación por
 * sección. Al superar el límite el contador avisa y el guardado se bloquea.
 */
export function WhoWeAreForm() {
  const { about } = useWhoWeAre();
  const updateAbout = useUpdateWhoWeAre();
  const [langTab, setLangTab] = useState<'es' | 'en'>('es');
  const [fields, update] = useSyncedFields(about?.updatedAt ?? null, () => fieldsFromAbout(about));
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({});
  const publication = useSectionPublication<AboutAdmin, AboutInput>({
    admin: about,
    mutation: updateAbout,
    serverFieldNames: SERVER_FIELDS,
  });

  const count = fields.textEs.length;
  const overLimit = count > MAX_ABOUT;
  const errors = { ...clientErrors, ...publication.serverErrors };

  const buildInput = (state: PublicationState): AboutInput => ({
    textEs: fields.textEs.trim(),
    textEn: fields.textEn.trim() === '' ? null : fields.textEn.trim(),
    publicationState: state,
  });

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const parsed = aboutSchema.safeParse(fields);
    if (!parsed.success) {
      const mapped: Record<string, string> = {};
      for (const issue of parsed.error.issues) {
        const key = issue.path.join('.');
        if (!(key in mapped)) {
          mapped[key] = issue.message;
        }
      }
      setClientErrors(mapped);
      return;
    }
    setClientErrors({});
    publication.save(buildInput);
  };

  return (
    <form onSubmit={handleSubmit} noValidate className="space-y-4">
      {publication.feedback && (
        <Notice variant={publication.feedback.variant}>{publication.feedback.message}</Notice>
      )}
      {publication.draftNotice && <Notice variant="info">{DRAFT_NOTICE}</Notice>}
      {updateAbout.error && Object.keys(publication.serverErrors).length === 0 && (
        <Notice variant="error">{generalErrorMessage(updateAbout.error)}</Notice>
      )}

      <Tabs
        tabs={LANGUAGE_TABS}
        active={langTab}
        onChange={(id) => setLangTab(id as 'es' | 'en')}
        label="Idioma del contenido"
      />

      {langTab === 'es' ? (
        <>
          <TextArea
            label="Texto (Español)"
            required
            value={fields.textEs}
            help={ABOUT_HELP}
            error={errors.textEs}
            onChange={(event) => update({ textEs: event.target.value })}
          />
          <p className={overLimit ? 'text-sm font-medium text-red-700' : 'text-sm text-slate-600'}>
            {count} de 1.000 caracteres
          </p>
          {overLimit && (
            <p role="alert" className="text-sm font-medium text-red-700">
              {aboutLimitMessage(count)}
            </p>
          )}
        </>
      ) : (
        <>
          <p className="text-sm text-slate-600">{ENGLISH_HELP}</p>
          <TextArea
            label="Texto (English)"
            value={fields.textEn}
            onChange={(event) => update({ textEn: event.target.value })}
          />
        </>
      )}

      <div className="flex flex-wrap items-center justify-end gap-3">
        <Button type="submit" loading={publication.busy} disabled={overLimit}>
          Guardar
        </Button>
        {about && (
          <PublishControls
            name="Quiénes somos"
            state={about.publicationState}
            loading={publication.busy}
            onPublish={() => publication.publish(buildInput)}
            onUnpublish={() => publication.unpublish(buildInput)}
          />
        )}
      </div>
    </form>
  );
}
