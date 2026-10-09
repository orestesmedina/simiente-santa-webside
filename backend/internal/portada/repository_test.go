package portada

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/audit"
)

// text devuelve un pgtype.Text válido (columna con valor).
func text(s string) pgtype.Text { return pgtype.Text{String: s, Valid: true} }

// TestMapHomeIdentity verifica el mapeo pgtype ↔ dominio del singleton de
// identidad: los campos anulables NULL llegan como nil y los presentes como
// puntero, y el estado se traduce a PublicationState (§8.1.6).
func TestMapHomeIdentity(t *testing.T) {
	id := uuid.New()
	row := gendb.HomeIdentity{
		ID:               pgtype.UUID{Bytes: [16]byte(id), Valid: true},
		Singleton:        true,
		NameEs:           "Iglesia Simiente",
		NameEn:           text("Seed Church"),
		TaglineEs:        pgtype.Text{}, // NULL
		LogoFile:         text("img_0123456789abcdef0123456789abcdef0123.png"),
		CoverImageFile:   pgtype.Text{}, // NULL
		PublicationState: "published",
		CreatedAt:        pgtype.Timestamptz{Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Valid: true},
	}

	got := mapHomeIdentity(row)

	if got.ID != id {
		t.Errorf("ID = %v, se esperaba %v", got.ID, id)
	}
	if got.NameEs != "Iglesia Simiente" {
		t.Errorf("NameEs = %q", got.NameEs)
	}
	if got.NameEn == nil || *got.NameEn != "Seed Church" {
		t.Errorf("NameEn = %v, se esperaba puntero a Seed Church", got.NameEn)
	}
	if got.TaglineEs != nil {
		t.Errorf("TaglineEs = %v, se esperaba nil para NULL", got.TaglineEs)
	}
	if got.LogoFile == nil || *got.LogoFile != "img_0123456789abcdef0123456789abcdef0123.png" {
		t.Errorf("LogoFile = %v", got.LogoFile)
	}
	if got.CoverImageFile != nil {
		t.Errorf("CoverImageFile = %v, se esperaba nil para NULL", got.CoverImageFile)
	}
	if got.PublicationState != StatePublished {
		t.Errorf("PublicationState = %q, se esperaba %q", got.PublicationState, StatePublished)
	}
}

func TestMapHomeSingletons(t *testing.T) {
	about := mapHomeAbout(gendb.HomeAbout{
		ID:               pgtype.UUID{Bytes: [16]byte(uuid.New()), Valid: true},
		TextEs:           "Somos una comunidad",
		TextEn:           pgtype.Text{}, // NULL
		PublicationState: "published",
	})
	if about.TextEs != "Somos una comunidad" || about.TextEn != nil || about.PublicationState != StatePublished {
		t.Errorf("About = %+v", about)
	}

	contact := mapHomeContact(gendb.HomeContact{
		ID:        pgtype.UUID{Bytes: [16]byte(uuid.New()), Valid: true},
		AddressEs: "Calle 1",
		AddressEn: text("1st Street"),
		Email:     "info@simiente.org",
		Phone:     "+34612345678",
	})
	if contact.Email != "info@simiente.org" || contact.Phone != "+34612345678" {
		t.Errorf("Contact = %+v", contact)
	}
	if contact.AddressEn == nil || *contact.AddressEn != "1st Street" {
		t.Errorf("AddressEn = %v", contact.AddressEn)
	}
}

// TestUUIDHelpers verifica la traducción de UUID anulables.
func TestUUIDHelpers(t *testing.T) {
	id := uuid.New()
	if got := uuidValue(pgtype.UUID{Bytes: [16]byte(id), Valid: true}); got != id {
		t.Errorf("uuidValue = %v, se esperaba %v", got, id)
	}
	if got := uuidPtr(pgtype.UUID{}); got != nil {
		t.Errorf("uuidPtr(NULL) = %v, se esperaba nil", got)
	}
	if got := uuidPtr(pgtype.UUID{Bytes: [16]byte(id), Valid: true}); got == nil || *got != id {
		t.Errorf("uuidPtr(válido) = %v", got)
	}
	if got := pgUUID(id); !got.Valid || uuid.UUID(got.Bytes) != id {
		t.Errorf("pgUUID = %+v", got)
	}
	if got := pgUUIDPtr(nil); got.Valid {
		t.Errorf("pgUUIDPtr(nil) = %+v, se esperaba NULL", got)
	}
	if got := pgUUIDPtr(&id); !got.Valid {
		t.Errorf("pgUUIDPtr(válido) = %+v", got)
	}
}

