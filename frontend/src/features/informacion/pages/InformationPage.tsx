import { useState, type ReactNode } from 'react';
import type { PortadaAdmin, PublicationState } from '../../../api/portada';
import { Button } from '../../../components/Button';
import { EmptyState } from '../../../components/EmptyState';
import { Notice } from '../../../components/Notice';
import { StatusPill } from '../../../components/StatusPill';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { SOCIAL_NETWORK_LABELS, SOCIAL_NETWORKS } from '../../publico/social';
import { ContactForm } from '../components/ContactForm';
import { IdentityForm } from '../components/IdentityForm';
import { WhoWeAreForm } from '../components/WhoWeAreForm';
import { isForbidden } from '../errors';
import { usePortadaAdmin } from '../hooks/usePortadaAdmin';
import {
  EMPTY_SERVICES,
  EMPTY_WHATSAPP,
  INFO_LOAD_ERROR,
  MODULE_HELP,
  MODULE_TITLE,
  NO_PERMISSION_ERROR,
  NO_SOCIAL_LINK,
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

function ItemSummary({ label, state }: { label: string; state: PublicationState }) {
  return (
    <li className="flex flex-wrap items-center justify-between gap-2 border-b border-slate-200 py-2 last:border-b-0">
      <span className="text-slate-900">{label}</span>
      <StatusPill value={state} />
    </li>
  );
}

/**
 * Contenido de cada pestaña. T333 muestra el **estado** de cada pieza (US3
 * esc. 7) con el agregado del panel; T334–T336 sustituyen cada resumen por el
 * formulario o listado real de la sección.
 */
function renderTabContent(data: PortadaAdmin, tab: InfoTab): ReactNode {
  switch (tab) {
    case 'identidad': {
      const identity = data.identity;
      return (
        <SectionCard title={TAB_IDENTITY} state={identity?.publicationState}>
          <IdentityForm />
        </SectionCard>
      );
    }
    case 'quienes-somos': {
      const about = data.about;
      return (
        <SectionCard title={TAB_ABOUT} state={about?.publicationState}>
          <WhoWeAreForm />
        </SectionCard>
      );
    }
    case 'contacto': {
      const contact = data.contact;
      return (
        <SectionCard title={TAB_CONTACT} state={contact?.publicationState}>
          <ContactForm />
        </SectionCard>
      );
    }
    case 'horario': {
      const items = data.schedule.items;
      return (
        <SectionCard title={TAB_SCHEDULE}>
          {items.length > 0 ? (
            <ul>
              {items.map((item) => (
                <ItemSummary
                  key={item.id}
                  label={`${item.nameEs} · ${item.startTime}`}
                  state={item.publicationState}
                />
              ))}
            </ul>
          ) : (
            <EmptyState message={EMPTY_SERVICES} />
          )}
        </SectionCard>
      );
    }
    case 'whatsapp': {
      const items = data.whatsapp.items;
      return (
        <SectionCard title={TAB_WHATSAPP}>
          {items.length > 0 ? (
            <ul>
              {items.map((item) => (
                <ItemSummary
                  key={item.id}
                  label={`${item.nameEs} · ${item.kind === 'direct' ? 'Mensaje directo' : 'Grupo'}`}
                  state={item.publicationState}
                />
              ))}
            </ul>
          ) : (
            <EmptyState message={EMPTY_WHATSAPP} />
          )}
        </SectionCard>
      );
    }
    case 'redes': {
      const byNetwork = new Map(data.socials.items.map((item) => [item.network, item]));
      return (
        <SectionCard title={TAB_SOCIALS}>
          <ul>
            {SOCIAL_NETWORKS.map((network) => {
              const link = byNetwork.get(network);
              return (
                <li
                  key={network}
                  className="flex flex-wrap items-center justify-between gap-2 border-b border-slate-200 py-2 last:border-b-0"
                >
                  <span className="text-slate-900">{SOCIAL_NETWORK_LABELS[network]}</span>
                  {link ? (
                    <StatusPill value={link.publicationState} />
                  ) : (
                    <span className="text-slate-600">{NO_SOCIAL_LINK}</span>
                  )}
                </li>
              );
            })}
          </ul>
        </SectionCard>
      );
    }
  }
}

/**
 * Vista general del módulo (ux.md §4.3, US2/US3): título, banda de ayuda y las
 * seis pestañas con el estado de cada pieza. Cubre los estados de carga, error
 * (con reintento) y sin permiso del servidor; los formularios y listados reales
 * llegan en T334–T336.
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
