package agenda

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/testutil"
)

// fakeRepository es el repositorio falso del service: devuelve datos guionizados
// y recuerda las acciones de auditoría registradas best-effort. Los métodos que
// no se ejercitan en cada prueba devuelven el valor cero.
//
// Para la cara pública aplica la MISMA función de dominio de la regla R4-2
// (isUpcoming) sobre los datos guionizados y recuerda el corte con el que se le
// consultó: las pruebas del service comprueban así el filtro "publicado y
// próximo" y los parámetros del corte, mientras que el SQL real (que es quien
// filtra en producción) se verifica en las pruebas de integración (T412/T413).
type fakeRepository struct {
	recorded  []audit.Action
	recordErr error

	events     []Event
	eventDates map[uuid.UUID][]EventDate
	activities []Activity
	categories []Category

	cutoffs     []cutoffCall
	blockLimits []int
}

// cutoffCall es una llamada al repositorio con el corte "hoy"/"ahora".
type cutoffCall struct {
	today   string
	nowTime string
}

// upcomingDates devuelve las fechas próximas guionizadas de un evento.
func (f *fakeRepository) upcomingDates(eventID uuid.UUID, today, nowTime string) []EventDate {
	out := make([]EventDate, 0, len(f.eventDates[eventID]))
	for _, date := range f.eventDates[eventID] {
		if isUpcoming(date.Date, today, nowTime, date.EndTime) {
			out = append(out, date)
		}
	}
	return out
}

// compile-time check: el falso satisface el puerto del service.
var _ Repository = (*fakeRepository)(nil)

func (f *fakeRepository) RecordAction(_ context.Context, action audit.Action) error {
	if f.recordErr != nil {
		return f.recordErr
	}
	f.recorded = append(f.recorded, action)
	return nil
}

func (f *fakeRepository) ListCategories(context.Context) ([]Category, error) {
	return f.categories, nil
}

func (f *fakeRepository) GetCategoryByID(context.Context, uuid.UUID) (Category, bool, error) {
	return Category{}, false, nil
}

func (f *fakeRepository) InsertCategory(context.Context, Category, ...audit.Action) (Category, error) {
	return Category{}, nil
}

func (f *fakeRepository) UpdateCategory(context.Context, Category, ...audit.Action) (Category, error) {
	return Category{}, nil
}

func (f *fakeRepository) DeleteCategory(context.Context, uuid.UUID, ...audit.Action) (bool, error) {
	return false, nil
}

func (f *fakeRepository) CountCategoryUsage(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}

func (f *fakeRepository) InsertEvent(context.Context, Event, []EventDate, ...audit.Action) (Event, error) {
	return Event{}, nil
}

func (f *fakeRepository) UpdateEvent(context.Context, Event, []EventDate, ...audit.Action) (Event, error) {
	return Event{}, nil
}

func (f *fakeRepository) DeleteEvent(context.Context, uuid.UUID, ...audit.Action) (bool, error) {
	return false, nil
}

func (f *fakeRepository) GetEventByID(context.Context, uuid.UUID) (Event, bool, error) {
	return Event{}, false, nil
}

func (f *fakeRepository) ListEventDates(context.Context, uuid.UUID) ([]EventDate, error) {
	return nil, nil
}

func (f *fakeRepository) ListUpcomingEventDates(_ context.Context, eventID uuid.UUID, today, nowTime string) ([]EventDate, error) {
	return f.upcomingDates(eventID, today, nowTime), nil
}

func (f *fakeRepository) ExistsEventDuplicate(context.Context, string, string, []string, *uuid.UUID) (bool, error) {
	return false, nil
}

func (f *fakeRepository) ListEvents(context.Context, *PublicationState, paginate.Params) ([]Event, error) {
	return nil, nil
}

func (f *fakeRepository) CountEvents(context.Context, *PublicationState) (int64, error) {
	return 0, nil
}

func (f *fakeRepository) ListUpcomingEvents(_ context.Context, today, nowTime string, limit int) ([]Event, error) {
	f.cutoffs = append(f.cutoffs, cutoffCall{today: today, nowTime: nowTime})
	f.blockLimits = append(f.blockLimits, limit)

	out := make([]Event, 0, len(f.events))
	for _, event := range f.events {
		if event.PublicationState != StatePublished {
			continue
		}
		if len(f.upcomingDates(event.ID, today, nowTime)) == 0 {
			continue
		}
		if len(out) >= limit {
			break
		}
		out = append(out, event)
	}
	return out, nil
}

