//go:build integration

package usuarios

import (
	"context"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/audit"
)

// Pruebas de integración de repository_audit.go contra PostgreSQL real:
// inserción, filtros, paginación, campos derivados y solo inserción (FR-025).

func TestIntegrationLoginEventsInsertAndFilters(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos")
	user := mustInsertUser(t, repo, "ana@ejemplo.com", role.ID, true)

	success, err := repo.InsertLoginEvent(ctx, audit.Event{UserID: &user.ID, Result: audit.ResultSuccess, IP: "10.0.0.1"})
	if err != nil {
		t.Fatalf("InsertLoginEvent(success): %v", err)
	}
	if _, err := repo.InsertLoginEvent(ctx, audit.Event{UserID: &user.ID, Result: audit.ResultFailure, IP: "10.0.0.2"}); err != nil {
		t.Fatalf("InsertLoginEvent(failure): %v", err)
	}
	// Intento con correo inexistente: user_id NULL y sin crear nada (FR-003).
	if _, err := repo.InsertLoginEvent(ctx, audit.Event{Result: audit.ResultFailure, IP: "10.0.0.3"}); err != nil {
		t.Fatalf("InsertLoginEvent(anónimo): %v", err)
	}

	all, err := repo.ListLoginEvents(ctx, AuditFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListLoginEvents: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("ListLoginEvents = %d, se esperaban 3", len(all))
	}
	if total, err := repo.CountLoginEvents(ctx, AuditFilter{}); err != nil || total != 3 {
		t.Fatalf("CountLoginEvents = %d (%v)", total, err)
	}

	// Filtro por cuenta: excluye el intento sin cuenta asociada.
	byUser, err := repo.ListLoginEvents(ctx, AuditFilter{UserID: &user.ID, Limit: 10})
	if err != nil {
		t.Fatalf("ListLoginEvents(user): %v", err)
	}
	if len(byUser) != 2 {
		t.Fatalf("página por cuenta = %d, se esperaban 2", len(byUser))
	}

	// Los campos derivados de un intento identificado se resuelven con JOIN.
	if byUser[0].UserEmail == nil || *byUser[0].UserEmail != "ana@ejemplo.com" {
		t.Fatalf("userEmail derivado = %v", byUser[0].UserEmail)
	}
	if byUser[0].UserName == nil || *byUser[0].UserName != "Ana Pérez" {
		t.Fatalf("userName derivado = %v", byUser[0].UserName)
	}

	// Un intento sin cuenta no guarda ni muestra correo (FR-026).
	anonymous, err := repo.ListLoginEvents(ctx, AuditFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListLoginEvents: %v", err)
	}
	seenAnonymous := false
	for _, event := range anonymous {
		if event.UserID == nil {
			seenAnonymous = true
			if event.UserEmail != nil || event.UserName != nil {
				t.Fatalf("el intento sin cuenta filtró datos: %+v", event)
			}
		}
	}
	if !seenAnonymous {
		t.Fatal("no se encontró el intento sin cuenta asociada")
	}

	// Semirango [from, to): `from` es inclusivo (>=) y `to` es exclusivo (<).
	// El primer intento fue el más antiguo: todos caen en [created, ∞) y
	// ninguno es anterior a created.
	created := success.CreatedAt
	if total, err := repo.CountLoginEvents(ctx, AuditFilter{From: &created}); err != nil || total != 3 {
		t.Fatalf("filtro from inclusivo = %d (%v), se esperaban 3", total, err)
	}
	if total, err := repo.CountLoginEvents(ctx, AuditFilter{To: &created}); err != nil || total != 0 {
		t.Fatalf("filtro to exclusivo = %d (%v), se esperaba 0", total, err)
	}

	// Paginación: LIMIT/OFFSET.
	page, err := repo.ListLoginEvents(ctx, AuditFilter{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("ListLoginEvents(paginado): %v", err)
	}
	if len(page) != 1 {
		t.Fatalf("página de 1 = %d filas", len(page))
	}
}

func TestIntegrationAdminActionsFiltersAndRoleTarget(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos")
	// El rol objetivo de la acción se elimina luego: no puede tener cuentas
	// asignadas (FR-017), por eso las cuentas usan `role`, no `deletable`.
	deletable := mustInsertRole(t, repo, "Temporal", "noticias")
	actor := mustInsertUser(t, repo, "admin@ejemplo.com", role.ID, true)
	target := mustInsertUser(t, repo, "objetivo@ejemplo.com", role.ID, true)

	label := "Temporal"
	if _, err := repo.InsertAdminAction(ctx, audit.Action{
		ActorUserID: &actor.ID, Code: audit.ActionUserUpdate, TargetKind: audit.TargetUser,
		TargetUserID: &target.ID, TargetLabel: "objetivo@ejemplo.com", Result: audit.ResultSuccess,
	}); err != nil {
		t.Fatalf("InsertAdminAction(user): %v", err)
	}
	if _, err := repo.InsertAdminAction(ctx, audit.Action{
		ActorUserID: &actor.ID, Code: audit.ActionRoleDelete, TargetKind: audit.TargetRole,
		TargetRoleID: &deletable.ID, TargetLabel: label, Result: audit.ResultSuccess,
	}); err != nil {
		t.Fatalf("InsertAdminAction(role): %v", err)
	}

	// El filtro por cuenta es actor OR objetivo.
	byTarget, err := repo.ListAdminActions(ctx, AuditFilter{UserID: &target.ID, Limit: 10})
	if err != nil || len(byTarget) != 1 {
		t.Fatalf("filtro por objetivo = %d (%v)", len(byTarget), err)
	}
	if byTarget[0].ActorEmail == nil || *byTarget[0].ActorEmail != "admin@ejemplo.com" {
		t.Fatalf("actorEmail derivado = %v", byTarget[0].ActorEmail)
	}

	// Eliminar el rol deja target_role_id NULL y conserva target_label (FR-025).
	if _, err := repo.DeleteRole(ctx, deletable.ID); err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}
	entries, err := repo.ListAdminActions(ctx, AuditFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListAdminActions: %v", err)
	}
	var roleAction *AdminActionEntry
	for i := range entries {
		if entries[i].Action == audit.ActionRoleDelete {
			roleAction = &entries[i]
		}
	}
	if roleAction == nil {
		t.Fatal("no se encontró la acción role.delete")
	}
	if roleAction.TargetID != nil {
		t.Fatalf("targetId tras eliminar el rol = %v, se esperaba nil", roleAction.TargetID)
	}
	if roleAction.TargetLabel == nil || *roleAction.TargetLabel != label {
		t.Fatalf("targetLabel tras eliminar el rol = %v, se esperaba %q", roleAction.TargetLabel, label)
	}

	if total, err := repo.CountAdminActions(ctx, AuditFilter{}); err != nil || total != 2 {
		t.Fatalf("CountAdminActions = %d (%v)", total, err)
	}
}

