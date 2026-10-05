package usuarios

import (
	"context"

	"github.com/google/uuid"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/audit"
)

// Consultas de auditoría (FR-022…FR-025): historial de intentos de acceso y de
// acciones administrativas. SOLO inserción y consulta: el registro es de solo
// inserción (FR-025) y no existe ninguna sentencia de modificación ni borrado
// sobre `login_events`/`admin_actions`. Los campos derivados del contrato
// (userName/userEmail, actorName/actorEmail) se resuelven con LEFT JOIN en la
// consulta y se componen aquí (F-01): no se guardan duplicados.

// InsertLoginEvent registra un intento de inicio de sesión (FR-022). UserID es
// nil cuando el correo no corresponde a ninguna cuenta: el intento queda sin
// asociación y sin crear nada (FR-003). El correo probado NO se guarda (FR-026).
func (r *repository) InsertLoginEvent(ctx context.Context, event audit.Event) (LoginEvent, error) {
	row, err := r.q.InsertLoginEvent(ctx, gendb.InsertLoginEventParams{
		UserID: pgUUIDPtr(event.UserID),
		Result: string(event.Result),
		Ip:     event.IP,
	})
	if err != nil {
		return LoginEvent{}, wrap(err, "insert login event", "", "")
	}
	return mapInsertLoginEventRow(row), nil
}

// RecordLoginSuccess escribe en UNA única transacción la fila del intento
// exitoso y la proyección del último acceso de la cuenta (`last_login_at` +
// `last_login_ip`, FR-021/FR-022). Es el camino fail-closed del login: sin
// registro, no hay acceso (R23). El `created_at` de la fila es el propio reloj
// de la base, de modo que la proyección y el historial coinciden.
func (r *repository) RecordLoginSuccess(ctx context.Context, userID uuid.UUID, ip string) (LoginEvent, error) {
	var inserted LoginEvent
	err := r.withTx(ctx, func(tx *repository) error {
		row, insertErr := tx.InsertLoginEvent(ctx, audit.Event{
			UserID: &userID,
			Result: audit.ResultSuccess,
			IP:     ip,
		})
		if insertErr != nil {
			return insertErr
		}
		inserted = row
		return tx.UpdateUserLastLogin(ctx, userID, row.CreatedAt, ip)
	})
	if err != nil {
		return LoginEvent{}, err
	}
	return inserted, nil
}

// InsertAdminAction registra una acción administrativa sensible (FR-023),
// incluidos los intentos que fallan o se deniegan. Es de solo inserción.
func (r *repository) InsertAdminAction(ctx context.Context, action audit.Action) (AdminAction, error) {
	row, err := r.q.InsertAdminAction(ctx, gendb.InsertAdminActionParams{
		ActorUserID:  pgUUIDPtr(action.ActorUserID),
		Action:       action.Code,
		TargetKind:   string(action.TargetKind),
		TargetUserID: pgUUIDPtr(action.TargetUserID),
		TargetRoleID: pgUUIDPtr(action.TargetRoleID),
		TargetLabel:  pgLabel(action.TargetLabel),
		Result:       string(action.Result),
	})
	if err != nil {
		return AdminAction{}, wrap(err, "insert admin action", "", "")
	}
	return mapInsertAdminActionRow(row), nil
}

// ListLoginEvents devuelve una página del historial de accesos con los datos de
// cuenta derivados (FR-022/FR-024). Filtra por cuenta y semirango `[from, to)`.
func (r *repository) ListLoginEvents(ctx context.Context, filter AuditFilter) ([]AccessEvent, error) {
	rows, err := r.q.ListLoginEvents(ctx, gendb.ListLoginEventsParams{
		UserID: pgUUIDPtr(filter.UserID),
		From:   pgTimestamptzOpt(filter.From),
		To:     pgTimestamptzOpt(filter.To),
		Off:    int32(filter.Offset),
		Lim:    int32(filter.Limit),
	})
	if err != nil {
		return nil, wrap(err, "list login events", "", "")
	}
	events := make([]AccessEvent, 0, len(rows))
	for _, row := range rows {
		events = append(events, mapListLoginEventsRow(row))
	}
	return events, nil
}

// CountLoginEvents cuenta el historial con los mismos filtros que el listado.
func (r *repository) CountLoginEvents(ctx context.Context, filter AuditFilter) (int64, error) {
	total, err := r.q.CountLoginEvents(ctx, gendb.CountLoginEventsParams{
		UserID: pgUUIDPtr(filter.UserID),
		From:   pgTimestamptzOpt(filter.From),
		To:     pgTimestamptzOpt(filter.To),
	})
	if err != nil {
		return 0, wrap(err, "count login events", "", "")
	}
	return total, nil
}

// ListAdminActions devuelve una página del historial de acciones con el actor
// derivado (FR-023/FR-024). El filtro por cuenta incluye a la cuenta que hizo
// la acción Y a la cuenta objetivo (actor OR target).
func (r *repository) ListAdminActions(ctx context.Context, filter AuditFilter) ([]AdminActionEntry, error) {
	rows, err := r.q.ListAdminActions(ctx, gendb.ListAdminActionsParams{
		UserID: pgUUIDPtr(filter.UserID),
		From:   pgTimestamptzOpt(filter.From),
		To:     pgTimestamptzOpt(filter.To),
		Off:    int32(filter.Offset),
		Lim:    int32(filter.Limit),
	})
	if err != nil {
		return nil, wrap(err, "list admin actions", "", "")
	}
	entries := make([]AdminActionEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, mapListAdminActionsRow(row))
	}
	return entries, nil
}

// CountAdminActions cuenta el historial con los mismos filtros que el listado.
func (r *repository) CountAdminActions(ctx context.Context, filter AuditFilter) (int64, error) {
	total, err := r.q.CountAdminActions(ctx, gendb.CountAdminActionsParams{
		UserID: pgUUIDPtr(filter.UserID),
		From:   pgTimestamptzOpt(filter.From),
		To:     pgTimestamptzOpt(filter.To),
	})
	if err != nil {
		return 0, wrap(err, "count admin actions", "", "")
	}
	return total, nil
}
