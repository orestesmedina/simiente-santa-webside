package usuarios

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/apperr"
)

// Tests unitarios del mapeo pgtype.* → dominio (§8.1.6) y de la traducción de
// errores de PostgreSQL a apperr. No tocan la base de datos.

func TestUUIDPtrNullAndValue(t *testing.T) {
	if got := uuidPtr(pgtype.UUID{}); got != nil {
		t.Fatalf("uuidPtr(NULL) = %v, se esperaba nil", got)
	}
	want := uuid.New()
	got := uuidPtr(pgtype.UUID{Bytes: [16]byte(want), Valid: true})
	if got == nil || *got != want {
		t.Fatalf("uuidPtr(valid) = %v, se esperaba %v", got, want)
	}
}

func TestTimePtrNullAndValue(t *testing.T) {
	if got := timePtr(pgtype.Timestamptz{}); got != nil {
		t.Fatalf("timePtr(NULL) = %v, se esperaba nil", got)
	}
	want := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	got := timePtr(pgtype.Timestamptz{Time: want, Valid: true})
	if got == nil || !got.Equal(want) {
		t.Fatalf("timePtr(valid) = %v, se esperaba %v", got, want)
	}
}

func TestTextPtrNullAndValue(t *testing.T) {
	if got := textPtr(pgtype.Text{}); got != nil {
		t.Fatalf("textPtr(NULL) = %v, se esperaba nil", got)
	}
	got := textPtr(pgtype.Text{String: "127.0.0.1", Valid: true})
	if got == nil || *got != "127.0.0.1" {
		t.Fatalf("textPtr(valid) = %v, se esperaba 127.0.0.1", got)
	}
}

func TestPGLabelEmptyIsNull(t *testing.T) {
	if got := pgLabel(""); got.Valid {
		t.Fatalf("pgLabel(\"\") = %+v, se esperaba NULL", got)
	}
	if got := pgLabel("ana@ejemplo.com"); !got.Valid || got.String != "ana@ejemplo.com" {
		t.Fatalf("pgLabel(valid) = %+v", got)
	}
}

func TestParamHelpers(t *testing.T) {
	id := uuid.New()
	pID := pgUUID(id)
	if !pID.Valid || uuid.UUID(pID.Bytes) != id {
		t.Fatalf("pgUUID = %+v", pID)
	}
	if ptr := pgUUIDPtr(&id); !ptr.Valid {
		t.Fatalf("pgUUIDPtr(no nil) = %+v", ptr)
	}
	if ptr := pgUUIDPtr(nil); ptr.Valid {
		t.Fatalf("pgUUIDPtr(nil) = %+v, se esperaba NULL", ptr)
	}

	at := time.Date(2026, 10, 4, 8, 0, 0, 0, time.UTC)
	if v := pgTimestamptz(at); !v.Valid || !v.Time.Equal(at) {
		t.Fatalf("pgTimestamptz = %+v", v)
	}
	if v := pgTimestamptzOpt(&at); !v.Valid || !v.Time.Equal(at) {
		t.Fatalf("pgTimestamptzOpt(no nil) = %+v", v)
	}
	if v := pgTimestamptzOpt(nil); v.Valid {
		t.Fatalf("pgTimestamptzOpt(nil) = %+v, se esperaba NULL", v)
	}

	ip := "10.0.0.1"
	if v := pgTextPtr(&ip); !v.Valid || v.String != ip {
		t.Fatalf("pgTextPtr(no nil) = %+v", v)
	}
	if v := pgTextPtr(nil); v.Valid {
		t.Fatalf("pgTextPtr(nil) = %+v, se esperaba NULL", v)
	}
}

