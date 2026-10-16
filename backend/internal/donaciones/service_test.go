package donaciones

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
)

// fakeRepository es el repositorio falso de los tests del service: un almacén en
// memoria que imita el comportamiento observable del repository real (solo
// publicado en la vía pública, orden `CASE kind → iban primero, sort_order, id`,
// tope, duplicado exacto por tipo+número con exclusión de la propia fila,
// borradores incluidos en el panel) y recuerda lo que se le pidió, para poder
// comprobar los topes y las filas de auditoría.
type fakeRepository struct {
	methods []MetodoDonacion
	texts   []TextoDonacion

	// Auditoría: `recorded` son las acciones best-effort (denegaciones y
	// rechazos) y `committed` las que acompañan a una mutación en su transacción.
	recorded  []audit.Action
	committed []audit.Action

	// Guionización de errores y de la detección de duplicados.
	readErr           error
	mutationErr       error
	getErr            error
	duplicateCheckErr error
	recordErr         error
	duplicate         bool

	// Observaciones de las consultas y de las mutaciones: `publicLimits` recuerda
	// el tope pedido a la vía pública y `textUpserts` las claves de los textos que
	// llegaron a guardarse (para fijar el no-op de `PutTexto` sin cambios).
	publicLimits []int
	textUpserts  []TextKey
}

var _ Repository = (*fakeRepository)(nil)

func newFakeRepository() *fakeRepository { return &fakeRepository{} }

// seed agrega métodos y textos al almacén sin pasar por el service.
func (f *fakeRepository) seedMethods(methods ...MetodoDonacion) {
	f.methods = append(f.methods, methods...)
}

func (f *fakeRepository) seedTexts(texts ...TextoDonacion) {
	f.texts = append(f.texts, texts...)
}

func (f *fakeRepository) publishedMethods(limit int) []MetodoDonacion {
	items := make([]MetodoDonacion, 0, len(f.methods))
	for _, method := range f.methods {
		if method.PublicationState == StatePublished {
			items = append(items, method)
		}
	}
	sortMethods(items)
	if limit >= 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func sortMethods(items []MetodoDonacion) {
	sort.SliceStable(items, func(i, j int) bool {
		if kindRank(items[i].Kind) != kindRank(items[j].Kind) {
			return kindRank(items[i].Kind) < kindRank(items[j].Kind)
		}
		if items[i].SortOrder != items[j].SortOrder {
			return items[i].SortOrder < items[j].SortOrder
		}
		return items[i].ID.String() < items[j].ID.String()
	})
}

func kindRank(k MetodoKind) int {
	if k == KindIban {
		return 0
	}
	return 1
}

func (f *fakeRepository) findMethodIndex(id uuid.UUID) int {
	for i := range f.methods {
		if f.methods[i].ID == id {
			return i
		}
	}
	return -1
}

func (f *fakeRepository) findTextIndex(key TextKey) int {
	for i := range f.texts {
		if f.texts[i].TextKey == key {
			return i
		}
	}
	return -1
}

// --- Auditoría ---

func (f *fakeRepository) RecordAction(_ context.Context, action audit.Action) error {
	if f.recordErr != nil {
		return f.recordErr
	}
	f.recorded = append(f.recorded, action)
	return nil
}

// --- Vía pública ---

func (f *fakeRepository) ListPublishedDonationMethods(_ context.Context, limit int) ([]MetodoDonacion, error) {
	f.publicLimits = append(f.publicLimits, limit)
	if f.readErr != nil {
		return nil, f.readErr
	}
	return f.publishedMethods(limit), nil
}

func (f *fakeRepository) ListPublishedDonationTexts(_ context.Context) ([]TextoDonacion, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	items := make([]TextoDonacion, 0, len(f.texts))
	for _, text := range f.texts {
		if text.PublicationState == StatePublished {
			items = append(items, text)
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].TextKey < items[j].TextKey })
	return items, nil
}

// --- Panel: métodos ---