func (f *fakeRepository) InsertActivity(context.Context, Activity, ...audit.Action) (Activity, error) {
	return Activity{}, nil
}

func (f *fakeRepository) UpdateActivity(context.Context, Activity, ...audit.Action) (Activity, error) {
	return Activity{}, nil
}

func (f *fakeRepository) DeleteActivity(context.Context, uuid.UUID, ...audit.Action) (bool, error) {
	return false, nil
}

func (f *fakeRepository) GetActivityByID(context.Context, uuid.UUID) (Activity, bool, error) {
	return Activity{}, false, nil
}

func (f *fakeRepository) ListActivities(context.Context, *PublicationState, paginate.Params) ([]Activity, error) {
	return nil, nil
}

func (f *fakeRepository) CountActivities(context.Context, *PublicationState) (int64, error) {
	return 0, nil
}

func (f *fakeRepository) ListUpcomingActivities(_ context.Context, today, nowTime string, limit int) ([]Activity, error) {
	f.cutoffs = append(f.cutoffs, cutoffCall{today: today, nowTime: nowTime})
	f.blockLimits = append(f.blockLimits, limit)

	out := make([]Activity, 0, len(f.activities))
	for _, activity := range f.activities {
		if activity.PublicationState != StatePublished {
			continue
		}
		if !isUpcoming(activity.Date, today, nowTime, activity.EndTime) {
			continue
		}
		if len(out) >= limit {
			break
		}
		out = append(out, activity)
	}
	return out, nil
}

// detailsOf devuelve los detalles del error de dominio (nil si no es apperr).
func detailsOf(t *testing.T, err error) map[string]any {
	t.Helper()
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("se esperaba un error de dominio, llegó %v", err)
	}
	return domainErr.Details
}

// newTestService construye el service con el reloj y la zona fijos de las
// pruebas (nunca el reloj real).
func newTestService(t *testing.T, repo Repository) *service {
	t.Helper()
	location, err := time.LoadLocation("America/Costa_Rica")
	if err != nil {
		t.Fatalf("cargar America/Costa_Rica: %v", err)
	}
	logger, _ := testutil.NewLogger()
	return NewService(ServiceDeps{
		Repository: repo,
		Location:   location,
		Now:        func() time.Time { return time.Date(2026, time.October, 11, 18, 0, 0, 0, time.UTC) },
		Logger:     logger,
	})
}

// TestLimitsAreTheSingleSourceOfTruth fija los límites de negocio revisables
// (T417): los usan T418–T421 y los cita el contrato (RG4-8).
func TestLimitsAreTheSingleSourceOfTruth(t *testing.T) {
	if MaxBlockItems != 200 {
		t.Errorf("MaxBlockItems = %d, se esperaban 200 (RG4-8)", MaxBlockItems)
	}
	if MaxDetailLength != 2000 {
		t.Errorf("MaxDetailLength = %d, se esperaban 2000 (FR-018/FR-019)", MaxDetailLength)
	}
	limits := map[string]int{
		"MaxTitleLength":        MaxTitleLength,
		"MaxPlaceLength":        MaxPlaceLength,
		"MaxURLLength":          MaxURLLength,
		"MaxOrganizerLength":    MaxOrganizerLength,
		"MaxFrequencyLength":    MaxFrequencyLength,
		"MaxImageAltLength":     MaxImageAltLength,
		"MaxCategoryNameLength": MaxCategoryNameLength,
		"MaxEventDates":         MaxEventDates,
		"MaxCategories":         MaxCategories,
	}
	want := map[string]int{
		"MaxTitleLength":        160,
		"MaxPlaceLength":        200,
		"MaxURLLength":          500,
		"MaxOrganizerLength":    160,
		"MaxFrequencyLength":    160,
		"MaxImageAltLength":     300,
		"MaxCategoryNameLength": 120,
		"MaxEventDates":         50,
		"MaxCategories":         100,
	}
	for name, got := range limits {
		if got != want[name] {
			t.Errorf("%s = %d, se esperaba %d", name, got, want[name])
		}
	}
}