func TestFullNamePtrComposition(t *testing.T) {
	tests := []struct {
		name      string
		firstName pgtype.Text
		lastName  pgtype.Text
		want      *string
	}{
		{"ambos nulos", pgtype.Text{}, pgtype.Text{}, nil},
		{"solo nombre", pgtype.Text{String: "Ana", Valid: true}, pgtype.Text{}, strPtr("Ana")},
		{"solo apellidos", pgtype.Text{}, pgtype.Text{String: "Pérez", Valid: true}, strPtr("Pérez")},
		{"ambos", pgtype.Text{String: "Ana", Valid: true}, pgtype.Text{String: "Pérez", Valid: true}, strPtr("Ana Pérez")},
		{"vacíos válidos", pgtype.Text{String: "", Valid: true}, pgtype.Text{String: "", Valid: true}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := fullNamePtr(tc.firstName, tc.lastName)
			switch {
			case tc.want == nil && got != nil:
				t.Fatalf("fullNamePtr() = %q, se esperaba nil", *got)
			case tc.want != nil && (got == nil || *got != *tc.want):
				t.Fatalf("fullNamePtr() = %v, se esperaba %q", got, *tc.want)
			}
		})
	}
}

func TestEnsureStringsNeverNil(t *testing.T) {
	if got := ensureStrings(nil); got == nil || len(got) != 0 {
		t.Fatalf("ensureStrings(nil) = %#v, se esperaba []", got)
	}
	in := []string{"eventos"}
	if got := ensureStrings(in); len(got) != 1 || got[0] != "eventos" {
		t.Fatalf("ensureStrings() = %#v", got)
	}
}

func TestMapUserRowsNullLastLogin(t *testing.T) {
	id := uuid.New()
	roleID := uuid.New()
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updated := created.Add(time.Hour)

	got := mapGetUserByIDRow(gendb.GetUserByIDRow{
		ID:                 pgtype.UUID{Bytes: [16]byte(id), Valid: true},
		Email:              "ana@ejemplo.com",
		FirstName:          "Ana",
		LastName:           "Pérez",
		Phone:              "+34 612 345 678",
		MustChangePassword: true,
		IsActive:           true,
		RoleID:             pgtype.UUID{Bytes: [16]byte(roleID), Valid: true},
		RoleName:           "Administrador",
		LastLoginAt:        pgtype.Timestamptz{},
		LastLoginIp:        pgtype.Text{},
		CreatedAt:          pgtype.Timestamptz{Time: created, Valid: true},
		UpdatedAt:          pgtype.Timestamptz{Time: updated, Valid: true},
	})

	if got.ID != id || got.RoleID != roleID || got.RoleName != "Administrador" {
		t.Fatalf("mapeo de ids/rol incorrecto: %+v", got)
	}
	if got.LastLoginAt != nil || got.LastLoginIP != nil {
		t.Fatalf("last_login nulos deberían ser nil: %v %v", got.LastLoginAt, got.LastLoginIP)
	}
	if !got.CreatedAt.Equal(created) || !got.UpdatedAt.Equal(updated) {
		t.Fatalf("fechas mal mapeadas: %+v", got)
	}
}

func TestMapUserRowsOtherVariants(t *testing.T) {
	id := uuid.New()
	roleID := uuid.New()
	at := time.Date(2026, 10, 4, 9, 30, 0, 0, time.UTC)
	ip := "10.0.0.7"

	byEmail := mapGetUserByEmailRow(gendb.GetUserByEmailRow{
		ID: pgUUID(id), Email: "ana@ejemplo.com", RoleID: pgUUID(roleID),
		LastLoginAt: pgtype.Timestamptz{Time: at, Valid: true},
		LastLoginIp: pgtype.Text{String: ip, Valid: true},
	})

	list := mapListUsersRow(gendb.ListUsersRow{
		ID: pgUUID(id), Email: "ana@ejemplo.com", RoleID: pgUUID(roleID),
		LastLoginAt: pgtype.Timestamptz{}, LastLoginIp: pgtype.Text{},
	})
	if list.LastLoginAt != nil || list.LastLoginIP != nil {
		t.Fatalf("fila de listado con last_login NULL: %+v", list)
	}

	insert := mapInsertUserRow(gendb.InsertUserRow{
		ID: pgUUID(id), Email: "ana@ejemplo.com", RoleID: pgUUID(roleID),
		LastLoginAt: pgtype.Timestamptz{}, LastLoginIp: pgtype.Text{},
	})
	if insert.RoleName != "" {
		t.Fatalf("InsertUser no trae RoleName: %q", insert.RoleName)
	}

	update := mapUpdateUserRow(gendb.UpdateUserRow{
		ID: pgUUID(id), Email: "ana@ejemplo.com", RoleID: pgUUID(roleID),
	})

	if byEmail.LastLoginAt == nil || !byEmail.LastLoginAt.Equal(at) {
		t.Fatalf("lastLoginAt por email = %v", byEmail.LastLoginAt)
	}
	if byEmail.LastLoginIP == nil || *byEmail.LastLoginIP != ip {
		t.Fatalf("lastLoginIp por email = %v", byEmail.LastLoginIP)
	}
	_ = update
}

