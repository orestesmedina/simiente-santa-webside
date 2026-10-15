import { useState } from 'react';
import type { DonacionAdmin } from '../../../api/ofrendas';
import { Button } from '../../../components/Button';
import { ConfirmDialog } from '../../../components/ConfirmDialog';
import { EmptyState } from '../../../components/EmptyState';
import { Notice, type NoticeVariant } from '../../../components/Notice';
import { PublishControls } from '../../../components/PublishControls';
import { SectionHeading } from '../../../components/publico/SectionHeading';
import { StatusPill } from '../../../components/StatusPill';
import { Table, type Column } from '../../../components/Table';
import { useDeleteDonacion } from '../hooks/useDeleteDonacion';
import { useUpdateDonacion } from '../hooks/useUpdateDonacion';
import {
  ADD_ACCOUNT,
  ADD_SINPE,
  COL_ACTIONS,
  COL_CURRENCY,
  COL_HOLDER,
  COL_INSTITUTION,
  COL_NUMBER,
  COL_ORDER,
  COL_STATE,
  CURRENCY_COLONES,
  CURRENCY_DOLARES,
  DELETED,
  DELETE_ACTION,
  DELETE_CANCEL,
  DELETE_CONFIRM_ACTION,
  DELETE_CONFIRM_TITLE,
  EDIT_ACTION,
  EMPTY_DASH,
  GROUP_IBAN_TITLE,
  GROUP_SINPE_TITLE,
  LOADING_ACCOUNTS,
  LOAD_ERROR_ACCOUNTS,
  PUBLISHED,
  RETRY,
  RETIRE_ACTION,
  RETIRE_CANCEL,
  RETIRE_CONFIRM_TITLE,
  TYPE_IBAN,
  TYPE_SINPE,
  UNPUBLISHED,
  deleteConfirmDescription,
  emptyGroupMessage,
  retireConfirmDescription,
} from '../messages';
import { NumeroDisplay } from './NumeroDisplay';

interface Feedback {
  variant: NoticeVariant;
  message: string;
}

export interface DonacionesListsByTypeProps {
  /** Métodos del módulo, ya en el orden del visitante (`iban` primero). */
  donaciones: DonacionAdmin[];
  /** Consulta del panel en vuelo: cada bloque muestra su estado de carga. */
  loading?: boolean;
  /** La consulta falló sin datos: cada bloque ofrece «Reintentar». */
  error?: boolean;
  onRetry?: () => void;
}

/** «Colones»/«Dólares» en IBAN; el SINPE nunca llega aquí (Q2). */
function currencyLabel(donacion: DonacionAdmin): string {
  if (donacion.currency === 'CRC') {
    return CURRENCY_COLONES;
  }
  if (donacion.currency === 'USD') {
    return CURRENCY_DOLARES;
  }
  return EMPTY_DASH;
}

/** Titular en español o «—» (el listado del panel sí usa el guion, ⟲ D-16). */
function holderLabel(donacion: DonacionAdmin): string {
  return donacion.holderEs?.trim() || EMPTY_DASH;
}

/**
 * Listado de métodos del panel **agrupado por tipo** (ux.md §4.3/§5.b, Q1,
 * US2/US3): dos bloques —«Cuentas IBAN» (kicker 01) y «Números de SINPE Móvil»
 * (kicker 02)— con su botón de agregar y una `Table` (tarjetas en móvil) con
 * **Orden**, **Número** (`NumeroDisplay`, copia exacta), **Entidad**, **Titular**
 * («—» si vacío), **Moneda** (solo la tabla IBAN: la columna no existe en SINPE,
 * Q2), **Estado** (`StatusPill`) y **Acciones** (Editar · Publicar/Retirar ·
 * Eliminar).
 *
 * El orden lo resuelve el servidor: `iban` antes que `sinpe` y, dentro de cada
 * grupo, `sortOrder`, `id` — el mismo orden que consume el público (FR-004). El
 * componente no reordena.
 *
 * Publicar es directo; retirar pide confirmación suave y eliminar una explícita
 * de no recuperable (FR-017). El alta y la edición abren `DonacionForm` en un
 * `Dialog` (T730): en esta tarea los botones quedan montados a la espera del
 * formulario.
 */