// TestIsUpcomingFollowsTheRule fija los tres casos de la regla R4-2 con fechas
// fijas (sin depender de la hora real).
func TestIsUpcomingFollowsTheRule(t *testing.T) {
	const today = "2026-10-11"
	const now = "12:00"

	endEarly := "10:00"
	endLate := "22:00"

	tests := []struct {
		name    string
		date    string
		endTime *string
		want    bool
	}{
		{name: "fecha futura, sin hora", date: "2026-10-12", want: true},
		{name: "fecha futura con hora", date: "2026-10-12", endTime: &endEarly, want: true},
		{name: "hoy sin hora de fin", date: today, want: true},
		{name: "hoy con fin por delante", date: today, endTime: &endLate, want: true},
		{name: "hoy con fin ya terminado", date: today, endTime: &endEarly, want: false},
		{name: "fecha pasada", date: "2026-10-10", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUpcoming(tt.date, today, now, tt.endTime); got != tt.want {
				t.Errorf("isUpcoming(%s, end=%v) = %v, se esperaba %v", tt.date, tt.endTime, got, tt.want)
			}
		})
	}
}

// TestCutoffUsesAppTimeZone fija el corte de "hoy"/"ahora" en la zona del lugar
// (APP_TZ, T409): a medianoche UTC ya es "ayer" en Costa Rica (UTC−06:00).
func TestCutoffUsesAppTimeZone(t *testing.T) {
	costaRica, err := time.LoadLocation("America/Costa_Rica")
	if err != nil {
		t.Fatalf("cargar America/Costa_Rica: %v", err)
	}
	instant := time.Date(2026, time.October, 17, 0, 30, 0, 0, time.UTC)

	t.Run("zona del lugar", func(t *testing.T) {
		svc := NewService(ServiceDeps{Location: costaRica, Now: func() time.Time { return instant }})
		today, nowTime, err := svc.cutoff()
		if err != nil {
			t.Fatalf("cutoff: %v", err)
		}
		if today != "2026-10-16" || nowTime != "18:30" {
			t.Errorf("corte = %s %s, se esperaba 2026-10-16 18:30", today, nowTime)
		}
	})

	t.Run("sin zona configurada falla con error claro", func(t *testing.T) {
		svc := NewService(ServiceDeps{Now: func() time.Time { return instant }})
		_, _, err := svc.cutoff()
		if err == nil {
			t.Fatal("cutoff sin zona debería fallar (fail-closed)")
		}
		if errorKind(err) != apperr.KindInternal {
			t.Errorf("kind = %v, se esperaba internal", errorKind(err))
		}
		if !strings.Contains(err.Error(), "APP_TZ") {
			t.Errorf("el error no identifica la configuración: %v", err)
		}
	})
}

// TestNormalizeTextAndOptional fija la normalización (analyze I6): trim y
// espacios internos colapsados, `""`/espacios → NULL y tildes/ñ/¿¡/emojis
// intactos.
func TestNormalizeTextAndOptional(t *testing.T) {
	if got := normalizeText("  Noche   de   aposentos \t "); got != "Noche de aposentos" {
		t.Errorf("normalizeText = %q", got)
	}
	if got := normalizeText("¿¡Qué!? 🎉 ñandú"); got != "¿¡Qué!? 🎉 ñandú" {
		t.Errorf("normalizeText alteró el contenido: %q", got)
	}
	if got := normalizeOptional("   "); got != nil {
		t.Errorf("normalizeOptional(espacios) = %v, se esperaba nil", *got)
	}
	if got := normalizeOptional(" Prayer night "); got == nil || *got != "Prayer night" {
		t.Errorf("normalizeOptional = %v", got)
	}
	if got := publicationStateOr(""); got != StateDraft {
		t.Errorf("publicationStateOr(\"\") = %q, se esperaba draft", got)
	}
	if got := publicationStateOr("published"); got != StatePublished {
		t.Errorf("publicationStateOr(published) = %q", got)
	}
	for _, lang := range []string{LangES, LangEN} {
		if !validLang(lang) {
			t.Errorf("validLang(%q) = false", lang)
		}
	}
	for _, lang := range []string{"", "fr", "ES"} {
		if validLang(lang) {
			t.Errorf("validLang(%q) = true", lang)
		}
	}
}