func TestMapUserAuthNilPermissions(t *testing.T) {
	got := mapUserAuth(gendb.GetUserAuthByEmailRow{
		ID: pgUUID(uuid.New()), RoleID: pgUUID(uuid.New()),
		Permissions: nil,
	})
	if got.Permissions == nil || len(got.Permissions) != 0 {
		t.Fatalf("Permissions = %#v, se esperaba []", got.Permissions)
	}
}

func TestMapRoleAndPermission(t *testing.T) {
	id := uuid.New()
	created := time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC)

	plain := mapRole(gendb.Role{ID: pgUUID(id), Name: "Editor", CreatedAt: pgtype.Timestamptz{Time: created, Valid: true}})
	if plain.Name != "Editor" || plain.Permissions != nil {
		t.Fatalf("mapRole = %+v", plain)
	}

	detailed := mapRoleWithDetails(gendb.GetRoleByIDRow{
		ID: pgUUID(id), Name: "Editor", Permissions: nil, UserCount: 2,
		CreatedAt: pgtype.Timestamptz{Time: created, Valid: true},
	})
	if detailed.Permissions == nil || detailed.UserCount != 2 {
		t.Fatalf("mapRoleWithDetails = %+v", detailed)
	}

	listed := mapListRolesRow(gendb.ListRolesRow{ID: pgUUID(id), Name: "Editor", Permissions: []string{"eventos"}})
	if len(listed.Permissions) != 1 || listed.Permissions[0] != "eventos" {
		t.Fatalf("mapListRolesRow = %+v", listed)
	}

	perm := mapPermission(gendb.Permission{ID: pgUUID(id), Code: "eventos", Label: "Eventos"})
	if perm.Code != "eventos" || perm.Label != "Eventos" {
		t.Fatalf("mapPermission = %+v", perm)
	}
}