export function DonacionesListsByType({
  donaciones,
  loading = false,
  error = false,
  onRetry,
}: DonacionesListsByTypeProps) {
  const updateDonacion = useUpdateDonacion();
  const deleteDonacion = useDeleteDonacion();

  const [feedback, setFeedback] = useState<Feedback | null>(null);
  const [deleting, setDeleting] = useState<DonacionAdmin | null>(null);

  const iban = donaciones.filter((donacion) => donacion.kind === 'iban');
  const sinpe = donaciones.filter((donacion) => donacion.kind === 'sinpe');

  /** Columnas comunes; `withCurrency` solo en IBAN (Q2). */
  function columns(withCurrency: boolean): Column<DonacionAdmin>[] {
    const list: Column<DonacionAdmin>[] = [
      { key: 'sortOrder', header: COL_ORDER, render: (d) => String(d.sortOrder) },
      {
        key: 'number',
        header: COL_NUMBER,
        render: (d) => <NumeroDisplay kind={d.kind} number={d.number} />,
      },
      { key: 'institution', header: COL_INSTITUTION, render: (d) => d.institution },
      { key: 'holder', header: COL_HOLDER, render: holderLabel },
    ];
    if (withCurrency) {
      list.push({ key: 'currency', header: COL_CURRENCY, render: currencyLabel });
    }
    list.push({
      key: 'state',
      header: COL_STATE,
      render: (d) => <StatusPill value={d.publicationState} />,
    });
    list.push({
      key: 'actions',
      header: COL_ACTIONS,
      render: (d) => (
        <div className="flex flex-wrap items-center gap-2">
          {/* El alta y la edición abren `DonacionForm` (T730). */}
          <Button size="sm" variant="secondary" aria-haspopup="dialog">
            {EDIT_ACTION}
          </Button>
          <PublishControls
            name={d.institution}
            state={d.publicationState}
            loading={updateDonacion.isPending}
            retireLabel={RETIRE_ACTION}
            confirmTitle={RETIRE_CONFIRM_TITLE}
            confirmDescription={retireConfirmDescription}
            cancelText={RETIRE_CANCEL}
            onPublish={() =>
              updateDonacion.mutate(
                { id: d.id, input: { publicationState: 'published' } },
                { onSuccess: () => setFeedback({ variant: 'success', message: PUBLISHED }) },
              )
            }
            onUnpublish={() =>
              updateDonacion.mutate(
                { id: d.id, input: { publicationState: 'draft' } },
                { onSuccess: () => setFeedback({ variant: 'success', message: UNPUBLISHED }) },
              )
            }
          />
          <Button size="sm" variant="danger" aria-haspopup="dialog" onClick={() => setDeleting(d)}>
            {DELETE_ACTION}
          </Button>
        </div>
      ),
    });
    return list;
  }

  /** Un bloque de tipo con su encabezado, botón y estados (§5.b). */
  function renderBlock(
    kind: 'iban' | 'sinpe',
    kicker: string,
    title: string,
    addLabel: string,
    items: DonacionAdmin[],
    withCurrency: boolean,
  ) {
    const showHeaderButton = loading || items.length > 0;
    return (
      <section key={kind} className="space-y-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <SectionHeading kicker={kicker}>{title}</SectionHeading>
          {showHeaderButton && <Button aria-haspopup="dialog">{addLabel}</Button>}
        </div>
        {loading ? (
          <EmptyState kind="loading" message={LOADING_ACCOUNTS} />
        ) : error ? (
          <EmptyState
            kind="error"
            message={LOAD_ERROR_ACCOUNTS}
            action={<Button onClick={() => onRetry?.()}>{RETRY}</Button>}
          />
        ) : items.length === 0 ? (
          <EmptyState
            message={emptyGroupMessage(kind)}
            action={<Button aria-haspopup="dialog">{addLabel}</Button>}
          />
        ) : (
          <Table
            columns={columns(withCurrency)}
            rows={items}
            rowKey={(d) => d.id}
            caption={title}
          />
        )}
      </section>
    );
  }

  return (
    <div className="space-y-4">
      {feedback && <Notice variant={feedback.variant}>{feedback.message}</Notice>}

      {renderBlock('iban', '01', GROUP_IBAN_TITLE, ADD_ACCOUNT, iban, true)}
      {renderBlock('sinpe', '02', GROUP_SINPE_TITLE, ADD_SINPE, sinpe, false)}

      {deleting && (
        <ConfirmDialog
          title={DELETE_CONFIRM_TITLE}
          description={deleteConfirmDescription(
            deleting.kind === 'iban' ? TYPE_IBAN : TYPE_SINPE,
            deleting.institution,
            deleting.number,
          )}
          confirmText={DELETE_CONFIRM_ACTION}
          cancelText={DELETE_CANCEL}
          danger
          loading={deleteDonacion.isPending}
          onConfirm={() =>
            deleteDonacion.mutate(deleting.id, {
              onSuccess: () => {
                setDeleting(null);
                setFeedback({ variant: 'success', message: DELETED });
              },
            })
          }
          onClose={() => setDeleting(null)}
        />
      )}
    </div>
  );
}
