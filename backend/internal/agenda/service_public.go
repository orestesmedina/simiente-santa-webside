package agenda

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
)

// service_public.go implementa la cara pública de la agenda (T418): «lo que
// viene» (FR-001…FR-008). Resuelve el idioma en el servidor con fallback
// `en → es` por campo (R3-2/FR-008), lee SOLO las variantes públicas
// (`…Upcoming`: publicado **y** próximo, FR-013/SC-002/SC-009), ordena
// cronológicamente con desempate estable (FR-003), omite las claves de bloque
// vacías (SC-012) y acota cada bloque a `MaxBlockItems` (RG4-8).
//
// El cliente NO reimplementa el fallback ni el filtro: los DTOs de salida ya
// traen los valores resueltos y solo lo publicado y próximo.
//
// El DTO público NUNCA lleva `publicationState` (FR-013): el visitante no sabe
// si un elemento está en borrador; simplemente no existe para él.

// --- DTOs de salida pública (contrato `ProximamentePublica`) ---

// PublicProximamente es «lo que viene» publicado, ya resuelto a `lang`. Las
// claves de bloque están presentes **solo** si tienen elementos (SC-012) y, si
// están, sus arrays nunca son null (§8.1.2).
type PublicProximamente struct {
	Lang       string           `json:"lang"`
	Events     []PublicEvent    `json:"events,omitempty"`
	Activities []PublicActivity `json:"activities,omitempty"`
}

// PublicEvent es un evento publicado con al menos una fecha próxima: `Dates[0]`
// es "la próxima fecha" que el visitante identifica (FR-004).
type PublicEvent struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Frequency *string           `json:"frequency,omitempty"`
	Dates     []PublicOcurrence `json:"dates"`
	Place     string            `json:"place"`
	MapURL    *string           `json:"mapUrl,omitempty"`
	Detail    string            `json:"detail"`
	Organizer *string           `json:"organizer,omitempty"`
	Category  *PublicCategory   `json:"category,omitempty"`
	Image     *PublicImage      `json:"image,omitempty"`
}

// PublicOcurrence es una fecha concreta del evento con su horario opcional (Q2).
// Solo fechas próximas viajan en la respuesta pública (FR-002/FR-006).
type PublicOcurrence struct {
	Date      string  `json:"date"`
	StartTime *string `json:"startTime,omitempty"`
	EndTime   *string `json:"endTime,omitempty"`
}

// PublicActivity es una actividad publicada con fecha única (Decisión 3).
type PublicActivity struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Date      string          `json:"date"`
	StartTime *string         `json:"startTime,omitempty"`
	EndTime   *string         `json:"endTime,omitempty"`
	Place     string          `json:"place"`
	MapURL    *string         `json:"mapUrl,omitempty"`
	Detail    string          `json:"detail"`
	Organizer *string         `json:"organizer,omitempty"`
	Category  *PublicCategory `json:"category,omitempty"`
	Image     *PublicImage    `json:"image,omitempty"`
}

// PublicCategory es la categoría publicada del elemento, ya localizada (Q6).
type PublicCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PublicImage es la imagen del elemento con su texto alternativo accesible
// (FR-020).
type PublicImage struct {
	URL string `json:"url"`
	Alt string `json:"alt"`
}

// --- Lectura pública ---

// GetProximamente devuelve los próximos eventos y las próximas actividades
// publicadas, resueltos a `lang` (FR-001…FR-008). Un idioma fuera de `es|en` es
// `400 invalid` (details.lang). No hay ninguna vía pública a borradores ni a
// elementos vencidos: todas las lecturas usan las consultas `…Upcoming` con el
// corte de la regla R4-2.
func (s *service) GetProximamente(ctx context.Context, lang string) (PublicProximamente, error) {
	lang, err := normalizeLang(lang)
	if err != nil {
		return PublicProximamente{}, err
	}

	today, nowTime, err := s.cutoff()
	if err != nil {
		return PublicProximamente{}, err
	}

	events, err := s.repository.ListUpcomingEvents(ctx, today, nowTime, MaxBlockItems)
	if err != nil {
		return PublicProximamente{}, fmt.Errorf("leer los próximos eventos: %w", err)
	}
	activities, err := s.repository.ListUpcomingActivities(ctx, today, nowTime, MaxBlockItems)
	if err != nil {
		return PublicProximamente{}, fmt.Errorf("leer las próximas actividades: %w", err)
	}

	categories, err := s.categoryNames(ctx, events, activities)
	if err != nil {
		return PublicProximamente{}, err
	}

	publicEvents := make([]PublicEvent, 0, len(events))
	// Las fechas de TODOS los eventos del bloque se leen en UNA sola consulta
	// (por lote): la vista pública no puede hacer una consulta por evento.
	datesByEvent, err := s.upcomingDatesByEvent(ctx, events, today, nowTime)
	if err != nil {
		return PublicProximamente{}, err
	}
	for _, event := range events {
		dates := datesByEvent[event.ID]
		// Defensa: un evento sin ninguna fecha próxima no aparece (la consulta
		// pública ya lo filtra; FR-002/FR-013).
		if len(dates) == 0 {
			continue
		}
		publicEvents = append(publicEvents, localizeEvent(event, dates, categoryFor(categories, event.CategoryID), lang))
	}

	publicActivities := make([]PublicActivity, 0, len(activities))
	for _, activity := range activities {
		publicActivities = append(publicActivities, localizeActivity(activity, categoryFor(categories, activity.CategoryID), lang))
	}

	sortEvents(publicEvents)
	sortActivities(publicActivities)

	out := PublicProximamente{Lang: lang}
	if len(publicEvents) > 0 {
		out.Events = publicEvents
	}
	if len(publicActivities) > 0 {
		out.Activities = publicActivities
	}
	return out, nil
}