func (f *fakeRepository) InsertDonationMethod(_ context.Context, method MetodoDonacion, actions ...audit.Action) (MetodoDonacion, error) {
	if f.mutationErr != nil {
		return MetodoDonacion{}, f.mutationErr
	}
	if f.duplicate {
		return MetodoDonacion{}, apperr.Conflict(messageDuplicate)
	}
	method.ID = uuid.New()
	method.CreatedAt, method.UpdatedAt = time.Now(), time.Now()
	f.methods = append(f.methods, method)
	f.committed = append(f.committed, actions...)
	return method, nil
}

func (f *fakeRepository) UpdateDonationMethod(_ context.Context, method MetodoDonacion, actions ...audit.Action) (MetodoDonacion, error) {
	if f.mutationErr != nil {
		return MetodoDonacion{}, f.mutationErr
	}
	if f.duplicate {
		return MetodoDonacion{}, apperr.Conflict(messageDuplicate)
	}
	index := f.findMethodIndex(method.ID)
	if index < 0 {
		return MetodoDonacion{}, apperr.NotFound(messageNotFound)
	}
	method.CreatedAt = f.methods[index].CreatedAt
	method.UpdatedAt = time.Now()
	f.methods[index] = method
	f.committed = append(f.committed, actions...)
	return method, nil
}

func (f *fakeRepository) DeleteDonationMethod(_ context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	if f.mutationErr != nil {
		return false, f.mutationErr
	}
	index := f.findMethodIndex(id)
	if index < 0 {
		return false, nil
	}
	f.methods = append(f.methods[:index], f.methods[index+1:]...)
	f.committed = append(f.committed, actions...)
	return true, nil
}

func (f *fakeRepository) GetDonationMethodByID(_ context.Context, id uuid.UUID) (MetodoDonacion, bool, error) {
	if f.getErr != nil {
		return MetodoDonacion{}, false, f.getErr
	}
	if f.readErr != nil {
		return MetodoDonacion{}, false, f.readErr
	}
	index := f.findMethodIndex(id)
	if index < 0 {
		return MetodoDonacion{}, false, nil
	}
	return f.methods[index], true, nil
}

func (f *fakeRepository) ListDonationMethods(_ context.Context, state *PublicationState, params paginate.Params) ([]MetodoDonacion, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	items := make([]MetodoDonacion, 0, len(f.methods))
	for _, method := range f.methods {
		if state != nil && method.PublicationState != *state {
			continue
		}
		items = append(items, method)
	}
	sortMethods(items)
	if params.Offset < len(items) {
		items = items[params.Offset:]
	} else {
		items = nil
	}
	if len(items) > params.Limit {
		items = items[:params.Limit]
	}
	return items, nil
}

func (f *fakeRepository) CountDonationMethods(_ context.Context, state *PublicationState) (int64, error) {
	if f.readErr != nil {
		return 0, f.readErr
	}
	var total int64
	for _, method := range f.methods {
		if state != nil && method.PublicationState != *state {
			continue
		}
		total++
	}
	return total, nil
}

func (f *fakeRepository) ExistsDonationMethodDuplicate(_ context.Context, kind MetodoKind, number string, excludeID *uuid.UUID) (bool, error) {
	if f.duplicateCheckErr != nil {
		return false, f.duplicateCheckErr
	}
	if f.readErr != nil {
		return false, f.readErr
	}
	if f.duplicate {
		return true, nil
	}
	for _, method := range f.methods {
		if excludeID != nil && method.ID == *excludeID {
			continue
		}
		if method.Kind == kind && method.Number == number {
			return true, nil
		}
	}
	return false, nil
}

// --- Panel: textos ---

func (f *fakeRepository) ListDonationTexts(_ context.Context) ([]TextoDonacion, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	items := append([]TextoDonacion(nil), f.texts...)
	sort.SliceStable(items, func(i, j int) bool { return items[i].TextKey < items[j].TextKey })
	return items, nil
}

func (f *fakeRepository) GetDonationTextByKey(_ context.Context, key TextKey) (TextoDonacion, bool, error) {
	if f.readErr != nil {
		return TextoDonacion{}, false, f.readErr
	}
	index := f.findTextIndex(key)
	if index < 0 {
		return TextoDonacion{}, false, nil
	}
	return f.texts[index], true, nil
}

