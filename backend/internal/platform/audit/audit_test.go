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
		"event.create", "event.update", "event.delete", "event.publish",
		"event.unpublish", "event.image.upload",
		"activity.create", "activity.update", "activity.delete", "activity.publish",
		"activity.unpublish", "activity.image.upload",
		"category.create", "category.update", "category.delete",
		"ministry.create", "ministry.update", "ministry.delete", "ministry.publish",
		"ministry.unpublish", "ministry.leader.show", "ministry.leader.hide",
		"ministry.image.upload",
		"donation.create", "donation.update", "donation.delete", "donation.publish",
		"donation.unpublish", "donation.text.update", "donation.text.publish",
		"donation.text.unpublish",
		"media.sync", "media.description.update", "media.hide", "media.show",
		"media.merge", "media.mark.resolve",
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
	for _, code := range []string{"", "user.delete", "role.password_reset", "USER.CREATE", "home.page.update", "event.archive", "activity.dup", "category.publish", "ministry.archive", "ministries.create", "ministry.leader.show.all", "donation.archive", "donations.create", "donation.text.create", "donation.image.upload", "media.archive", "media.create", "media.delete", "media.publish", "medias.sync", "media.mark.show"} {
		if ValidActionCode(code) {
			t.Errorf("ValidActionCode(%q) = true, se esperaba false", code)
		}
	}
}

// TestF4ActionCodesValidate fija el registro cerrado de F4 (FR-017/R4-10): los
// 15 códigos `event.*`, `activity.*` y `category.*` (los tres del catálogo de
// categorías incluidos) se registran sobre un objetivo de contenido
// (`target_kind = 'content'`), con `targetLabel` —nunca el cuerpo de la
// petición— y sin FK de usuario ni de rol. Los códigos de F2/F3 siguen
// intactos (regresión) y un código fuera de la lista falla.
func TestF4ActionCodesValidate(t *testing.T) {
	actor := uuid.New()

	f4Codes := []string{
		ActionEventCreate, ActionEventUpdate, ActionEventDelete, ActionEventPublish,
		ActionEventUnpublish, ActionEventImageUpload,
		ActionActivityCreate, ActionActivityUpdate, ActionActivityDelete,
		ActionActivityPublish, ActionActivityUnpublish, ActionActivityImageUpload,
		ActionCategoryCreate, ActionCategoryUpdate, ActionCategoryDelete,
	}
	if len(f4Codes) != 15 {
		t.Fatalf("se esperaban 15 códigos de F4, hay %d", len(f4Codes))
	}

	labels := map[string]string{
		ActionEventCreate:         "Eventos · Noche de aposentos",
		ActionEventUpdate:         "Eventos · Noche de aposentos",
		ActionEventDelete:         "Eventos · Noche de aposentos",
		ActionEventPublish:        "Eventos · Noche de aposentos",
		ActionEventUnpublish:      "Eventos · Noche de aposentos",
		ActionEventImageUpload:    "Eventos · Imagen · img_11111111-1111-1111-1111-111111111111.jpg",
		ActionActivityCreate:      "Actividades · Repartir comida",
		ActionActivityUpdate:      "Actividades · Repartir comida",
		ActionActivityDelete:      "Actividades · Repartir comida",
		ActionActivityPublish:     "Actividades · Repartir comida",
		ActionActivityUnpublish:   "Actividades · Repartir comida",
		ActionActivityImageUpload: "Actividades · Imagen · img_22222222-2222-2222-2222-222222222222.png",
		ActionCategoryCreate:      "Categorías · Acción social",
		ActionCategoryUpdate:      "Categorías · Acción social",
		ActionCategoryDelete:      "Categorías · Acción social",
	}

	for _, code := range f4Codes {
		t.Run(code, func(t *testing.T) {
			action := Action{
				ActorUserID: &actor,
				Code:        code,
				TargetKind:  TargetContent,
				TargetLabel: labels[code],
				Result:      ResultSuccess,
			}
			if err := action.Validate(); err != nil {
				t.Fatalf("Validate() = %v, se esperaba nil", err)
			}
			if !ValidActionCode(code) {
				t.Fatalf("ValidActionCode(%q) = false", code)
			}
		})
	}

	// Los códigos de F2 y F3 siguen intactos (regresión del registro cerrado).
	for _, code := range []string{ActionUserCreate, ActionUserPasswordReset, ActionRoleDelete,
		ActionHomeIdentityUpdate, ActionHomeImageUpload, ActionHomePublish, ActionHomeUnpublish} {
		if !ValidActionCode(code) {
			t.Errorf("ValidActionCode(%q) = false: el código previo se perdió", code)
		}
	}

	// Fuera del registro: los códigos nuevos son los de F4 y solo esos.
	for _, code := range []string{"event.archive", "activity.schedule", "category.publish", "events.create"} {
		action := Action{ActorUserID: &actor, Code: code, TargetKind: TargetContent, TargetLabel: "Eventos · X", Result: ResultSuccess}
		if err := action.Validate(); !errors.Is(err, ErrUnknownActionCode) {
			t.Fatalf("Validate(%q) = %v, se esperaba %v", code, err, ErrUnknownActionCode)
		}
	}
}

