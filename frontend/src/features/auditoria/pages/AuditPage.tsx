import { useState, type Dispatch, type SetStateAction } from 'react';
import { ApiError } from '../../../api/client';
import type { AccessEventItem, AdminActionItem } from '../../../api/auditoria';
import { Button } from '../../../components/Button';
import { DateRangeFilter } from '../../../components/DateRangeFilter';
import { EmptyState } from '../../../components/EmptyState';
import { Notice } from '../../../components/Notice';
import { Pagination } from '../../../components/Pagination';
import { Select, type SelectOption } from '../../../components/Select';
import { StatusPill } from '../../../components/StatusPill';
import { Table, type Column } from '../../../components/Table';
import { Tabs, type TabItem } from '../../../components/Tabs';
import { formatDateTime } from '../../../lib/format';
import { useUsers } from '../../usuarios/hooks/useUsers';
import { useAccessEvents } from '../hooks/useAccessEvents';
import { useAdminActions } from '../hooks/useAdminActions';
import {
  LOAD_AUDIT_ERROR,
  NO_PERMISSION_ERROR,
  UNIDENTIFIED_ATTEMPT,
  actionLabel,
  emptyAuditMessage,
} from '../messages';

const TABS: TabItem[] = [
  { id: 'accesos', label: 'Historial de accesos' },
  { id: 'acciones', label: 'Acciones administrativas' },
];

const ACCESS_COLUMNS: Column<AccessEventItem>[] = [
  { key: 'date', header: 'Fecha y hora', render: (event) => formatDateTime(event.createdAt) },
  {
    key: 'account',
    header: 'Cuenta',
    render: (event) => event.userName ?? UNIDENTIFIED_ATTEMPT,
  },
  {
    key: 'result',
    header: 'Resultado',
    render: (event) => <StatusPill value={event.result === 'success' ? 'success' : 'failure'} />,
  },
  { key: 'ip', header: 'IP de origen', render: (event) => event.ip },
];

const ACTION_COLUMNS: Column<AdminActionItem>[] = [
  { key: 'date', header: 'Fecha y hora', render: (item) => formatDateTime(item.createdAt) },
  { key: 'actor', header: 'Quién', render: (item) => item.actorName ?? 'Sistema' },
  { key: 'action', header: 'Acción', render: (item) => actionLabel(item.action) },
  { key: 'target', header: 'Sobre', render: (item) => item.targetLabel ?? '—' },
  {
    key: 'result',
    header: 'Resultado',
    render: (item) => (
      <StatusPill
        value={
          item.result === 'success'
            ? 'completed'
            : item.result === 'denied'
              ? 'denied'
              : 'not-completed'
        }
      />
    ),
  },
];

type AuditTab = 'accesos' | 'acciones';

interface HistoryFilters {
  /** Valor del selector de cuenta antes de pulsar "Filtrar". */
  accountDraft: string;
  fromDraft: string;
  toDraft: string;
  /** Filtros aplicados que viajan a la API. */
  account: string;
  from: string;
  to: string;
  page: number;
}

const INITIAL_FILTERS: HistoryFilters = {
  accountDraft: '',
  fromDraft: '',
  toDraft: '',
  account: '',
  from: '',
  to: '',
  page: 1,
};

function startOfDayIso(date: string): string | undefined {
  return date ? `${date}T00:00:00Z` : undefined;
}

/** `to` es exclusivo `[from, to)`: para incluir el día elegido, inicio del siguiente. */
function endExclusiveIso(date: string): string | undefined {
  if (!date) {
    return undefined;
  }
  const day = new Date(`${date}T00:00:00Z`);
  day.setUTCDate(day.getUTCDate() + 1);
  const year = day.getUTCFullYear();
  const month = String(day.getUTCMonth() + 1).padStart(2, '0');
  const dayOfMonth = String(day.getUTCDate()).padStart(2, '0');
  return `${year}-${month}-${dayOfMonth}T00:00:00Z`;
}

function isForbidden(error: unknown): boolean {
  return error instanceof ApiError && error.status === 403;
}

function hasFilters(filters: HistoryFilters): boolean {
  return Boolean(filters.account || filters.from || filters.to);
}

function summary(first: number, last: number, total: number): string {
  return `Mostrando ${first}–${last} de ${total} registros`;
}

interface AuditFiltersProps {
  accountOptions: SelectOption[];
  filters: HistoryFilters;
  onDraftChange: (patch: Partial<HistoryFilters>) => void;
  onApply: () => void;
  onClear: () => void;
}

function AuditFilters({
  accountOptions,
  filters,
  onDraftChange,
  onApply,
  onClear,
}: AuditFiltersProps) {
  return (
    <div className="flex flex-wrap items-end gap-3 rounded border border-slate-200 p-4">
      <Select
        label="Cuenta"
        options={accountOptions}
        value={filters.accountDraft}
        onChange={(event) => onDraftChange({ accountDraft: event.target.value })}
      />
      <DateRangeFilter
        from={filters.fromDraft}
        to={filters.toDraft}
        onFromChange={(value) => onDraftChange({ fromDraft: value })}
        onToChange={(value) => onDraftChange({ toDraft: value })}
        onApply={onApply}
        onClear={onClear}
      />
    </div>
  );
}

/**
 * Sección de registro de solo lectura (ux.md §3.10, US8): dos historiales
 * —accesos y acciones administrativas— con filtros por cuenta y rango de
 * fechas, paginación que conserva los filtros y **sin ningún control de
 * edición ni borrado** (FR-025). Un intento sin cuenta asociada se muestra como
 * "Intento sin cuenta asociada", sin correo alguno (F-01/FR-003).
 */
