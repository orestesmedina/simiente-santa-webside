package ministerios

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/testutil"
)

// detailsOf devuelve los detalles del error de dominio de validación (falla si no
// es un error de dominio invalid).
func detailsOf(t *testing.T, err error) map[string]any {
	t.Helper()
	if err == nil {
		t.Fatal("se esperaba un error de dominio, llegó nil")
	}
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("se esperaba un error de dominio, llegó %v", err)
	}
	if domainErr.Kind != apperr.KindInvalid {
		t.Fatalf("kind = %v, se esperaba invalid (%v)", domainErr.Kind, err)
	}
	return domainErr.Details
}

// TestLimitsAreTheSingleSourceOfTruth fija los límites de negocio revisables
// (T612): los usan T613–T615 y los cita el contrato.
func TestLimitsAreTheSingleSourceOfTruth(t *testing.T) {
	limits := map[string]int{
		"MaxNameLength":        MaxNameLength,
		"MaxDescriptionLength": MaxDescriptionLength,
		"MaxLeaderNameLength":  MaxLeaderNameLength,
		"MinLeaderPhoneLength": MinLeaderPhoneLength,
		"MaxLeaderPhoneLength": MaxLeaderPhoneLength,
		"MinLeaderEmailLength": MinLeaderEmailLength,
		"MaxLeaderEmailLength": MaxLeaderEmailLength,
		"MaxScheduleLength":    MaxScheduleLength,
		"MaxPlaceLength":       MaxPlaceLength,
		"MaxURLLength":         MaxURLLength,
		"MaxImageAltLength":    MaxImageAltLength,
		"MinSortOrder":         MinSortOrder,
		"MaxSortOrder":         MaxSortOrder,
		"maxPublicItems":       maxPublicItems,
		"maxPortadaItems":      maxPortadaItems,
		"maxTargetLabel":       maxTargetLabel,
	}
	want := map[string]int{
		"MaxNameLength":        160,
		"MaxDescriptionLength": 2000,
		"MaxLeaderNameLength":  160,
		"MinLeaderPhoneLength": 7,
		"MaxLeaderPhoneLength": 32,
		"MinLeaderEmailLength": 3,
		"MaxLeaderEmailLength": 254,
		"MaxScheduleLength":    200,
		"MaxPlaceLength":       200,
		"MaxURLLength":         500,
		"MaxImageAltLength":    300,
		"MinSortOrder":         -2147483648,
		"MaxSortOrder":         2147483647,
		"maxPublicItems":       200,
		"maxPortadaItems":      5,
		"maxTargetLabel":       254,
	}
	for name, got := range limits {
		if got != want[name] {
			t.Errorf("%s = %d, se esperaba %d", name, got, want[name])
		}
	}
}