// --- Fechas del bloque ---

// upcomingDatesByEvent lee en UNA sola consulta las fechas próximas de todos los
// eventos del bloque (regla R4-2) y las agrupa por evento conservando el orden
// ascendente que devuelve la consulta. Sin eventos no se consulta la base.
func (s *service) upcomingDatesByEvent(ctx context.Context, events []Event, today, nowTime string) (map[uuid.UUID][]EventDate, error) {
	if len(events) == 0 {
		return nil, nil
	}
	dates, err := s.repository.ListUpcomingEventDatesForEvents(ctx, eventIDs(events), today, nowTime)
	if err != nil {
		return nil, fmt.Errorf("leer las fechas próximas de los eventos: %w", err)
	}
	return groupDatesByEvent(dates), nil
}

// groupDatesByEvent agrupa las fechas por evento conservando el orden en que
// llegan (ascendente por fecha y hora, que es el de las consultas): la primera
// de cada evento es "la próxima fecha" (FR-004).
func groupDatesByEvent(dates []EventDate) map[uuid.UUID][]EventDate {
	byEvent := make(map[uuid.UUID][]EventDate, len(dates))
	for _, date := range dates {
		byEvent[date.EventID] = append(byEvent[date.EventID], date)
	}
	return byEvent
}

// eventIDs devuelve los identificadores de una lista de eventos (para las
// consultas por lote).
func eventIDs(events []Event) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}
	return ids
}

// --- Categorías del bloque ---
// categoryNames resuelve el nombre de las categorías usadas por los elementos
// del bloque con UNA sola lectura del catálogo (Q6/R4-5): solo se consulta si
// algún elemento tiene categoría.
func (s *service) categoryNames(ctx context.Context, events []Event, activities []Activity) (map[uuid.UUID]Category, error) {
	needed := false
	for _, event := range events {
		needed = needed || event.CategoryID != nil
	}
	for _, activity := range activities {
		needed = needed || activity.CategoryID != nil
	}
	if !needed {
		return nil, nil
	}

	categories, err := s.repository.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("leer el catálogo de categorías: %w", err)
	}
	byID := make(map[uuid.UUID]Category, len(categories))
	for _, category := range categories {
		byID[category.ID] = category
	}
	return byID, nil
}

// categoryFor resuelve la categoría del elemento desde el índice del catálogo:
// nil cuando el elemento no tiene categoría o no está en el catálogo.
func categoryFor(categories map[uuid.UUID]Category, id *uuid.UUID) *Category {
	if id == nil {
		return nil
	}
	category, ok := categories[*id]
	if !ok {
		return nil
	}
	return &category
}

// --- Localización (fallback en → es por campo, R3-2) ---

// pickString resuelve un campo obligatorio en español: si se pide inglés y hay
// traducción, la usa; si no, devuelve el español (nunca vacío, SC-006).
func pickString(es string, en *string, lang string) string {
	if lang == LangEN && en != nil {
		return *en
	}
	return es
}

// pickOptional resuelve un campo opcional: prefiere la traducción pedida y cae al
// español; si tampoco hay español, devuelve nil (el campo se omite).
func pickOptional(es, en *string, lang string) *string {
	if lang == LangEN && en != nil {
		return en
	}
	return es
}