// TestMapHomeCollections verifica los mapeos de servicio, canal y red.
func TestMapHomeCollections(t *testing.T) {
	service := mapHomeService(gendb.HomeService{
		ID:               pgtype.UUID{Bytes: [16]byte(uuid.New()), Valid: true},
		DayOfWeek:        0,
		StartTime:        "10:00",
		EndTime:          text("12:00"),
		NameEs:           "Culto",
		PlaceEs:          "Templo",
		PublicationState: "draft",
		SortOrder:        3,
	})
	if service.DayOfWeek != 0 || service.StartTime != "10:00" {
		t.Errorf("Service = %+v", service)
	}
	if service.EndTime == nil || *service.EndTime != "12:00" {
		t.Errorf("EndTime = %v, se esperaba 12:00", service.EndTime)
	}
	if service.SortOrder != 3 || service.PublicationState != StateDraft {
		t.Errorf("Service = %+v", service)
	}

	channel := mapHomeWhatsappChannel(gendb.HomeWhatsappChannel{
		ID:          pgtype.UUID{Bytes: [16]byte(uuid.New()), Valid: true},
		Kind:        KindGroup,
		Destination: "https://chat.whatsapp.com/abc",
		NameEs:      "Grupo jóvenes",
	})
	if channel.Kind != KindGroup || channel.Destination != "https://chat.whatsapp.com/abc" {
		t.Errorf("WhatsappChannel = %+v", channel)
	}

	link := mapHomeSocialLink(gendb.HomeSocialLink{
		ID:      pgtype.UUID{Bytes: [16]byte(uuid.New()), Valid: true},
		Network: "facebook",
		Url:     "https://facebook.com/simiente",
	})
	if link.Network != "facebook" || link.URL != "https://facebook.com/simiente" {
		t.Errorf("SocialLink = %+v", link)
	}
}

// TestPublicationStateValid fija el registro cerrado de estados (FR-013).
func TestPublicationStateValid(t *testing.T) {
	tests := []struct {
		state PublicationState
		want  bool
	}{
		{StateDraft, true},
		{StatePublished, true},
		{PublicationState("archived"), false},
		{PublicationState(""), false},
	}
	for _, tt := range tests {
		if got := tt.state.Valid(); got != tt.want {
			t.Errorf("PublicationState(%q).Valid() = %v, se esperaba %v", tt.state, got, tt.want)
		}
	}
}

// TestPGTextHelpers verifica la traducción de strings opcionales: "" se guarda
// como NULL (analyze I6) y un valor no vacío como TEXT válido.
func TestPGTextHelpers(t *testing.T) {
	if v := pgText(""); v.Valid {
		t.Errorf("pgText(\"\") = %+v, se esperaba NULL", v)
	}
	if v := pgText("hola"); !v.Valid || v.String != "hola" {
		t.Errorf("pgText(\"hola\") = %+v", v)
	}
	if v := pgTextPtr(nil); v.Valid {
		t.Errorf("pgTextPtr(nil) = %+v, se esperaba NULL", v)
	}
	if v := pgTextPtr(textOf("x")); !v.Valid || v.String != "x" {
		t.Errorf("pgTextPtr = %+v", v)
	}
	if v := pgLabel(""); v.Valid {
		t.Errorf("pgLabel(\"\") = %+v, se esperaba NULL", v)
	}
	if v := textPtr(pgtype.Text{}); v != nil {
		t.Errorf("textPtr(NULL) = %v, se esperaba nil", v)
	}
}

// textOf es un atajo para construir un *string en las pruebas.
func textOf(s string) *string { return &s }

// TestAuditParamsMapsContentAction verifica el mapeo audit.Action →
// InsertAdminActionParams (R3-11): un objetivo `content` no lleva FK y conserva
// la etiqueta; una etiqueta vacía se guarda como NULL.
func TestAuditParamsMapsContentAction(t *testing.T) {
	actor := uuid.New()
	params := auditParams(audit.Action{
		ActorUserID: &actor,
		Code:        audit.ActionHomeIdentityUpdate,
		TargetKind:  audit.TargetContent,
		TargetLabel: "Portada · Identidad",
		Result:      audit.ResultSuccess,
	})

	if !params.ActorUserID.Valid || uuid.UUID(params.ActorUserID.Bytes) != actor {
		t.Errorf("ActorUserID = %+v", params.ActorUserID)
	}
	if params.Action != audit.ActionHomeIdentityUpdate {
		t.Errorf("Action = %q", params.Action)
	}
	if params.TargetKind != "content" {
		t.Errorf("TargetKind = %q, se esperaba content", params.TargetKind)
	}
	if params.TargetUserID.Valid || params.TargetRoleID.Valid {
		t.Errorf("las FK de contenido deben ir NULL: %+v / %+v", params.TargetUserID, params.TargetRoleID)
	}
	if !params.TargetLabel.Valid || params.TargetLabel.String != "Portada · Identidad" {
		t.Errorf("TargetLabel = %+v", params.TargetLabel)
	}
	if params.Result != "success" {
		t.Errorf("Result = %q", params.Result)
	}

	// Etiqueta vacía (no es el caso de `content`, pero el mapeo es genérico).
	empty := auditParams(audit.Action{TargetKind: audit.TargetContent})
	if empty.TargetLabel.Valid {
		t.Errorf("TargetLabel vacío = %+v, se esperaba NULL", empty.TargetLabel)
	}
}