func TestIntegrationLoginProjectionCoherence(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos")
	user := mustInsertUser(t, repo, "ana@ejemplo.com", role.ID, true)

	// El login exitoso escribe su fila y la proyección en una única transacción.
	inserted, err := repo.RecordLoginSuccess(ctx, user.ID, "10.0.0.9")
	if err != nil {
		t.Fatalf("RecordLoginSuccess: %v", err)
	}

	account, err := repo.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if account.LastLoginAt == nil || !account.LastLoginAt.Equal(inserted.CreatedAt) {
		t.Fatalf("lastLoginAt = %v, se esperaba %v (fila success)", account.LastLoginAt, inserted.CreatedAt)
	}
	if account.LastLoginIP == nil || *account.LastLoginIP != "10.0.0.9" {
		t.Fatalf("lastLoginIp = %v", account.LastLoginIP)
	}
}

// TestIntegrationDeniedAndRejectedSurviveMissingTarget fija FR-023 contra
// PostgreSQL real: una denegación o un rechazo contra un {id} inexistente deja
// su fila (denied/failure) con el objetivo en NULL, sin perderse por la FK.
func TestIntegrationDeniedAndRejectedSurviveMissingTarget(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	role := mustInsertRole(t, repo, "Editor", "eventos")
	actor := mustInsertUser(t, repo, "admin@ejemplo.com", role.ID, true)
	missing := uuid.New()
	service := NewAuditService(repo, nil)

	if err := service.RecordDenied(ctx, audit.Denial{
		ActorUserID: &actor.ID, Method: http.MethodPatch,
		Path: "/api/v1/admin/usuarios/" + missing.String(),
	}); err != nil {
		t.Fatalf("RecordDenied con objetivo inexistente: %v", err)
	}
	service.RecordRejectedBestEffort(ctx, Rejection{
		ActorUserID: &actor.ID, Method: http.MethodDelete,
		Path: "/api/v1/admin/roles/" + missing.String(),
	})

	entries, err := repo.ListAdminActions(ctx, AuditFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListAdminActions: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("acciones registradas = %d, se esperaban 2", len(entries))
	}

	byAction := make(map[string]AdminActionEntry, len(entries))
	for _, entry := range entries {
		byAction[entry.Action] = entry
	}

	denied, ok := byAction[audit.ActionUserUpdate]
	if !ok {
		t.Fatal("no se registró la denegación de user.update")
	}
	if denied.Result != audit.ResultDenied || denied.TargetID != nil {
		t.Fatalf("fila denied = %+v, se esperaba result=denied con targetId NULL", denied)
	}
	if denied.TargetKind != audit.TargetUser {
		t.Fatalf("target_kind = %q, se esperaba %q", denied.TargetKind, audit.TargetUser)
	}

	rejected, ok := byAction[audit.ActionRoleDelete]
	if !ok {
		t.Fatal("no se registró el rechazo de role.delete")
	}
	if rejected.Result != audit.ResultFailure || rejected.TargetID != nil {
		t.Fatalf("fila failure = %+v, se esperaba result=failure con targetId NULL", rejected)
	}
	if rejected.TargetKind != audit.TargetRole {
		t.Fatalf("target_kind = %q, se esperaba %q", rejected.TargetKind, audit.TargetRole)
	}
}

// TestIntegrationAuditIsInsertOnly comprueba que repository_audit.go no
// contiene ninguna sentencia UPDATE/DELETE sobre el registro (FR-025).
func TestIntegrationAuditIsInsertOnly(t *testing.T) {
	source, err := os.ReadFile("repository_audit.go")
	if err != nil {
		t.Fatalf("leer repository_audit.go: %v", err)
	}
	re := regexp.MustCompile(`(?i)\b(UPDATE|DELETE)\s`)
	if match := re.FindString(string(source)); match != "" {
		t.Fatalf("repository_audit.go contiene %q (el registro es de solo inserción, FR-025)", match)
	}
}