export function AuditPage() {
  const [tab, setTab] = useState<AuditTab>('accesos');
  const [accessFilters, setAccessFilters] = useState<HistoryFilters>(INITIAL_FILTERS);
  const [actionFilters, setActionFilters] = useState<HistoryFilters>(INITIAL_FILTERS);

  const accountsQuery = useUsers({ limit: 100 });

  const accessQuery = useAccessEvents({
    userId: accessFilters.account || undefined,
    from: startOfDayIso(accessFilters.from),
    to: endExclusiveIso(accessFilters.to),
    page: accessFilters.page,
    enabled: tab === 'accesos',
  });
  const actionsQuery = useAdminActions({
    userId: actionFilters.account || undefined,
    from: startOfDayIso(actionFilters.from),
    to: endExclusiveIso(actionFilters.to),
    page: actionFilters.page,
    enabled: tab === 'acciones',
  });

  const accountOptions: SelectOption[] = [
    { value: '', label: accountsQuery.isLoading ? 'Cargando cuentas…' : 'Todas' },
    ...(accountsQuery.data?.items ?? []).map((user) => ({
      value: user.id,
      label: `${user.firstName} ${user.lastName}`,
    })),
  ];

  const applyFilters = (setState: Dispatch<SetStateAction<HistoryFilters>>) => {
    setState((current) => ({
      ...current,
      account: current.accountDraft,
      from: current.fromDraft,
      to: current.toDraft,
      page: 1,
    }));
  };

  const changeDraft = (
    setState: Dispatch<SetStateAction<HistoryFilters>>,
    patch: Partial<HistoryFilters>,
  ) => {
    setState((current) => ({ ...current, ...patch }));
  };

  return (
    <section className="space-y-4">
      <h1 className="text-2xl font-bold text-slate-900">Auditoría</h1>

      <Tabs
        tabs={TABS}
        active={tab}
        onChange={(id) => setTab(id as AuditTab)}
        label="Historiales de auditoría"
      />

      {tab === 'accesos' && (
        <div className="space-y-4">
          <AuditFilters
            accountOptions={accountOptions}
            filters={accessFilters}
            onDraftChange={(patch) => changeDraft(setAccessFilters, patch)}
            onApply={() => applyFilters(setAccessFilters)}
            onClear={() => setAccessFilters(INITIAL_FILTERS)}
          />

          {accessQuery.isError && isForbidden(accessQuery.error) && (
            <Notice variant="error">{NO_PERMISSION_ERROR}</Notice>
          )}

          {accessQuery.isPending && <EmptyState kind="loading" message="Cargando registros…" />}

          {accessQuery.isError && !isForbidden(accessQuery.error) && (
            <EmptyState
              kind="error"
              message={LOAD_AUDIT_ERROR}
              action={<Button onClick={() => void accessQuery.refetch()}>Reintentar</Button>}
            />
          )}

          {accessQuery.isSuccess && accessQuery.data.items.length === 0 && (
            <EmptyState message={emptyAuditMessage(hasFilters(accessFilters))} />
          )}

          {accessQuery.isSuccess && accessQuery.data.items.length > 0 && (
            <div>
              <Table
                columns={ACCESS_COLUMNS}
                rows={accessQuery.data.items}
                rowKey={(event) => event.id}
                caption="Historial de accesos al panel"
              />
              <Pagination
                page={accessFilters.page}
                totalPages={Math.max(1, Math.ceil(accessQuery.data.total / accessQuery.data.limit))}
                onChange={(page) => setAccessFilters((current) => ({ ...current, page }))}
                summaryText={summary(
                  (accessFilters.page - 1) * accessQuery.data.limit + 1,
                  Math.min(accessFilters.page * accessQuery.data.limit, accessQuery.data.total),
                  accessQuery.data.total,
                )}
              />
            </div>
          )}
        </div>
      )}

      {tab === 'acciones' && (
        <div className="space-y-4">
          <AuditFilters
            accountOptions={accountOptions}
            filters={actionFilters}
            onDraftChange={(patch) => changeDraft(setActionFilters, patch)}
            onApply={() => applyFilters(setActionFilters)}
            onClear={() => setActionFilters(INITIAL_FILTERS)}
          />

          {actionsQuery.isError && isForbidden(actionsQuery.error) && (
            <Notice variant="error">{NO_PERMISSION_ERROR}</Notice>
          )}

          {actionsQuery.isPending && <EmptyState kind="loading" message="Cargando registros…" />}

          {actionsQuery.isError && !isForbidden(actionsQuery.error) && (
            <EmptyState
              kind="error"
              message={LOAD_AUDIT_ERROR}
              action={<Button onClick={() => void actionsQuery.refetch()}>Reintentar</Button>}
            />
          )}

          {actionsQuery.isSuccess && actionsQuery.data.items.length === 0 && (
            <EmptyState message={emptyAuditMessage(hasFilters(actionFilters))} />
          )}

          {actionsQuery.isSuccess && actionsQuery.data.items.length > 0 && (
            <div>
              <Table
                columns={ACTION_COLUMNS}
                rows={actionsQuery.data.items}
                rowKey={(item) => item.id}
                caption="Historial de acciones administrativas"
              />
              <Pagination
                page={actionFilters.page}
                totalPages={Math.max(
                  1,
                  Math.ceil(actionsQuery.data.total / actionsQuery.data.limit),
                )}
                onChange={(page) => setActionFilters((current) => ({ ...current, page }))}
                summaryText={summary(
                  (actionFilters.page - 1) * actionsQuery.data.limit + 1,
                  Math.min(actionFilters.page * actionsQuery.data.limit, actionsQuery.data.total),
                  actionsQuery.data.total,
                )}
              />
            </div>
          )}
        </div>
      )}
    </section>
  );
}
