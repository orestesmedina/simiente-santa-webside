package audit

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestActionCodesRegistry(t *testing.T) {
	want := []string{
		"user.create", "user.update", "user.activate", "user.deactivate",
		"user.password_reset", "role.create", "role.update", "role.delete",
		"home.identity.update", "home.about.update", "home.contact.update",
		"home.schedule.create", "home.schedule.update", "home.schedule.delete",
		"home.whatsapp.create", "home.whatsapp.update", "home.whatsapp.delete",
		"home.social.create", "home.social.update", "home.social.delete",
		"home.image.upload", "home.publish", "home.unpublish",
	}
	if len(ActionCodes) != len(want) {
		t.Fatalf("ActionCodes tiene %d códigos, se esperaban %d", len(ActionCodes), len(want))
	}
	for i, code := range want {
		if ActionCodes[i] != code {
			t.Errorf("ActionCodes[%d] = %q, se esperaba %q", i, ActionCodes[i], code)
		}
		if !ValidActionCode(code) {
			t.Errorf("ValidActionCode(%q) = false, se esperaba true", code)
		}
	}
	for _, code := range []string{"", "user.delete", "role.password_reset", "USER.CREATE", "home.page.update"} {
		if ValidActionCode(code) {
			t.Errorf("ValidActionCode(%q) = true, se esperaba false", code)
		}
	}
}

// TestTargetKindsRegistry fija el registro cerrado de tipos de objetivo
// (FR-023/R3-11): los de F2 más el `content` de F3.
func TestTargetKindsRegistry(t *testing.T) {
	want := []TargetKind{TargetUser, TargetRole, TargetContent}
	for _, kind := range want {
		if string(kind) == "" {
			t.Errorf("tipo de objetivo vacío")
		}
	}
	if TargetContent != "content" {
		t.Errorf("TargetContent = %q, se esperaba content", TargetContent)
	}
}

func TestActionResultValid(t *testing.T) {
	tests := []struct {
		result Result
		want   bool
	}{
		{ResultSuccess, true},
		{ResultFailure, true},
		{ResultDenied, true},
		{Result(""), false},
		{Result("ok"), false},
	}
	for _, tt := range tests {
		if got := tt.result.Valid(); got != tt.want {
			t.Errorf("Result(%q).Valid() = %v, se esperaba %v", tt.result, got, tt.want)
		}
	}
}