// localizeEvent resuelve el evento publicado al idioma pedido, con sus fechas
// próximas y su categoría e imagen (FR-004/FR-007/FR-020).
func localizeEvent(event Event, dates []EventDate, category *Category, lang string) PublicEvent {
	public := PublicEvent{
		ID:        event.ID.String(),
		Title:     pickString(event.TitleEs, event.TitleEn, lang),
		Frequency: pickOptional(event.FrequencyEs, event.FrequencyEn, lang),
		Place:     pickString(event.PlaceEs, event.PlaceEn, lang),
		MapURL:    event.MapURL,
		Detail:    pickString(event.DetailEs, event.DetailEn, lang),
		Organizer: pickOptional(event.OrganizerEs, event.OrganizerEn, lang),
		Category:  localizeCategory(category, lang),
		Image:     localizeImage(event.ImageFile, event.ImageAltEs, event.ImageAltEn, lang),
		Dates:     make([]PublicOcurrence, 0, len(dates)),
	}
	for _, date := range dates {
		public.Dates = append(public.Dates, PublicOcurrence{
			Date:      date.Date,
			StartTime: date.StartTime,
			EndTime:   date.EndTime,
		})
	}
	return public
}

// localizeActivity resuelve la actividad publicada al idioma pedido.
func localizeActivity(activity Activity, category *Category, lang string) PublicActivity {
	return PublicActivity{
		ID:        activity.ID.String(),
		Title:     pickString(activity.TitleEs, activity.TitleEn, lang),
		Date:      activity.Date,
		StartTime: activity.StartTime,
		EndTime:   activity.EndTime,
		Place:     pickString(activity.PlaceEs, activity.PlaceEn, lang),
		MapURL:    activity.MapURL,
		Detail:    pickString(activity.DetailEs, activity.DetailEn, lang),
		Organizer: pickOptional(activity.OrganizerEs, activity.OrganizerEn, lang),
		Category:  localizeCategory(category, lang),
		Image:     localizeImage(activity.ImageFile, activity.ImageAltEs, activity.ImageAltEn, lang),
	}
}

// localizeCategory resuelve la categoría (nil si el elemento no tiene).
func localizeCategory(category *Category, lang string) *PublicCategory {
	if category == nil {
		return nil
	}
	return &PublicCategory{
		ID:   category.ID.String(),
		Name: pickString(category.NameEs, category.NameEn, lang),
	}
}

// localizeImage arma la imagen pública con su URL relativa y su texto
// alternativo resuelto (FR-020). Sin archivo no hay imagen.
func localizeImage(file, altEs, altEn *string, lang string) *PublicImage {
	if file == nil {
		return nil
	}
	return &PublicImage{
		URL: MediaPathPrefix + *file,
		Alt: pickString(derefString(altEs), altEn, lang),
	}
}

// derefString devuelve el valor de un *string o "" si es nil.
func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// --- Orden (FR-003, edge case "dos elementos en la misma fecha") ---

// sortEvents ordena el bloque por la primera fecha próxima con desempate estable
// (`title`, `id`): dos elementos en la misma fecha salen siempre en el mismo
// orden. La consulta pública ya ordena (para el tope de `MaxBlockItems`); el
// service lo reafirma sobre los valores que ve el visitante.
func sortEvents(events []PublicEvent) {
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].Dates[0].Date != events[j].Dates[0].Date {
			return events[i].Dates[0].Date < events[j].Dates[0].Date
		}
		if events[i].Title != events[j].Title {
			return events[i].Title < events[j].Title
		}
		return events[i].ID < events[j].ID
	})
}

// sortActivities ordena el bloque por fecha, hora de inicio (sin hora primero) y
// desempate estable (`title`, `id`), como la consulta pública.
func sortActivities(activities []PublicActivity) {
	sort.SliceStable(activities, func(i, j int) bool {
		if activities[i].Date != activities[j].Date {
			return activities[i].Date < activities[j].Date
		}
		left, right := derefString(activities[i].StartTime), derefString(activities[j].StartTime)
		if left != right {
			return left < right
		}
		if activities[i].Title != activities[j].Title {
			return activities[i].Title < activities[j].Title
		}
		return activities[i].ID < activities[j].ID
	})
}

// --- Extracto «Lo que viene» de la portada (US7/FR-025…FR-027, Q8) ---

// Constantes del bloque de la portada (P4-18/R4-16). Son **decisiones firmes**
// del humano del 2026-10-12: no son parámetros de la API ni defaults
// configurables.
const (
	// maxPortadaItems es el tope del bloque «Lo que viene» (FR-025): nunca se
	// muestran más de 5 ítems y el bloque no se pagina.
	maxPortadaItems = 5
	// portadaWindowMonths es la ventana temporal del bloque (FR-025): mandan los
	// ítems de los próximos 2 meses de calendario; si dentro de ella no hay
	// ninguno, entra el respaldo (los siguientes próximos).
	portadaWindowMonths = 2
)