// TestF6ActionCodesValidate fija el registro cerrado de F6 (FR-017/R6-12): los 8
// códigos `ministry.*` —incluidos los dos de la visibilidad de la ficha del
// encargado (`leader.show`/`leader.hide`)— se registran sobre un objetivo de
// contenido (`target_kind = 'content'`), con `targetLabel` —nunca el cuerpo de
// la petición— y sin FK de usuario ni de rol. Los códigos de F2/F3/F4 siguen
// intactos (regresión) y un código fuera de la lista falla.
func TestF6ActionCodesValidate(t *testing.T) {
	actor := uuid.New()

	f6Codes := []string{
		ActionMinistryCreate, ActionMinistryUpdate, ActionMinistryDelete,
		ActionMinistryPublish, ActionMinistryUnpublish,
		ActionMinistryLeaderShow, ActionMinistryLeaderHide,
		ActionMinistryImageUpload,
	}
	if len(f6Codes) != 8 {
		t.Fatalf("se esperaban 8 códigos de F6, hay %d", len(f6Codes))
	}

	labels := map[string]string{
		ActionMinistryCreate:      "Ministerios · Alabanza",
		ActionMinistryUpdate:      "Ministerios · Alabanza",
		ActionMinistryDelete:      "Ministerios · Alabanza",
		ActionMinistryPublish:     "Ministerios · Alabanza",
		ActionMinistryUnpublish:   "Ministerios · Alabanza",
		ActionMinistryLeaderShow:  "Ministerios · Alabanza",
		ActionMinistryLeaderHide:  "Ministerios · Alabanza",
		ActionMinistryImageUpload: "Ministerios · Imagen · img_33333333-3333-3333-3333-333333333333.webp",
	}

	for _, code := range f6Codes {
		t.Run(code, func(t *testing.T) {
			action := Action{
				ActorUserID: &actor,
				Code:        code,
				TargetKind:  TargetContent,
				TargetLabel: labels[code],
				Result:      ResultSuccess,
			}
			if err := action.Validate(); err != nil {
				t.Fatalf("Validate() = %v, se esperaba nil", err)
			}
			if !ValidActionCode(code) {
				t.Fatalf("ValidActionCode(%q) = false", code)
			}
		})
	}

	// Los códigos de F2, F3 y F4 siguen intactos (regresión del registro
	// cerrado: F6 solo añade).
	for _, code := range []string{ActionUserCreate, ActionUserPasswordReset, ActionRoleDelete,
		ActionHomeIdentityUpdate, ActionHomeImageUpload, ActionHomePublish, ActionHomeUnpublish,
		ActionEventCreate, ActionEventImageUpload, ActionActivityPublish,
		ActionCategoryCreate, ActionCategoryDelete} {
		if !ValidActionCode(code) {
			t.Errorf("ValidActionCode(%q) = false: el código previo se perdió", code)
		}
	}

	// Fuera del registro: el prefijo no basta, el código tiene que existir.
	for _, code := range []string{"ministry.archive", "ministry.leader.rotate", "ministries.create"} {
		action := Action{ActorUserID: &actor, Code: code, TargetKind: TargetContent, TargetLabel: "Ministerios · X", Result: ResultSuccess}
		if err := action.Validate(); !errors.Is(err, ErrUnknownActionCode) {
			t.Fatalf("Validate(%q) = %v, se esperaba %v", code, err, ErrUnknownActionCode)
		}
	}
}