func TestMapAuditRows(t *testing.T) {
	id := uuid.New()
	when := time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC)

	// Intento no identificado: user_id NULL y sin datos de cuenta.
	anon := mapListLoginEventsRow(gendb.ListLoginEventsRow{
		ID: pgUUID(id), UserID: pgtype.UUID{}, Result: "failure", Ip: "192.168.1.9",
		UserEmail:     pgtype.Text{},
		UserFirstName: pgtype.Text{},
		UserLastName:  pgtype.Text{},
		CreatedAt:     pgtype.Timestamptz{Time: when, Valid: true},
	})
	if anon.UserID != nil || anon.UserEmail != nil || anon.UserName != nil {
		t.Fatalf("intento sin cuenta no debe traer datos derivados: %+v", anon)
	}

	identified := mapListLoginEventsRow(gendb.ListLoginEventsRow{
		ID: pgUUID(id), UserID: pgUUID(uuid.New()), Result: "success", Ip: "10.0.0.1",
		UserEmail:     pgtype.Text{String: "ana@ejemplo.com", Valid: true},
		UserFirstName: pgtype.Text{String: "Ana", Valid: true},
		UserLastName:  pgtype.Text{String: "Pérez", Valid: true},
		CreatedAt:     pgtype.Timestamptz{Time: when, Valid: true},
	})
	if identified.UserEmail == nil || identified.UserName == nil || *identified.UserName != "Ana Pérez" {
		t.Fatalf("datos derivados del intento identificado: %+v", identified)
	}

	inserted := mapInsertLoginEventRow(gendb.InsertLoginEventRow{
		ID: pgUUID(id), UserID: pgtype.UUID{}, Result: "failure", Ip: "10.0.0.2",
		CreatedAt: pgtype.Timestamptz{Time: when, Valid: true},
	})
	if inserted.UserID != nil || inserted.Result != "failure" {
		t.Fatalf("mapInsertLoginEventRow = %+v", inserted)
	}

	// Acción sin objetivo (creación rechazada) y sin etiqueta.
	action := mapInsertAdminActionRow(gendb.InsertAdminActionRow{
		ID: pgUUID(id), ActorUserID: pgtype.UUID{}, Action: "role.create",
		TargetKind: "role", TargetUserID: pgtype.UUID{}, TargetRoleID: pgtype.UUID{},
		TargetLabel: pgtype.Text{}, Result: "failure",
		CreatedAt: pgtype.Timestamptz{Time: when, Valid: true},
	})
	if action.TargetUserID != nil || action.TargetRoleID != nil || action.TargetLabel != nil {
		t.Fatalf("acción sin objetivo: %+v", action)
	}

	entry := mapListAdminActionsRow(gendb.ListAdminActionsRow{
		ID: pgUUID(id), ActorUserID: pgUUID(uuid.New()), Action: "user.create",
		ActorEmail:     pgtype.Text{String: "admin@ejemplo.com", Valid: true},
		ActorFirstName: pgtype.Text{String: "Ada", Valid: true},
		ActorLastName:  pgtype.Text{String: "Lovelace", Valid: true},
		TargetKind:     "user", TargetID: pgUUID(uuid.New()),
		TargetLabel: pgtype.Text{String: "ana@ejemplo.com", Valid: true},
		Result:      "success",
		CreatedAt:   pgtype.Timestamptz{Time: when, Valid: true},
	})
	if entry.ActorEmail == nil || entry.ActorName == nil || *entry.ActorName != "Ada Lovelace" {
		t.Fatalf("actor derivado: %+v", entry)
	}
	if entry.TargetID == nil || entry.TargetLabel == nil {
		t.Fatalf("objetivo aplanado: %+v", entry)
	}
}

func TestClassifyErrors(t *testing.T) {
	notFound := classify(pgx.ErrNoRows, "La cuenta no existe", "")
	assertKind(t, notFound, apperr.KindNotFound)

	unique := classify(&pgconn.PgError{Code: codeUniqueViolation}, "", "correo duplicado")
	assertKind(t, unique, apperr.KindConflict)

	fk := classify(&pgconn.PgError{Code: codeForeignKeyViolation}, "", "")
	assertKind(t, fk, apperr.KindInvalid)

	check := classify(&pgconn.PgError{Code: codeCheckViolation}, "", "")
	assertKind(t, check, apperr.KindInvalid)

	plain := errors.New("otro error")
	if got := classify(plain, "", ""); !errors.Is(got, plain) {
		t.Fatalf("classify(otro) = %v, se esperaba el original", got)
	}
}

func TestWrapPreservesCause(t *testing.T) {
	base := &pgconn.PgError{Code: codeUniqueViolation}
	err := wrap(base, "insert user", "", "correo duplicado")
	assertKind(t, err, apperr.KindConflict)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatal("wrap() debe conservar la causa con %w")
	}
	if err == nil || err.Error() == "" {
		t.Fatal("wrap() devolvió un error vacío")
	}
}

func TestWrapFKConflict(t *testing.T) {
	// Una violación de FK en un borrado restringido es un conflicto (409).
	err := wrapFKConflict(&pgconn.PgError{Code: codeForeignKeyViolation}, "delete role", "rol en uso")
	assertKind(t, err, apperr.KindConflict)

	// Cualquier otro error se traduce como classify.
	if got := wrapFKConflict(nil, "delete role", ""); got != nil {
		t.Fatalf("wrapFKConflict(nil) = %v", got)
	}
	notFound := wrapFKConflict(pgx.ErrNoRows, "delete role", "")
	assertKind(t, notFound, apperr.KindNotFound)
}

func assertKind(t *testing.T, err error, kind apperr.Kind) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("se esperaba *apperr.Error, se obtuvo %v", err)
	}
	if appErr.Kind != kind {
		t.Fatalf("kind = %v, se esperaba %v", appErr.Kind, kind)
	}
}

func strPtr(s string) *string { return &s }
