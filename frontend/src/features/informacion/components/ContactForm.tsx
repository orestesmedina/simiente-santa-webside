import { useState } from 'react';
import { z } from 'zod';
import type { ContactAdmin, ContactInput, PublicationState } from '../../../api/portada';
import { Button } from '../../../components/Button';
import { Field } from '../../../components/Field';
import { Notice } from '../../../components/Notice';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { generalErrorMessage } from '../errors';
import { useContact } from '../hooks/useContact';
import { useSectionPublication } from '../hooks/useSectionPublication';
import { useSyncedFields } from '../hooks/useSyncedFields';
import { useUpdateContact } from '../hooks/useUpdateContact';
import {
  DRAFT_NOTICE,
  EMAIL_INVALID,
  ENGLISH_HELP,
  PHONE_INVALID,
  REQUIRED_ADDRESS,
  REQUIRED_EMAIL,
  REQUIRED_PHONE,
} from '../messages';
import { isValidEmail, isValidPhone } from '../validation';
import { PublishControls } from './PublishControls';
import { TextArea } from './TextArea';

const LANGUAGE_TABS: TabItem[] = [
  { id: 'es', label: 'Español' },
  { id: 'en', label: 'English (opcional)' },
];

const SERVER_FIELDS = ['addressEs', 'addressEn', 'email', 'phone'] as const;

interface ContactFields {
  addressEs: string;
  addressEn: string;
  email: string;
  phone: string;
}

function fieldsFromContact(contact: ContactAdmin | null): ContactFields {
  return {
    addressEs: contact?.addressEs ?? '',
    addressEn: contact?.addressEn ?? '',
    email: contact?.email ?? '',
    phone: contact?.phone ?? '',
  };
}

const contactSchema = z.object({
  addressEs: z.string().trim().min(1, REQUIRED_ADDRESS),
  email: z.string().trim().min(1, REQUIRED_EMAIL).refine(isValidEmail, EMAIL_INVALID),
  phone: z.string().trim().min(1, REQUIRED_PHONE).refine(isValidPhone, PHONE_INVALID),
});

/**
 * Contacto (ux.md §4.9, FR-007/FR-013/FR-015): dirección es/en (español
 * obligatorio), correo y teléfono obligatorios con formato; la **sección** se
 * publica/retira por su cuenta (D-2), de modo que nunca quede medio publicada.
 */
export function ContactForm() {
  const { contact } = useContact();
  const updateContact = useUpdateContact();
  const [langTab, setLangTab] = useState<'es' | 'en'>('es');
  const [fields, update] = useSyncedFields(contact?.updatedAt ?? null, () =>
    fieldsFromContact(contact),
  );
  const [clientErrors, setClientErrors] = useState<Record<string, string>>({});
  const publication = useSectionPublication<ContactAdmin, ContactInput>({
    admin: contact,
    mutation: updateContact,
    serverFieldNames: SERVER_FIELDS,
  });

  const errors = { ...clientErrors, ...publication.serverErrors };

  const buildInput = (state: PublicationState): ContactInput => ({
    addressEs: fields.addressEs.trim(),
    addressEn: fields.addressEn.trim() === '' ? null : fields.addressEn.trim(),
    email: fields.email.trim(),
    phone: fields.phone.trim(),
    publicationState: state,
  });

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const parsed = contactSchema.safeParse(fields);
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
      {updateContact.error && Object.keys(publication.serverErrors).length === 0 && (
        <Notice variant="error">{generalErrorMessage(updateContact.error)}</Notice>
      )}

      <Tabs
        tabs={LANGUAGE_TABS}
        active={langTab}
        onChange={(id) => setLangTab(id as 'es' | 'en')}
        label="Idioma del contenido"
      />

      {langTab === 'es' ? (
        <TextArea
          label="Dirección (Español)"
          required
          value={fields.addressEs}
          error={errors.addressEs}
          onChange={(event) => update({ addressEs: event.target.value })}
        />
      ) : (
        <>
          <p className="text-sm text-slate-600">{ENGLISH_HELP}</p>
          <TextArea
            label="Dirección (English)"
            value={fields.addressEn}
            onChange={(event) => update({ addressEn: event.target.value })}
          />
        </>
      )}

      <Field
        label="Correo"
        type="email"
        required
        value={fields.email}
        error={errors.email}
        onChange={(event) => update({ email: event.target.value })}
      />
      <Field
        label="Teléfono"
        required
        value={fields.phone}
        help="Con código del país: +506 8888 8888."
        error={errors.phone}
        onChange={(event) => update({ phone: event.target.value })}
      />

      <div className="flex flex-wrap items-center justify-end gap-3">
        <Button type="submit" loading={publication.busy}>
          Guardar
        </Button>
        {contact && (
          <PublishControls
            name="Contacto"
            state={contact.publicationState}
            loading={publication.busy}
            onPublish={() => publication.publish(buildInput)}
            onUnpublish={() => publication.unpublish(buildInput)}
          />
        )}
      </div>
    </form>
  );
}