// TestParseDateRejectsUnrealDates fija FR-015: solo fechas de calendario reales
// en "YYYY-MM-DD" (2026-13-32 → invalid con details.date).
func TestParseDateRejectsUnrealDates(t *testing.T) {
	if got, err := parseDate(" 2026-10-17 "); err != nil || got != "2026-10-17" {
		t.Fatalf("parseDate(válida) = %q (%v)", got, err)
	}
	for _, invalid := range []string{"", "2026-13-32", "2026-02-30", "17/10/2026", "2026-10-17T00:00:00Z", "hoy"} {
		_, err := parseDate(invalid)
		if errorKind(err) != apperr.KindInvalid {
			t.Errorf("parseDate(%q) = %v, se esperaba invalid", invalid, err)
		}
		if _, ok := detailsOf(t, err)["date"]; !ok {
			t.Errorf("parseDate(%q) sin details.date: %v", invalid, err)
		}
	}
}

// TestNormalizeRangeChecksTheSchedule fija Q2 y la decisión firme de la puerta 2:
// `endTime` exige `startTime` (400 con details.endTime) y es posterior a él.
func TestNormalizeRangeChecksTheSchedule(t *testing.T) {
	t.Run("sin horas", func(t *testing.T) {
		start, end, err := normalizeRange("", "")
		if err != nil || start != nil || end != nil {
			t.Fatalf("normalizeRange(vacío) = %v/%v (%v)", start, end, err)
		}
	})

	t.Run("solo inicio", func(t *testing.T) {
		start, end, err := normalizeRange("19:00", "")
		if err != nil || start == nil || *start != "19:00" || end != nil {
			t.Fatalf("normalizeRange(solo inicio) = %v/%v (%v)", start, end, err)
		}
	})

	t.Run("fin sin inicio", func(t *testing.T) {
		_, _, err := normalizeRange("", "21:00")
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("fin sin inicio = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["endTime"]; !ok {
			t.Errorf("sin details.endTime: %v", err)
		}
	})

	t.Run("fin anterior o igual", func(t *testing.T) {
		for _, end := range []string{"18:00", "19:00"} {
			_, _, err := normalizeRange("19:00", end)
			if errorKind(err) != apperr.KindInvalid {
				t.Errorf("fin %s = %v, se esperaba invalid", end, err)
			}
			if _, ok := detailsOf(t, err)["endTime"]; !ok {
				t.Errorf("fin %s sin details.endTime: %v", end, err)
			}
		}
	})

	t.Run("hora mal formada", func(t *testing.T) {
		for _, start := range []string{"25:00", "7:30", "19:60", "tarde"} {
			if _, _, err := normalizeRange(start, ""); errorKind(err) != apperr.KindInvalid {
				t.Errorf("inicio %q = %v, se esperaba invalid", start, err)
			}
		}
	})
}