// TestF7ActionCodesValidate fija el registro cerrado de F7 (FR-017/R7-10): los 8
// códigos `donation.*` —incluidos los tres de los textos
// (`text.update`/`text.publish`/`text.unpublish`)— se registran sobre un
// objetivo de contenido (`target_kind = 'content'`), con `targetLabel` —nunca el
// cuerpo de la petición— y sin FK de usuario ni de rol. Los códigos de
// F2/F3/F4/F6 siguen intactos (regresión) y un código fuera de la lista falla.
func TestF7ActionCodesValidate(t *testing.T) {
	actor := uuid.New()

	f7Codes := []string{
		ActionDonationCreate, ActionDonationUpdate, ActionDonationDelete,
		ActionDonationPublish, ActionDonationUnpublish,
		ActionDonationTextUpdate, ActionDonationTextPublish, ActionDonationTextUnpublish,
	}
	if len(f7Codes) != 8 {
		t.Fatalf("se esperaban 8 códigos de F7, hay %d", len(f7Codes))
	}

	labels := map[string]string{
		ActionDonationCreate:        "Donaciones · BAC San José · CR05015202001234567890",
		ActionDonationUpdate:        "Donaciones · BAC San José · CR05015202001234567890",
		ActionDonationDelete:        "Donaciones · BAC San José · CR05015202001234567890",
		ActionDonationPublish:       "Donaciones · BAC San José · CR05015202001234567890",
		ActionDonationUnpublish:     "Donaciones · BAC San José · CR05015202001234567890",
		ActionDonationTextUpdate:    "Donaciones · Texto introductorio",
		ActionDonationTextPublish:   "Donaciones · Texto de uso",
		ActionDonationTextUnpublish: "Donaciones · Texto de uso",
	}

	for _, code := range f7Codes {
		t.Run(code, func(t *testing.T) {
			action := Action{
				ActorUserID: &actor,
				Code:        code,
				TargetKind:  TargetContent,
				TargetLabel: labels[code],
				Result:      ResultSuccess,
			}
			if err := action.Validate(); err != nil {
				t.Fatalf("Validate() = %v, se esperaba nil", err)
			}
			if !ValidActionCode(code) {
				t.Fatalf("ValidActionCode(%q) = false", code)
			}
		})
	}

	// Los códigos de F2, F3, F4 y F6 siguen intactos (regresión del registro
	// cerrado: F7 solo añade).
	for _, code := range []string{ActionUserCreate, ActionUserPasswordReset, ActionRoleDelete,
		ActionHomeIdentityUpdate, ActionHomeImageUpload, ActionHomePublish, ActionHomeUnpublish,
		ActionEventCreate, ActionEventImageUpload, ActionActivityPublish,
		ActionCategoryCreate, ActionCategoryDelete,
		ActionMinistryCreate, ActionMinistryLeaderShow, ActionMinistryImageUpload} {
		if !ValidActionCode(code) {
			t.Errorf("ValidActionCode(%q) = false: el código previo se perdió", code)
		}
	}

	// Fuera del registro: los códigos nuevos son los de F7 y solo esos.
	for _, code := range []string{"donation.archive", "donations.create", "donation.text.create", "donation.image.upload"} {
		action := Action{ActorUserID: &actor, Code: code, TargetKind: TargetContent, TargetLabel: "Donaciones · X", Result: ResultSuccess}
		if err := action.Validate(); !errors.Is(err, ErrUnknownActionCode) {
			t.Fatalf("Validate(%q) = %v, se esperaba %v", code, err, ErrUnknownActionCode)
		}
	}
}