func (f *fakeRepository) UpsertDonationText(_ context.Context, text TextoDonacion, actions ...audit.Action) (TextoDonacion, error) {
	if f.mutationErr != nil {
		return TextoDonacion{}, f.mutationErr
	}
	f.textUpserts = append(f.textUpserts, text.TextKey)
	index := f.findTextIndex(text.TextKey)
	if index < 0 {
		text.ID = uuid.New()
		text.CreatedAt, text.UpdatedAt = time.Now(), time.Now()
		f.texts = append(f.texts, text)
	} else {
		text.ID = f.texts[index].ID
		text.CreatedAt = f.texts[index].CreatedAt
		text.UpdatedAt = time.Now()
		f.texts[index] = text
	}
	f.committed = append(f.committed, actions...)
	return text, nil
}

// --- Helpers de prueba compartidos ---

// detailsOf devuelve los detalles del error de dominio de validación (falla si
// no es un error de dominio invalid).
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

// newTestService construye el service con el repositorio falso.
func newTestService(repo Repository) *service {
	return NewService(ServiceDeps{Repository: repo})
}

// publishedMethod construye un método publicado mínimo de un tipo dado.
func publishedMethod(kind MetodoKind, number string) MetodoDonacion {
	method := MetodoDonacion{
		ID:               uuid.New(),
		Kind:             kind,
		Number:           number,
		Institution:      "Banco Nacional",
		PublicationState: StatePublished,
	}
	if kind == KindIban {
		currency := CurrencyCRC
		method.Currency = &currency
	}
	return method
}

// --- T712 · límites, normalización y validación de dominio ---

// TestLimitsAreTheSingleSourceOfTruth fija los límites de negocio revisables
// (T712): los usan T713–T716 y los cita el contrato.
func TestLimitsAreTheSingleSourceOfTruth(t *testing.T) {
	limits := map[string]int{
		"MaxInstitutionLength": MaxInstitutionLength,
		"MaxHolderLength":      MaxHolderLength,
		"MaxTextLength":        MaxTextLength,
		"MaxNumberInputLength": MaxNumberInputLength,
		"MaxNumberLength":      MaxNumberLength,
		"MinSortOrder":         MinSortOrder,
		"MaxSortOrder":         MaxSortOrder,
		"maxPublicItems":       maxPublicItems,
		"maxPortadaItems":      maxPortadaItems,
		"maxTargetLabel":       maxTargetLabel,
	}
	want := map[string]int{
		"MaxInstitutionLength": 160,
		"MaxHolderLength":      160,
		"MaxTextLength":        2000,
		"MaxNumberInputLength": 60,
		"MaxNumberLength":      34,
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

// TestNormalizeTextKeepsContentAndCollapsesSpaces fija la normalización (I6):
// trim, espacios/tabuladores internos colapsados, saltos de línea conservados y
// todo el contenido intacto (tildes, ñ, ¿¡ y emojis).
func TestNormalizeTextKeepsContentAndCollapsesSpaces(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "trim", input: "  Banco Nacional  ", want: "Banco Nacional"},
		{name: "espacios internos", input: "Banco   Nacional\tCR", want: "Banco Nacional CR"},
		{name: "saltos de línea intactos", input: "Primera línea\nSegunda línea", want: "Primera línea\nSegunda línea"},
		{name: "tildes, ñ y signos", input: "  ¿Cómo donar? ¡Dona!  ", want: "¿Cómo donar? ¡Dona!"},
		{name: "emojis", input: "Ofrenda 🙌🏽", want: "Ofrenda 🙌🏽"},
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
			t.Errorf("normalizeOptional(%q) = %v, se esperaba nil", input, got)
		}
	}
	if got := normalizeOptional("  Ana Pérez  "); got == nil || *got != "Ana Pérez" {
		t.Errorf("normalizeOptional(válido) = %v", got)
	}
}

// TestNormalizeLang fija el idioma por defecto y el 400 invalid con details.lang
// (FR-008): es la única implementación (la usan T713 y T714).
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
// el estado vacío es `draft`.
func TestPublicationStateOrDefaultsToDraft(t *testing.T) {
	if got := publicationStateOr(""); got != StateDraft {
		t.Errorf("publicationStateOr(\"\") = %q, se esperaba draft", got)
	}
	if got := publicationStateOr("published"); got != StatePublished {
		t.Errorf("publicationStateOr(published) = %q", got)
	}
}