// TestNormalizeEventDatesChecksTheList fija FR-015/R4-11: entre 1 y 50 fechas,
// sin repetir, cada una con su horario coherente; las fechas pasadas están
// permitidas (el elemento no será público sin ninguna próxima).
func TestNormalizeEventDatesChecksTheList(t *testing.T) {
	t.Run("sin fechas", func(t *testing.T) {
		_, err := normalizeEventDates(nil)
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("sin fechas = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["dates"]; !ok {
			t.Errorf("sin details.dates: %v", err)
		}
	})

	t.Run("más de 50", func(t *testing.T) {
		inputs := make([]FechaInput, 0, MaxEventDates+1)
		for i := 0; i <= MaxEventDates; i++ {
			inputs = append(inputs, FechaInput{Date: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, i).Format(DateFormat)})
		}
		if _, err := normalizeEventDates(inputs); errorKind(err) != apperr.KindInvalid {
			t.Fatalf("51 fechas = %v, se esperaba invalid", err)
		}
	})

	t.Run("fecha repetida", func(t *testing.T) {
		_, err := normalizeEventDates([]FechaInput{
			{Date: "2026-10-17"},
			{Date: "2026-10-17"},
		})
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("repetida = %v, se esperaba invalid", err)
		}
		if _, ok := detailsOf(t, err)["dates"]; !ok {
			t.Errorf("sin details.dates: %v", err)
		}
	})

	t.Run("fecha inexistente", func(t *testing.T) {
		_, err := normalizeEventDates([]FechaInput{{Date: "2026-13-32"}})
		if errorKind(err) != apperr.KindInvalid {
			t.Fatalf("fecha inexistente = %v, se esperaba invalid", err)
		}
	})

	t.Run("lista válida con horarios", func(t *testing.T) {
		dates, err := normalizeEventDates([]FechaInput{
			{Date: "2026-10-10"}, // pasada: permitida
			{Date: "2026-10-17", StartTime: "19:00", EndTime: "21:00"},
			{Date: "2026-10-24", StartTime: "19:00"},
		})
		if err != nil {
			t.Fatalf("normalizeEventDates: %v", err)
		}
		if len(dates) != 3 {
			t.Fatalf("fechas = %d, se esperaban 3", len(dates))
		}
		if dates[0].StartTime != nil || dates[0].EndTime != nil {
			t.Errorf("la fecha sin horario debe venir sin horas: %+v", dates[0])
		}
		if dates[1].StartTime == nil || *dates[1].StartTime != "19:00" || dates[1].EndTime == nil || *dates[1].EndTime != "21:00" {
			t.Errorf("horario de la segunda fecha = %+v", dates[1])
		}
		if dates[2].EndTime != nil {
			t.Errorf("la tercera fecha no debería tener fin: %+v", dates[2])
		}
	})
}

// TestValidateImageRequiresAlt fija FR-020: la imagen es opcional, pero con
// imagen el texto alternativo en español es obligatorio y el nombre debe ser el
// generado por el servidor.
func TestValidateImageRequiresAlt(t *testing.T) {
	if err := validateImage("", ""); err != nil {
		t.Errorf("sin imagen = %v, se esperaba nil", err)
	}
	if err := validateImage("img_11111111-1111-1111-1111-111111111111.jpg", "Noche de aposentos"); err != nil {
		t.Errorf("imagen con alt = %v, se esperaba nil", err)
	}

	err := validateImage("img_11111111-1111-1111-1111-111111111111.jpg", "   ")
	if errorKind(err) != apperr.KindInvalid {
		t.Fatalf("imagen sin alt = %v, se esperaba invalid", err)
	}
	if _, ok := detailsOf(t, err)["imageAltEs"]; !ok {
		t.Errorf("sin details.imageAltEs: %v", err)
	}

	err = validateImage("../../etc/passwd", "Algo")
	if errorKind(err) != apperr.KindInvalid {
		t.Fatalf("nombre inválido = %v, se esperaba invalid", err)
	}
	if _, ok := detailsOf(t, err)["imageFile"]; !ok {
		t.Errorf("sin details.imageFile: %v", err)
	}
}