func TestActionValidate(t *testing.T) {
	actor := uuid.New()
	target := uuid.New()
	role := uuid.New()

	validUser := Action{
		ActorUserID:  &actor,
		Code:         ActionUserUpdate,
		TargetKind:   TargetUser,
		TargetUserID: &target,
		TargetLabel:  "Ana Pérez",
		Result:       ResultSuccess,
	}
	validRole := Action{
		ActorUserID:  &actor,
		Code:         ActionRoleCreate,
		TargetKind:   TargetRole,
		TargetRoleID: &role,
		TargetLabel:  "Coordinación",
		Result:       ResultSuccess,
	}
	// La inicialización no tiene actor y es la única acción sin él (FR-007).
	validInit := Action{
		Code:         ActionUserCreate,
		TargetKind:   TargetUser,
		TargetUserID: &target,
		TargetLabel:  "Administrador",
		Result:       ResultSuccess,
	}
	validContent := Action{
		ActorUserID: &actor,
		Code:        ActionHomeIdentityUpdate,
		TargetKind:  TargetContent,
		TargetLabel: "Portada · Identidad",
		Result:      ResultSuccess,
	}

	tests := []struct {
		name    string
		action  Action
		wantErr error
	}{
		{name: "usuario válido", action: validUser, wantErr: nil},
		{name: "rol válido", action: validRole, wantErr: nil},
		{name: "inicialización sin actor", action: validInit, wantErr: nil},
		{name: "objetivo de contenido válido", action: validContent, wantErr: nil},
		{name: "resultado denegado", action: Action{ActorUserID: &actor, Code: ActionUserCreate, TargetKind: TargetUser, Result: ResultDenied}, wantErr: nil},
		{
			name:    "código desconocido",
			action:  Action{ActorUserID: &actor, Code: "user.delete", TargetKind: TargetUser, Result: ResultSuccess},
			wantErr: ErrUnknownActionCode,
		},
		{
			name:    "resultado inválido",
			action:  Action{ActorUserID: &actor, Code: ActionUserCreate, TargetKind: TargetUser, Result: Result("ok")},
			wantErr: ErrInvalidResult,
		},
		{
			name:    "tipo de objetivo desconocido",
			action:  Action{ActorUserID: &actor, Code: ActionUserCreate, TargetKind: TargetKind("grupo"), Result: ResultSuccess},
			wantErr: ErrUnknownTargetKind,
		},
		{
			name:    "usuario con target_role_id",
			action:  Action{ActorUserID: &actor, Code: ActionUserUpdate, TargetKind: TargetUser, TargetRoleID: &role, Result: ResultSuccess},
			wantErr: ErrTargetMismatch,
		},
		{
			name:    "rol con target_user_id",
			action:  Action{ActorUserID: &actor, Code: ActionRoleUpdate, TargetKind: TargetRole, TargetUserID: &target, Result: ResultSuccess},
			wantErr: ErrTargetMismatch,
		},
		{
			name:    "contenido con target_user_id",
			action:  Action{ActorUserID: &actor, Code: ActionHomeAboutUpdate, TargetKind: TargetContent, TargetUserID: &target, TargetLabel: "Portada · Quiénes somos", Result: ResultSuccess},
			wantErr: ErrTargetMismatch,
		},
		{
			name:    "contenido con target_role_id",
			action:  Action{ActorUserID: &actor, Code: ActionHomeContactUpdate, TargetKind: TargetContent, TargetRoleID: &role, TargetLabel: "Portada · Contacto", Result: ResultSuccess},
			wantErr: ErrTargetMismatch,
		},
		{
			name:    "contenido sin etiqueta",
			action:  Action{ActorUserID: &actor, Code: ActionHomeImageUpload, TargetKind: TargetContent, Result: ResultSuccess},
			wantErr: ErrMissingTargetLabel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, se esperaba nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() = %v, se esperaba %v", err, tt.wantErr)
			}
		})
	}
}

func TestEventValidate(t *testing.T) {
	user := uuid.New()

	tests := []struct {
		name    string
		event   Event
		wantErr error
	}{
		{name: "éxito con cuenta", event: Event{UserID: &user, Result: ResultSuccess, IP: "127.0.0.1"}, wantErr: nil},
		{name: "fallo sin cuenta", event: Event{Result: ResultFailure, IP: "10.0.0.9"}, wantErr: nil},
		{name: "denied no es un intento", event: Event{Result: ResultDenied, IP: "10.0.0.9"}, wantErr: ErrInvalidResult},
		{name: "resultado desconocido", event: Event{Result: Result("ok"), IP: "10.0.0.9"}, wantErr: ErrInvalidResult},
		{name: "sin IP", event: Event{Result: ResultSuccess}, wantErr: ErrMissingIP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, se esperaba nil", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() = %v, se esperaba %v", err, tt.wantErr)
			}
		})
	}
}

// recordingFake verifica en compilación que un dominio puede implementar
// Recorder sin que platform/audit sepa de él.
type recordingFake struct {
	denials []Denial
}

func (f *recordingFake) RecordDenied(_ context.Context, denial Denial) error {
	f.denials = append(f.denials, denial)
	return nil
}

var _ Recorder = (*recordingFake)(nil)

func TestRecorderContract(t *testing.T) {
	actor := uuid.New()
	var recorder Recorder = &recordingFake{}

	if err := recorder.RecordDenied(context.Background(), Denial{ActorUserID: &actor, Method: "POST", Path: "/api/v1/admin/usuarios"}); err != nil {
		t.Fatalf("RecordDenied() = %v, se esperaba nil", err)
	}
}