// TestF9ActionCodesValidate fija el registro cerrado de F9 (FR-022/P9-14): los
// 6 códigos `media.*` —la búsqueda, la descripción corta, ocultar/mostrar, la
// unión y la resolución de marcas— se registran sobre un objetivo de contenido
// (`target_kind = 'content'`), con `targetLabel` —nunca el cuerpo de la
// petición— y sin FK de usuario ni de rol. F9 **no** tiene `create`/`delete`
// (no hay alta manual ni borrado: FR-019). Los códigos de F2/F3/F4/F6/F7 siguen
// intactos (regresión) y un código fuera de la lista falla.
func TestF9ActionCodesValidate(t *testing.T) {
	actor := uuid.New()

	f9Codes := []string{
		ActionMediaSync, ActionMediaDescriptionUpdate, ActionMediaHide,
		ActionMediaShow, ActionMediaMerge, ActionMediaMarkResolve,
	}
	if len(f9Codes) != 6 {
		t.Fatalf("se esperaban 6 códigos de F9, hay %d", len(f9Codes))
	}

	labels := map[string]string{
		ActionMediaSync:              "Medios · Búsqueda · 3 entraron, 1 unió, 2 marcaron",
		ActionMediaDescriptionUpdate: "Medios · Potencial sin Resultados La Lección de la Higuera",
		ActionMediaHide:              "Medios · Potencial sin Resultados La Lección de la Higuera",
		ActionMediaShow:              "Medios · Potencial sin Resultados La Lección de la Higuera",
		ActionMediaMerge:             "Medios · Potencial sin Resultados La Lección de la Higuera",
		ActionMediaMarkResolve:       "Medios · Potencial sin Resultados La Lección de la Higuera · possible_duplicate · visible",
	}

	for _, code := range f9Codes {
		t.Run(code, func(t *testing.T) {
			action := Action{
				ActorUserID: &actor,
				Code:        code,
				TargetKind:  TargetContent,
				TargetLabel: labels[code],
				Result:      ResultSuccess,
			}
			if err := action.Validate(); err != nil {
				t.Fatalf("Validate() = %v, se esperaba nil", err)
			}
			if !ValidActionCode(code) {
				t.Fatalf("ValidActionCode(%q) = false", code)
			}
		})
	}

	// Los códigos de F2, F3, F4, F6 y F7 siguen intactos (regresión del
	// registro cerrado: F9 solo añade).
	for _, code := range []string{ActionUserCreate, ActionUserPasswordReset, ActionRoleDelete,
		ActionHomeIdentityUpdate, ActionHomeImageUpload, ActionHomePublish, ActionHomeUnpublish,
		ActionEventCreate, ActionEventImageUpload, ActionActivityPublish,
		ActionCategoryCreate, ActionCategoryDelete,
		ActionMinistryCreate, ActionMinistryLeaderShow, ActionMinistryImageUpload,
		ActionDonationCreate, ActionDonationTextUpdate, ActionDonationTextUnpublish} {
		if !ValidActionCode(code) {
			t.Errorf("ValidActionCode(%q) = false: el código previo se perdió", code)
		}
	}

	// Fuera del registro: los códigos nuevos son los de F9 y solo esos.
	for _, code := range []string{"media.archive", "media.create", "media.delete", "media.publish", "medias.sync", "media.mark.show"} {
		action := Action{ActorUserID: &actor, Code: code, TargetKind: TargetContent, TargetLabel: "Medios · X", Result: ResultSuccess}
		if err := action.Validate(); !errors.Is(err, ErrUnknownActionCode) {
			t.Fatalf("Validate(%q) = %v, se esperaba %v", code, err, ErrUnknownActionCode)
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