// TestMethodFromInputNormalizesAndAppliesDefaults fija la traducción del DTO del
// contrato a la entidad: número normalizado, institution recortada, opcionales
// vacíos a nil y defaults del contrato (draft, sortOrder 0).
func TestMethodFromInputNormalizesAndAppliesDefaults(t *testing.T) {
	method, err := methodFromInput(DonacionInput{
		Kind:             "iban",
		Number:           "  cr05 0152-0200-1234-5678-90  ",
		Institution:      "  Banco Nacional  ",
		HolderEs:         "  Ana Pérez  ",
		HolderEn:         "   ",
		Currency:         "CRC",
		PublicationState: "",
	})
	if err != nil {
		t.Fatalf("methodFromInput = %v", err)
	}
	if method.Number != "CR05015202001234567890" {
		t.Errorf("number = %q, se esperaba normalizado", method.Number)
	}
	if method.Institution != "Banco Nacional" {
		t.Errorf("institution = %q", method.Institution)
	}
	if method.HolderEs == nil || *method.HolderEs != "Ana Pérez" {
		t.Errorf("holderEs = %v", method.HolderEs)
	}
	if method.HolderEn != nil {
		t.Errorf("holderEn = %v, se esperaba nil (solo espacios)", method.HolderEn)
	}
	if method.PublicationState != StateDraft {
		t.Errorf("publicationState = %q, se esperaba draft", method.PublicationState)
	}
	if method.SortOrder != 0 {
		t.Errorf("sortOrder = %d, se esperaba 0", method.SortOrder)
	}

	// SINPE sin prefijo: se normaliza a `+506XXXXXXXX`.
	sinpe, err := methodFromInput(DonacionInput{Kind: "sinpe", Number: "8888-8888", Institution: "Banco"})
	if err != nil {
		t.Fatalf("methodFromInput(sinpe) = %v", err)
	}
	if sinpe.Number != "+50688888888" {
		t.Errorf("number = %q, se esperaba +50688888888", sinpe.Number)
	}
	if sinpe.Currency != nil {
		t.Errorf("currency = %v, el SINPE no lleva moneda", sinpe.Currency)
	}
}