// PortadaProximamente es el extracto del bloque «Lo que viene» de la portada
// (contrato homónimo): `items` está presente **solo** si hay ítems (clave
// omitida cuando no hay nada publicado y próximo: la portada muestra entonces su
// mensaje, FR-026) y nunca es `null` (§8.1.2).
type PortadaProximamente struct {
	Lang  string               `json:"lang"`
	Items []ItemProximoPortada `json:"items,omitempty"`
}

// ItemProximoPortada es un ítem del bloque: solo lo que la tarjeta compacta
// muestra (FR-025) — concepto, título, próxima fecha con horario opcional, lugar
// e imagen con su texto alternativo—, ya resuelto a `lang` (FR-008).
type ItemProximoPortada struct {
	Kind      string       `json:"kind"`
	ID        string       `json:"id"`
	Title     string       `json:"title"`
	Date      string       `json:"date"`
	StartTime *string      `json:"startTime,omitempty"`
	EndTime   *string      `json:"endTime,omitempty"`
	Place     string       `json:"place"`
	Image     *PublicImage `json:"image,omitempty"`
}

// GetPortadaProximamente devuelve el extracto del bloque «Lo que viene» de la
// portada (US7/FR-025…FR-027): hasta `maxPortadaItems` ítems —eventos y
// actividades mezclados por cercanía— de lo publicado y próximo (misma regla
// R4-2 que el resto del público), preferentemente de los próximos
// `portadaWindowMonths` meses y con respaldo a los siguientes próximos cuando la
// ventana está vacía (nunca se mezclan). Sin ítems devuelve la respuesta con
// `lang` y sin `items` (la portada muestra el bloque con su mensaje, FR-026).
// Un idioma fuera de `es|en` es `400 invalid` (details.lang).
func (s *service) GetPortadaProximamente(ctx context.Context, lang string) (PortadaProximamente, error) {
	lang, err := normalizeLang(lang)
	if err != nil {
		return PortadaProximamente{}, err
	}

	// El corte de "hoy"/"ahora" y el final de la ventana se calculan en la zona
	// del lugar (APP_TZ) y viajan parametrizados al SQL (R4-2/R4-16).
	today, nowTime, err := s.cutoff()
	if err != nil {
		return PortadaProximamente{}, err
	}
	windowEnd, err := portadaWindowEnd(today, portadaWindowMonths)
	if err != nil {
		return PortadaProximamente{}, err
	}

	items, err := s.repository.ListNextUpcomingItems(ctx, today, nowTime, windowEnd, maxPortadaItems)
	if err != nil {
		return PortadaProximamente{}, fmt.Errorf("leer el extracto de la portada: %w", err)
	}

	out := PortadaProximamente{Lang: lang}
	if len(items) > 0 {
		out.Items = make([]ItemProximoPortada, 0, len(items))
		for _, item := range items {
			out.Items = append(out.Items, localizeNextItem(item, lang))
		}
	}
	return out, nil
}

// portadaWindowEnd suma meses de CALENDARIO al corte (p. ej. `2026-10-12` → con
// 2 meses → `2026-12-12`; no son 60 días) y devuelve la fecha de cierre de la
// ventana del bloque. El corte lo genera `cutoff` en `APP_TZ`, así que la
// aritmética es la del calendario del lugar.
func portadaWindowEnd(today string, months int) (string, error) {
	parsed, err := time.Parse(DateFormat, today)
	if err != nil {
		return "", apperr.Internal(fmt.Errorf("fecha de corte inválida %q: %w", today, err))
	}
	return parsed.AddDate(0, months, 0).Format(DateFormat), nil
}

// localizeNextItem resuelve el ítem del extracto al idioma pedido con fallback
// `en → es` por campo (FR-008) y compone su imagen (`/api/v1/media/<file>`) con
// el texto alternativo resuelto (FR-020).
func localizeNextItem(item NextItem, lang string) ItemProximoPortada {
	return ItemProximoPortada{
		Kind:      string(item.Kind),
		ID:        item.ID.String(),
		Title:     pickString(item.TitleEs, item.TitleEn, lang),
		Date:      item.Date,
		StartTime: item.StartTime,
		EndTime:   item.EndTime,
		Place:     pickString(item.PlaceEs, item.PlaceEn, lang),
		Image:     localizeImage(item.ImageFile, item.ImageAltEs, item.ImageAltEn, lang),
	}
}

// normalizeLang normaliza y valida el idioma pedido por la superficie pública:
// vacío es el idioma base (`es`) y cualquier otro valor fuera de `es|en` es un
// `400 invalid` con `details.lang`.
func normalizeLang(lang string) (string, error) {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		lang = LangES
	}
	if !validLang(lang) {
		return "", apperr.Invalid(
			"El idioma solicitado no es válido",
			apperr.WithDetails(map[string]any{"lang": "solo se admite es o en"}),
		)
	}
	return lang, nil
}