// TestModuleCodesAndLabels fija la tabla de R4-10: cada módulo tiene su etiqueta
// y sus códigos del registro cerrado de platform/audit (incluidos los tres del
// catálogo de categorías).
func TestModuleCodesAndLabels(t *testing.T) {
	tests := []struct {
		module    Module
		label     string
		prefix    string
		hasImages bool
		hasState  bool
	}{
		{ModuleEvents, "Eventos", "event", true, true},
		{ModuleActivities, "Actividades", "activity", true, true},
		{ModuleCategories, "Categorías", "category", false, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.module), func(t *testing.T) {
			if got := tt.module.Label(); got != tt.label {
				t.Errorf("Label = %q, se esperaba %q", got, tt.label)
			}
			codes := []string{tt.module.CreateCode(), tt.module.UpdateCode(), tt.module.DeleteCode()}
			if got := tt.module.CreateCode(); got != tt.prefix+".create" {
				t.Errorf("CreateCode = %q, se esperaba %q", got, tt.prefix+".create")
			}
			if got := tt.module.UpdateCode(); got != tt.prefix+".update" {
				t.Errorf("UpdateCode = %q, se esperaba %q", got, tt.prefix+".update")
			}
			if got := tt.module.DeleteCode(); got != tt.prefix+".delete" {
				t.Errorf("DeleteCode = %q, se esperaba %q", got, tt.prefix+".delete")
			}
			if tt.hasImages {
				codes = append(codes, tt.module.ImageCode())
			}
			if tt.hasState {
				codes = append(codes, tt.module.PublishCode(), tt.module.UnpublishCode())
				if got := tt.module.StateCode(StatePublished); got != tt.prefix+".publish" {
					t.Errorf("StateCode(published) = %q", got)
				}
				if got := tt.module.StateCode(StateDraft); got != tt.prefix+".unpublish" {
					t.Errorf("StateCode(draft) = %q", got)
				}
			}
			for _, code := range codes {
				if !audit.ValidActionCode(code) {
					t.Errorf("el código %q no está en el registro cerrado de platform/audit", code)
				}
			}
		})
	}

	// El catálogo de categorías NO es contenido publicable (son etiquetas) ni
	// tiene imagen propia: sus códigos de estado y de imagen no existen.
	for _, code := range []string{"category.publish", "category.unpublish", "category.image.upload"} {
		if audit.ValidActionCode(code) {
			t.Errorf("el registro cerrado no debería admitir %q", code)
		}
	}
}

// TestContentLabels fija las etiquetas del objetivo de auditoría (R4-10).
func TestContentLabels(t *testing.T) {
	if got := contentLabel(ModuleEvents, "Noche de aposentos"); got != "Eventos · Noche de aposentos" {
		t.Errorf("contentLabel = %q", got)
	}
	if got := contentLabel(ModuleActivities, "  Repartir   comida "); got != "Actividades · Repartir comida" {
		t.Errorf("contentLabel = %q", got)
	}
	if got := contentLabel(ModuleCategories, ""); got != "Categorías" {
		t.Errorf("contentLabel sin elemento = %q", got)
	}
	if got := imageLabel(ModuleEvents, "img_11111111-1111-1111-1111-111111111111.jpg"); got != "Eventos · Imagen · img_11111111-1111-1111-1111-111111111111.jpg" {
		t.Errorf("imageLabel = %q", got)
	}

	actor := uuid.New()
	action := contentAction(actor, ModuleEvents.CreateCode(), contentLabel(ModuleEvents, "X"))
	if action.Code != audit.ActionEventCreate || action.TargetKind != audit.TargetContent || action.Result != audit.ResultSuccess {
		t.Errorf("contentAction = %+v", action)
	}
	if action.ActorUserID == nil || *action.ActorUserID != actor {
		t.Errorf("contentAction sin actor: %+v", action)
	}
}

// TestRecordContentFailureIsBestEffort fija R4-10: el rechazo se registra con
// `result='failure'` y un fallo del registro NO cambia la respuesta (queda en el
// log).
func TestRecordContentFailureIsBestEffort(t *testing.T) {
	actor := uuid.New()

	t.Run("registra el rechazo", func(t *testing.T) {
		repo := &fakeRepository{}
		svc := newTestService(t, repo)
		svc.recordContentFailure(context.Background(), actor, ModuleEvents.CreateCode(), "Eventos · X")

		if len(repo.recorded) != 1 {
			t.Fatalf("acciones registradas = %d, se esperaba 1", len(repo.recorded))
		}
		action := repo.recorded[0]
		if action.Result != audit.ResultFailure || action.Code != audit.ActionEventCreate {
			t.Errorf("acción registrada = %+v", action)
		}
		if action.TargetKind != audit.TargetContent || action.TargetLabel != "Eventos · X" {
			t.Errorf("objetivo = %+v", action)
		}
	})

	t.Run("un fallo del registro no rompe", func(t *testing.T) {
		repo := &fakeRepository{recordErr: errors.New("base caída")}
		logger, capture := testutil.NewLogger()
		svc := NewService(ServiceDeps{
			Repository: repo,
			Location:   time.UTC,
			Logger:     logger,
		})
		svc.recordContentFailure(context.Background(), actor, ModuleEvents.CreateCode(), "Eventos · X")

		if capture.Count() == 0 {
			t.Error("el fallo best-effort debería quedar en el log")
		}
	})
}