// TestNormalizeTextKeepsContentAndCollapsesSpaces fija la normalización de
// textos (analyze I6/A6-9): trim, espacios internos colapsados, saltos de línea
// conservados y todo el contenido intacto (tildes, ñ, ¿¡ y emojis).
func TestNormalizeTextKeepsContentAndCollapsesSpaces(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "trim", input: "  Alabanza  ", want: "Alabanza"},
		{name: "espacios internos", input: "Grupo   de\talabanza", want: "Grupo de alabanza"},
		{name: "saltos de línea intactos", input: "Primera línea\nSegunda línea", want: "Primera línea\nSegunda línea"},
		{name: "tildes, ñ y signos", input: "  ¿Quiénes somos? ¡Los niños!  ", want: "¿Quiénes somos? ¡Los niños!"},
		{name: "emojis", input: "Alabanza 🙌🏽", want: "Alabanza 🙌🏽"},
		{name: "solo espacios", input: "   ", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeText(tt.input); got != tt.want {
				t.Errorf("normalizeText(%q) = %q, se esperaba %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestNormalizeOptionalTurnsBlankIntoNil fija que `""`/espacios es "sin dato"
// (NULL) y nunca la cadena vacía que los CHECK de longitud rechazan.
func TestNormalizeOptionalTurnsBlankIntoNil(t *testing.T) {
	for _, input := range []string{"", "   ", "\t "} {
		if got := normalizeOptional(input); got != nil {
			t.Errorf("normalizeOptional(%q) = %q, se esperaba nil", input, *got)
		}
	}
	if got := normalizeOptional("  Todos los viernes  "); got == nil || *got != "Todos los viernes" {
		t.Errorf("normalizeOptional(válido) = %v", got)
	}
}

// TestNormalizeLang fija el idioma por defecto y el 400 invalid con details.lang
// (FR-008): es la única implementación (la usan T613 y T614).
func TestNormalizeLang(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "vacío es el idioma base", input: "", want: LangES},
		{name: "es", input: "es", want: LangES},
		{name: "EN en mayúsculas", input: " EN ", want: LangEN},
		{name: "fuera del registro", input: "fr", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeLang(tt.input)
			if tt.wantErr {
				details := detailsOf(t, err)
				if _, ok := details["lang"]; !ok {
					t.Fatalf("details = %v, se esperaba la clave lang", details)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeLang(%q) = %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("normalizeLang(%q) = %q, se esperaba %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestPublicationStateOrDefaultsToDraft fija el default del contrato (FR-012):
// el estado vacío es `draft` («se publica al guardar» es una decisión explícita).
func TestPublicationStateOrDefaultsToDraft(t *testing.T) {
	if got := publicationStateOr(""); got != StateDraft {
		t.Errorf("publicationStateOr(\"\") = %q, se esperaba draft", got)
	}
	if got := publicationStateOr("published"); got != StatePublished {
		t.Errorf("publicationStateOr(published) = %q", got)
	}
}

// TestMinistryFromInputNormalizesAndAppliesDefaults fija la traducción del DTO
// del contrato a la entidad (T612) y «Oculto por defecto» (P6-3): sin
// `leaderVisible` en el DTO, el ministerio nace con la ficha del encargado NO
// publicada aunque traiga nombre y contacto.
func TestMinistryFromInputNormalizesAndAppliesDefaults(t *testing.T) {
	ministry, err := ministryFromInput(MinisterioInput{
		NameEs:           "  Alabanza  ",
		DescriptionEs:    "Grupo   de alabanza",
		DescriptionEn:    "   ",
		LeaderNameEs:     "Ana Pérez",
		LeaderPhone:      " +58 412 000 0000 ",
		ScheduleEs:       "Todos los viernes, 7:00 p. m.",
		ImageFile:        "img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.webp",
		ImageAltEs:       "Grupo de alabanza",
		PublicationState: "",
	})
	if err != nil {
		t.Fatalf("ministryFromInput = %v", err)
	}

	if ministry.NameEs != "Alabanza" {
		t.Errorf("NameEs = %q, se esperaba normalizado", ministry.NameEs)
	}
	if ministry.DescriptionEs != "Grupo de alabanza" {
		t.Errorf("DescriptionEs = %q", ministry.DescriptionEs)
	}
	if ministry.DescriptionEn != nil {
		t.Errorf("DescriptionEn = %v, se esperaba nil (solo espacios)", ministry.DescriptionEn)
	}
	if ministry.LeaderPhone == nil || *ministry.LeaderPhone != "+58 412 000 0000" {
		t.Errorf("LeaderPhone = %v, se esperaba el valor recortado sin normalizar", ministry.LeaderPhone)
	}
	if ministry.LeaderVisible {
		t.Error("LeaderVisible = true, se esperaba false («Oculto por defecto», P6-3)")
	}
	if ministry.PublicationState != StateDraft {
		t.Errorf("PublicationState = %q, se esperaba draft (default del contrato)", ministry.PublicationState)
	}
	if ministry.SortOrder != 0 {
		t.Errorf("SortOrder = %d, se esperaba 0", ministry.SortOrder)
	}
}

// TestValidateMinistryAcceptsMinimalAndComplete fija los dos extremos válidos:
// solo nombre y descripción en español (Q1/Q2/Q4/Q6/Q8: el resto es opcional) y
// un ministerio con todos los datos.
func TestValidateMinistryAcceptsMinimalAndComplete(t *testing.T) {
	minimal := Ministry{NameEs: "Alabanza", DescriptionEs: "Grupo de alabanza"}
	if err := validateMinistry(minimal); err != nil {
		t.Errorf("validateMinistry(mínimo) = %v", err)
	}

	complete := Ministry{
		NameEs:        "Alabanza",
		NameEn:        textOf("Worship"),
		DescriptionEs: "Grupo de alabanza",
		DescriptionEn: textOf("Worship team"),
		LeaderNameEs:  textOf("Ana Pérez"),
		LeaderNameEn:  textOf("Ana Perez"),
		LeaderPhone:   textOf("+58 412 000 0000"),
		LeaderVisible: true,
		WhatsappURL:   textOf("https://ejemplo.com/grupo"),
		ScheduleEs:    textOf("Todos los viernes, 7:00 p. m."),
		ScheduleEn:    textOf("Every Friday, 7:00 pm"),
		PlaceEs:       textOf("Templo"),
		PlaceEn:       textOf("Temple"),
		ImageFile:     textOf("img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.png"),
		ImageAltEs:    textOf("Grupo de alabanza"),
		ImageAltEn:    textOf("Worship team"),
	}
	if err := validateMinistry(complete); err != nil {
		t.Errorf("validateMinistry(completo) = %v", err)
	}

	// El encargado con solo nombre es válido (Q8: "solo nombre" sí, "solo
	// contacto" no).
	onlyName := Ministry{NameEs: "Alabanza", DescriptionEs: "D", LeaderNameEs: textOf("Ana")}
	if err := validateMinistry(onlyName); err != nil {
		t.Errorf("validateMinistry(solo nombre del encargado) = %v", err)
	}
}

// TestValidateMinistryReportsEachField fija FR-014/FR-019/FR-020 con un motivo
// por campo en `details`, con las claves del contrato.
func TestValidateMinistryReportsEachField(t *testing.T) {
	longName := strings.Repeat("a", MaxNameLength+1)
	longDescription := strings.Repeat("b", MaxDescriptionLength+1)
	longSchedule := strings.Repeat("c", MaxScheduleLength+1)
	longPlace := strings.Repeat("d", MaxPlaceLength+1)
	longAlt := strings.Repeat("e", MaxImageAltLength+1)

	tests := []struct {
		name   string
		mutate func(*Ministry)
		field  string
	}{
		{
			name:   "nombre en español obligatorio",
			mutate: func(m *Ministry) { m.NameEs = "   " },
			field:  "nameEs",
		},
		{
			name:   "descripción en español obligatoria",
			mutate: func(m *Ministry) { m.DescriptionEs = "" },
			field:  "descriptionEs",
		},
		{
			name:   "nombre demasiado largo",
			mutate: func(m *Ministry) { m.NameEs = longName },
			field:  "nameEs",
		},
		{
			name:   "descripción demasiado larga",
			mutate: func(m *Ministry) { m.DescriptionEs = longDescription },
			field:  "descriptionEs",
		},
		{
			name:   "nombre en inglés demasiado largo",
			mutate: func(m *Ministry) { m.NameEn = textOf(longName) },
			field:  "nameEn",
		},
		{
			name:   "horario en inglés sin el español",
			mutate: func(m *Ministry) { m.ScheduleEn = textOf("Every Friday") },
			field:  "scheduleEs",
		},
		{
			name:   "horario demasiado largo",
			mutate: func(m *Ministry) { m.ScheduleEs = textOf(longSchedule) },
			field:  "scheduleEs",
		},
		{
			name:   "lugar en inglés sin el español",
			mutate: func(m *Ministry) { m.PlaceEn = textOf("Temple") },
			field:  "placeEs",
		},
		{
			name:   "lugar demasiado largo",
			mutate: func(m *Ministry) { m.PlaceEs = textOf(longPlace) },
			field:  "placeEs",
		},
		{
			name:   "nombre del encargado en inglés sin el español",
			mutate: func(m *Ministry) { m.LeaderNameEn = textOf("Ana") },
			field:  "leaderNameEs",
		},
		{
			name:   "nombre del encargado demasiado largo",
			mutate: func(m *Ministry) { m.LeaderNameEs = textOf(longName) },
			field:  "leaderNameEs",
		},
		{
			name: "celular y correo juntos",
			mutate: func(m *Ministry) {
				m.LeaderNameEs = textOf("Ana")
				m.LeaderPhone = textOf("+58 412 000 0000")
				m.LeaderEmail = textOf("ana@ejemplo.com")
			},
			field: "leaderEmail",
		},
		{
			name: "contacto sin nombre del encargado",
			mutate: func(m *Ministry) {
				m.LeaderEmail = textOf("ana@ejemplo.com")
			},
			field: "leaderNameEs",
		},
		{
			name: "celular demasiado corto",
			mutate: func(m *Ministry) {
				m.LeaderNameEs = textOf("Ana")
				m.LeaderPhone = textOf("12345")
			},
			field: "leaderPhone",
		},
		{
			name: "celular sin dígitos suficientes",
			mutate: func(m *Ministry) {
				m.LeaderNameEs = textOf("Ana")
				m.LeaderPhone = textOf("+58 (4) 12")
			},
			field: "leaderPhone",
		},
		{
			name: "celular demasiado largo",
			mutate: func(m *Ministry) {
				m.LeaderNameEs = textOf("Ana")
				m.LeaderPhone = textOf(strings.Repeat("1", MaxLeaderPhoneLength+1))
			},
			field: "leaderPhone",
		},
		{
			name:   "correo con formato inválido",
			mutate: func(m *Ministry) { m.LeaderNameEs = textOf("Ana"); m.LeaderEmail = textOf("no-es-correo") },
			field:  "leaderEmail",
		},
		{
			name:   "correo demasiado largo",
			mutate: func(m *Ministry) { m.LeaderEmail = textOf(strings.Repeat("a", MaxLeaderEmailLength) + "@x.co") },
			field:  "leaderEmail",
		},
		{
			name:   "imagen sin texto alternativo en español",
			mutate: func(m *Ministry) { m.ImageFile = textOf("img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.png") },
			field:  "imageAltEs",
		},
		{
			name:   "nombre de archivo que no generó el servidor",
			mutate: func(m *Ministry) { m.ImageFile = textOf("../../etc/passwd"); m.ImageAltEs = textOf("Alt") },
			field:  "imageFile",
		},
		{
			name:   "texto alternativo demasiado largo",
			mutate: func(m *Ministry) { m.ImageAltEs = textOf(longAlt) },
			field:  "imageAltEs",
		},
		{
			name:   "texto alternativo en inglés sin el español",
			mutate: func(m *Ministry) { m.ImageAltEn = textOf("Worship team") },
			field:  "imageAltEs",
		},
		{
			name:   "orden por encima del rango de la columna",
			mutate: func(m *Ministry) { m.SortOrder = MaxSortOrder + 1 },
			field:  "sortOrder",
		},
		{
			name:   "orden por debajo del rango de la columna",
			mutate: func(m *Ministry) { m.SortOrder = MinSortOrder - 1 },
			field:  "sortOrder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ministry := Ministry{NameEs: "Alabanza", DescriptionEs: "Grupo de alabanza"}
			tt.mutate(&ministry)

			details := detailsOf(t, validateMinistry(ministry))
			if _, ok := details[tt.field]; !ok {
				t.Fatalf("details = %v, se esperaba la clave %q", details, tt.field)
			}
		})
	}
}

// TestValidateMinistryLinkIsHTTPSAndSafe fija la tabla de casos de SC-009/
// FR-015/Q3 (R6-5): solo `https` válido y seguro, **sin restricción de
// dominios**, y el valor se sirve tal cual (no se transforma).
func TestValidateMinistryLinkIsHTTPSAndSafe(t *testing.T) {
	tests := []struct {
		name    string
		link    string
		wantErr bool
	}{
		{name: "grupo de WhatsApp", link: "https://chat.whatsapp.com/AbCdEf123"},
		{name: "dominio que no es de WhatsApp (Q3)", link: "https://ejemplo.com/grupo"},
		{name: "https con puerto y ruta", link: "https://ejemplo.com:8443/g?x=1&y=2#z"},
		{name: "límite exacto de 500", link: "https://ejemplo.com/" + strings.Repeat("a", 480)},
		{name: "javascript", link: "javascript:alert(1)", wantErr: true},
		{name: "JAVASCRIPT en mayúsculas", link: "JAVASCRIPT:alert(1)", wantErr: true},
		{name: "data", link: "data:text/html,<script>alert(1)</script>", wantErr: true},
		{name: "vbscript", link: "vbscript:msgbox(1)", wantErr: true},
		{name: "http", link: "http://ejemplo.com/grupo", wantErr: true},
		{name: "ruta relativa", link: "//evil.example/grupo", wantErr: true},
		{name: "sin host", link: "https://", wantErr: true},
		{name: "con espacio", link: "https://a b", wantErr: true},
		{name: "con salto de línea", link: "https://ejemplo.com/\ngrupo", wantErr: true},
		{name: "501 caracteres", link: "https://ejemplo.com/" + strings.Repeat("a", 481), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ministry := Ministry{
				NameEs:        "Alabanza",
				DescriptionEs: "Grupo de alabanza",
				WhatsappURL:   textOf(tt.link),
			}
			err := validateMinistry(ministry)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("validateMinistry(%q) = %v", tt.link, err)
				}
				return
			}
			details := detailsOf(t, err)
			if _, ok := details["whatsappUrl"]; !ok {
				t.Fatalf("details = %v, se esperaba la clave whatsappUrl", details)
			}
		})
	}

	// El enlace guardado es el enlace servido: la validación no transforma el
	// valor (FR-015).
	link := "https://ejemplo.com/grupo?x=1"
	ministry, err := ministryFromInput(MinisterioInput{
		NameEs:        "Alabanza",
		DescriptionEs: "D",
		WhatsappURL:   "  " + link + "  ",
	})
	if err != nil {
		t.Fatalf("ministryFromInput = %v", err)
	}
	if ministry.WhatsappURL == nil || *ministry.WhatsappURL != link {
		t.Errorf("WhatsappURL = %v, se esperaba %q sin transformar", ministry.WhatsappURL, link)
	}
}

// TestContentLabels fija las etiquetas de auditoría del módulo (R6-12).
func TestContentLabels(t *testing.T) {
	if got := contentLabel(""); got != "Ministerios" {
		t.Errorf("contentLabel(\"\") = %q", got)
	}
	if got := contentLabel("  Alabanza  "); got != "Ministerios · Alabanza" {
		t.Errorf("contentLabel = %q", got)
	}
	if got := imageLabel("img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.webp"); got !=
		"Ministerios · Imagen · img_aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa.webp" {
		t.Errorf("imageLabel = %q", got)
	}

	// El recorte defensivo mantiene la etiqueta dentro de lo que admite la tabla
	// (nunca se alcanza con los límites del módulo, pero no rompe el CHECK).
	huge := contentLabel(strings.Repeat("ñ", maxTargetLabel*2))
	if got := len([]rune(huge)); got != maxTargetLabel {
		t.Errorf("etiqueta recortada = %d caracteres, se esperaban %d", got, maxTargetLabel)
	}
}

// TestAuditCodesForTheModule fija los códigos del registro cerrado que usa el
// módulo (R6-12/A6-10).
func TestAuditCodesForTheModule(t *testing.T) {
	if got := stateCode(StatePublished); got != audit.ActionMinistryPublish {
		t.Errorf("stateCode(published) = %q", got)
	}
	if got := stateCode(StateDraft); got != audit.ActionMinistryUnpublish {
		t.Errorf("stateCode(draft) = %q", got)
	}
	if got := leaderCode(true); got != audit.ActionMinistryLeaderShow {
		t.Errorf("leaderCode(true) = %q", got)
	}
	if got := leaderCode(false); got != audit.ActionMinistryLeaderHide {
		t.Errorf("leaderCode(false) = %q", got)
	}
	if got := contentAction(uuid.New(), audit.ActionMinistryCreate, contentLabel("Alabanza")); got.TargetKind != audit.TargetContent {
		t.Errorf("TargetKind = %q, se esperaba content", got.TargetKind)
	}
}

// TestLogErrorUsesTheRequestLogger fija que un fallo best-effort se registra con
// el logger de la petición (que lleva el request_id) cuando está en el contexto,
// y con el logger base del service cuando no lo está.
func TestLogErrorUsesTheRequestLogger(t *testing.T) {
	baseLogger, baseCapture := testutil.NewLogger()
	requestLogger, requestCapture := testutil.NewLogger()
	svc := NewService(ServiceDeps{Repository: newFakeRepository(), Logger: baseLogger})

	svc.logError(context.Background(), "sin request logger", errors.New("boom"))
	if _, ok := baseCapture.Find("sin request logger"); !ok {
		t.Error("el logger base no registró el fallo")
	}

	ctx := httpserver.ContextWithRequestLogger(context.Background(), requestLogger)
	svc.logError(ctx, "con request logger", errors.New("boom"))
	if _, ok := requestCapture.Find("con request logger"); !ok {
		t.Error("el logger de la petición no registró el fallo")
	}
	if _, ok := baseCapture.Find("con request logger"); ok {
		t.Error("el fallo se registró en el logger base en vez del de la petición")
	}

	// Un service sin logger cae al logger por defecto sin reventar.
	NewService(ServiceDeps{Repository: newFakeRepository()}).logError(context.Background(), "sin logger", errors.New("boom"))
}

// TestMergeInvalidIgnoresNonDomainErrors fija que el traspaso de detalles de
// platform/validate es defensivo: un error que no es de dominio (o nil) no
// inventa campos.
func TestMergeInvalidIgnoresNonDomainErrors(t *testing.T) {
	details := map[string]any{}
	mergeInvalid(details, "whatsappUrl", nil)
	mergeInvalid(details, "whatsappUrl", errors.New("boom"))
	if len(details) != 0 {
		t.Errorf("details = %v, se esperaba vacío", details)
	}

	mergeInvalid(details, "whatsappUrl", apperr.Invalid("x", apperr.WithDetails(map[string]any{"URL": "motivo"})))
	if got, ok := details["whatsappUrl"]; !ok || got != "motivo" {
		t.Errorf("details = %v, se esperaba el motivo en whatsappUrl", details)
	}
}
