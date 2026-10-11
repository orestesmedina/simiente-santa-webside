import { useState, type ReactNode } from 'react';
import type { PortadaAdmin, PublicationState } from '../../../api/portada';
import { Button } from '../../../components/Button';
import { EmptyState } from '../../../components/EmptyState';
import { Notice } from '../../../components/Notice';
import { StatusPill } from '../../../components/StatusPill';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { ContactForm } from '../components/ContactForm';
import { IdentityForm } from '../components/IdentityForm';
import { ServicesList } from '../components/ServicesList';
import { SocialsList } from '../components/SocialsList';
import { WhatsAppList } from '../components/WhatsAppList';
import { WhoWeAreForm } from '../components/WhoWeAreForm';
import { isForbidden } from '../errors';
import { usePortadaAdmin } from '../hooks/usePortadaAdmin';
import {
  INFO_LOAD_ERROR,
  MODULE_HELP,
  MODULE_TITLE,
  NO_PERMISSION_ERROR,
  TAB_ABOUT,
  TAB_CONTACT,
  TAB_IDENTITY,
  TAB_SCHEDULE,
  TAB_SOCIALS,
  TAB_WHATSAPP,
} from '../messages';

/** Pestañas del módulo (ux.md §4.3). El panel va siempre en español (FR-016). */
export type InfoTab = 'identidad' | 'quienes-somos' | 'horario' | 'whatsapp' | 'redes' | 'contacto';

const TABS: TabItem[] = [
  { id: 'identidad', label: TAB_IDENTITY },
  { id: 'quienes-somos', label: TAB_ABOUT },
  { id: 'horario', label: TAB_SCHEDULE },
  { id: 'whatsapp', label: TAB_WHATSAPP },
  { id: 'redes', label: TAB_SOCIALS },
  { id: 'contacto', label: TAB_CONTACT },
];

interface SectionCardProps {
  title: string;
  state?: PublicationState;
  children: ReactNode;
}

function SectionCard({ title, state, children }: SectionCardProps) {
  return (
    <div className="space-y-3 rounded border border-slate-200 p-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h3 className="text-lg font-semibold text-slate-900">{title}</h3>
        {state && <StatusPill value={state} />}
      </div>
      {children}
    </div>
  );
}

/** Contenido de cada pestaña: formulario (singletons) o listado (colecciones). */
function renderTabContent(data: PortadaAdmin, tab: InfoTab): ReactNode {
  switch (tab) {
    case 'identidad':
      return (
        <SectionCard title={TAB_IDENTITY} state={data.identity?.publicationState}>
          <IdentityForm />
        </SectionCard>
      );
    case 'quienes-somos':
      return (
        <SectionCard title={TAB_ABOUT} state={data.about?.publicationState}>
          <WhoWeAreForm />
        </SectionCard>
      );
    case 'horario':
      return (
        <SectionCard title={TAB_SCHEDULE}>
          <ServicesList />
        </SectionCard>
      );
    case 'whatsapp':
      return (
        <SectionCard title={TAB_WHATSAPP}>
          <WhatsAppList />
        </SectionCard>
      );
    case 'redes':
      return (
        <SectionCard title={TAB_SOCIALS}>
          <SocialsList />
        </SectionCard>
      );
    case 'contacto':
      return (
        <SectionCard title={TAB_CONTACT} state={data.contact?.publicationState}>
          <ContactForm />
        </SectionCard>
      );
  }
}

/**
 * Vista general del módulo (ux.md §4.3, US2/US3): título, banda de ayuda y las
 * seis pestañas con el estado de cada pieza. Cubre los estados de carga, error
 * (con reintento) y sin permiso del servidor.
 */
export function InformationPage() {
  const [tab, setTab] = useState<InfoTab>('identidad');
  const query = usePortadaAdmin();

  return (
    <section className="space-y-4">
      <div>
        <h1 className="text-2xl font-bold text-slate-900">{MODULE_TITLE}</h1>
        <p className="mt-2 text-slate-700">{MODULE_HELP}</p>
      </div>

      <Tabs
        tabs={TABS}
        active={tab}
        onChange={(id) => setTab(id as InfoTab)}
        label="Secciones de la portada"
      />

      {query.isError && isForbidden(query.error) && (
        <Notice variant="error">{NO_PERMISSION_ERROR}</Notice>
      )}

      {query.isPending && <EmptyState kind="loading" message="Cargando información…" />}

      {query.isError && !isForbidden(query.error) && (
        <EmptyState
          kind="error"
          message={INFO_LOAD_ERROR}
          action={<Button onClick={() => void query.refetch()}>Reintentar</Button>}
        />
      )}

      {query.isSuccess && <div className="space-y-3">{renderTabContent(query.data, tab)}</div>}
    </section>
  );
}