// TestValidateMethodReportsEachField fija FR-014 con un motivo por campo en
// `details`, con las claves del contrato.
func TestValidateMethodReportsEachField(t *testing.T) {
	longInstitution := strings.Repeat("a", MaxInstitutionLength+1)
	longHolder := strings.Repeat("b", MaxHolderLength+1)

	base := func() MetodoDonacion {
		currency := CurrencyCRC
		return MetodoDonacion{
			Kind:             KindIban,
			Number:           "CR05015202001234567890",
			Institution:      "Banco Nacional",
			Currency:         &currency,
			PublicationState: StateDraft,
		}
	}

	tests := []struct {
		name   string
		mutate func(*MetodoDonacion)
		field  string
	}{
		{
			name:   "tipo obligatorio",
			mutate: func(m *MetodoDonacion) { m.Kind = "" },
			field:  "kind",
		},
		{
			name:   "tipo fuera del registro",
			mutate: func(m *MetodoDonacion) { m.Kind = "otro" },
			field:  "kind",
		},
		{
			name:   "número obligatorio",
			mutate: func(m *MetodoDonacion) { m.Number = "   " },
			field:  "number",
		},
		{
			name:   "IBAN mal formado",
			mutate: func(m *MetodoDonacion) { m.Number = "CR05 0152 0200 1234 5678 9" },
			field:  "number",
		},
		{
			name:   "IBAN con emoji",
			mutate: func(m *MetodoDonacion) { m.Number = "CR050152-0200-1234-5678-9😀" },
			field:  "number",
		},
		{
			name:   "SINPE no celular",
			mutate: func(m *MetodoDonacion) { m.Kind = KindSinpe; m.Number = "22222222"; m.Currency = nil },
			field:  "number",
		},
		{
			name:   "institución obligatoria",
			mutate: func(m *MetodoDonacion) { m.Institution = "  " },
			field:  "institution",
		},
		{
			name:   "institución demasiado larga",
			mutate: func(m *MetodoDonacion) { m.Institution = longInstitution },
			field:  "institution",
		},
		{
			name:   "moneda obligatoria en IBAN",
			mutate: func(m *MetodoDonacion) { m.Currency = nil },
			field:  "currency",
		},
		{
			name:   "moneda fuera del registro",
			mutate: func(m *MetodoDonacion) { c := Currency("EUR"); m.Currency = &c },
			field:  "currency",
		},
		{
			name: "moneda presente en SINPE",
			mutate: func(m *MetodoDonacion) {
				m.Kind = KindSinpe
				m.Number = "+50688888888"
			},
			field: "currency",
		},
		{
			name:   "titular en inglés sin el español",
			mutate: func(m *MetodoDonacion) { m.HolderEn = textOf("Ana Perez") },
			field:  "holderEs",
		},
		{
			name:   "titular demasiado largo",
			mutate: func(m *MetodoDonacion) { m.HolderEs = textOf(longHolder) },
			field:  "holderEs",
		},
		{
			name:   "estado fuera del registro",
			mutate: func(m *MetodoDonacion) { m.PublicationState = "archived" },
			field:  "publicationState",
		},
		{
			name:   "orden por encima del rango de la columna",
			mutate: func(m *MetodoDonacion) { m.SortOrder = MaxSortOrder + 1 },
			field:  "sortOrder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := base()
			tt.mutate(&method)
			details := detailsOf(t, validateMethod(&method))
			if _, ok := details[tt.field]; !ok {
				t.Fatalf("details = %v, se esperaba la clave %q", details, tt.field)
			}
		})
	}
}

// TestValidateMethodAcceptsMinimalAndComplete fija los extremos válidos: IBAN
// mínimo (kind, number, institution, currency) y un método completo.
func TestValidateMethodAcceptsMinimalAndComplete(t *testing.T) {
	currency := CurrencyCRC
	minimal := MetodoDonacion{
		Kind:             KindIban,
		Number:           "CR05015202001234567890",
		Institution:      "Banco Nacional",
		Currency:         &currency,
		PublicationState: StateDraft,
	}
	if err := validateMethod(&minimal); err != nil {
		t.Errorf("validateMethod(mínimo) = %v", err)
	}

	usd := CurrencyUSD
	complete := MetodoDonacion{
		Kind:             KindIban,
		Number:           "ES9121000418450200051332",
		Institution:      "Banco Santander",
		HolderEs:         textOf("Ana Pérez"),
		HolderEn:         textOf("Ana Perez"),
		Currency:         &usd,
		PublicationState: StatePublished,
		SortOrder:        3,
	}
	if err := validateMethod(&complete); err != nil {
		t.Errorf("validateMethod(completo) = %v", err)
	}
}

// TestValidateMethodKeepsContentOfTexts fija que las tildes, la ñ, los signos
// ¿¡ y los emojis quedan intactos en el titular (contenido, no código).
func TestValidateMethodKeepsContentOfTexts(t *testing.T) {
	method, err := methodFromInput(DonacionInput{
		Kind:        "iban",
		Number:      "CR05015202001234567890",
		Institution: "Banco",
		HolderEs:    "José Núñez ¿quién? 🙌🏽",
		Currency:    "CRC",
	})
	if err != nil {
		t.Fatalf("methodFromInput = %v", err)
	}
	if method.HolderEs == nil || *method.HolderEs != "José Núñez ¿quién? 🙌🏽" {
		t.Errorf("holderEs = %v, se esperaba intacto", method.HolderEs)
	}
}

// TestTextoFromInputNormalizesAndValidates fija la validación de los textos:
// `textEs` obligatorio, 2.000 caracteres por idioma, `textEn` opcional y
// contenido intacto.
func TestTextoFromInputNormalizesAndValidates(t *testing.T) {
	texto, err := textoFromInput(TextKeyIntro, DonacionTextoInput{
		TextEs:           "  Cómo donar 🙌🏽  ",
		TextEn:           "   ",
		PublicationState: "",
	})
	if err != nil {
		t.Fatalf("textoFromInput = %v", err)
	}
	if texto.BodyEs != "Cómo donar 🙌🏽" {
		t.Errorf("bodyEs = %q", texto.BodyEs)
	}
	if texto.BodyEn != nil {
		t.Errorf("bodyEn = %v, se esperaba nil", texto.BodyEn)
	}
	if texto.PublicationState != StateDraft {
		t.Errorf("publicationState = %q", texto.PublicationState)
	}

	// textEs obligatorio.
	_, err = textoFromInput(TextKeyIntro, DonacionTextoInput{TextEs: "   "})
	if _, ok := detailsOf(t, err)["textEs"]; !ok {
		t.Fatalf("se esperaba details.textEs: %v", err)
	}

	// 2.001 caracteres → details.textEs.
	long := strings.Repeat("a", MaxTextLength+1)
	_, err = textoFromInput(TextKeyIntro, DonacionTextoInput{TextEs: long})
	if _, ok := detailsOf(t, err)["textEs"]; !ok {
		t.Fatalf("se esperaba details.textEs (límite): %v", err)
	}

	// textEn demasiado largo → details.textEn.
	_, err = textoFromInput(TextKeyIntro, DonacionTextoInput{TextEs: "Ok", TextEn: long})
	if _, ok := detailsOf(t, err)["textEn"]; !ok {
		t.Fatalf("se esperaba details.textEn: %v", err)
	}

	// Clave fuera del catálogo → details.textKey.
	_, err = textoFromInput(TextKey("otro"), DonacionTextoInput{TextEs: "Ok"})
	if _, ok := detailsOf(t, err)["textKey"]; !ok {
		t.Fatalf("se esperaba details.textKey: %v", err)
	}
}

// TestContentLabels fija las etiquetas de auditoría del módulo (R7-10).
func TestContentLabels(t *testing.T) {
	if got := contentLabel(""); got != "Donaciones" {
		t.Errorf("contentLabel(\"\") = %q", got)
	}
	if got := contentLabel("  Banco Nacional · CR05015202001234567890  "); got !=
		"Donaciones · Banco Nacional · CR05015202001234567890" {
		t.Errorf("contentLabel = %q", got)
	}
	if got := textLabel(TextKeyIntro); got != "Donaciones · Texto introductorio" {
		t.Errorf("textLabel(intro) = %q", got)
	}
	if got := textLabel(TextKeyUsage); got != "Donaciones · Texto de uso" {
		t.Errorf("textLabel(usage) = %q", got)
	}
	if got := textLabel(TextKey("otro")); got != "Donaciones" {
		t.Errorf("textLabel(desconocida) = %q", got)
	}

	// El recorte defensivo mantiene la etiqueta dentro del tope de la tabla.
	huge := contentLabel(strings.Repeat("ñ", maxTargetLabel*2))
	if got := len([]rune(huge)); got != maxTargetLabel {
		t.Errorf("etiqueta recortada = %d caracteres, se esperaban %d", got, maxTargetLabel)
	}
}

// TestAuditCodesForTheModule fija los códigos del registro cerrado que usa el
// módulo (R7-10).
func TestAuditCodesForTheModule(t *testing.T) {
	if got := stateCode(StatePublished); got != audit.ActionDonationPublish {
		t.Errorf("stateCode(published) = %q", got)
	}
	if got := stateCode(StateDraft); got != audit.ActionDonationUnpublish {
		t.Errorf("stateCode(draft) = %q", got)
	}
	if got := textStateCode(StatePublished); got != audit.ActionDonationTextPublish {
		t.Errorf("textStateCode(published) = %q", got)
	}
	if got := textStateCode(StateDraft); got != audit.ActionDonationTextUnpublish {
		t.Errorf("textStateCode(draft) = %q", got)
	}
	if got := contentAction(uuid.New(), audit.ActionDonationCreate, contentLabel("Banco")); got.TargetKind != audit.TargetContent {
		t.Errorf("TargetKind = %q, se esperaba content", got.TargetKind)
	}
}
