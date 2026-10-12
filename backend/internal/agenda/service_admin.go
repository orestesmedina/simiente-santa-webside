package agenda

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/storage"
	"simiente-santa/backend/internal/platform/validate"
)

// service_admin.go implementa la gestión de los módulos desde el panel
// (T419–T421): eventos con su lista de fechas, actividades de fecha única y el
// catálogo de categorías. No conoce HTTP ni SQL: valida (FR-015), normaliza
// (analyze I6), aplica los límites, publica al guardar (FR-014), detecta
// duplicados (R4-9) y compone las filas de auditoría con la regla exacta de
// códigos (R4-10), que el repository inserta en la MISMA transacción (FR-017).

// --- DTOs del panel (contrato `EventoAdmin`, `EventoAdminList`) ---

// EventoAdmin es el evento completo del panel: ambos idiomas, todas sus fechas
// (próximas y pasadas), su estado de publicación y las señales del panel
// `nextDate` (siempre presente, nullable) e `isPast` (FR-010, Q4).
type EventoAdmin struct {
	ID               string       `json:"id"`
	TitleEs          string       `json:"titleEs"`
	TitleEn          *string      `json:"titleEn"`
	FrequencyEs      *string      `json:"frequencyEs"`
	FrequencyEn      *string      `json:"frequencyEn"`
	DetailEs         string       `json:"detailEs"`
	DetailEn         *string      `json:"detailEn"`
	PlaceEs          string       `json:"placeEs"`
	PlaceEn          *string      `json:"placeEn"`
	MapURL           *string      `json:"mapUrl"`
	OrganizerEs      *string      `json:"organizerEs"`
	OrganizerEn      *string      `json:"organizerEn"`
	CategoryID       *string      `json:"categoryId"`
	ImageFile        *string      `json:"imageFile"`
	ImageAltEs       *string      `json:"imageAltEs"`
	ImageAltEn       *string      `json:"imageAltEn"`
	PublicationState string       `json:"publicationState"`
	Dates            []FechaAdmin `json:"dates"`
	// NextDate es la primera fecha próxima (YYYY-MM-DD) o nil cuando el evento
	// no tiene ninguna: el panel señala entonces que falta cargar la próxima
	// fecha. Va SIN `omitempty`: siempre está en la respuesta (su valor puede
	// ser null).
	NextDate  *string   `json:"nextDate"`
	IsPast    bool      `json:"isPast"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// FechaAdmin es una fecha concreta del evento en el panel, con su horario
// opcional y su marca de pasado (Q4).
type FechaAdmin struct {
	ID        string  `json:"id"`
	Date      string  `json:"date"`
	StartTime *string `json:"startTime,omitempty"`
	EndTime   *string `json:"endTime,omitempty"`
	IsPast    bool    `json:"isPast"`
}

// EventoAdminList es el listado paginado del panel (§8.1.2/§8.1.3).
type EventoAdminList struct {
	Items  []EventoAdmin `json:"items"`
	Total  int64         `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// ImageUploadResult es el resultado de una subida (contrato `ImageUploadResult`):
// el nombre generado por el servidor y su URL pública relativa.
type ImageUploadResult struct {
	FileName  string `json:"fileName"`
	URL       string `json:"url"`
	MimeType  string `json:"mimeType"`
	SizeBytes int64  `json:"sizeBytes"`
}

// --- DTO de entrada del PATCH (contrato `EventoPatch`) ---

// EventoPatch es el cuerpo de PATCH /api/v1/admin/eventos/{id}: solo los campos
// presentes se tocan. Los campos anulables usan Optional[T] para distinguir el
// `null` explícito (que vacía el campo) del campo ausente; `dates` reemplaza la
// lista COMPLETA de fechas (R4-11).
type EventoPatch struct {
	TitleEs          *string          `json:"titleEs,omitempty"`
	TitleEn          Optional[string] `json:"titleEn"`
	FrequencyEs      Optional[string] `json:"frequencyEs"`
	FrequencyEn      Optional[string] `json:"frequencyEn"`
	DetailEs         *string          `json:"detailEs,omitempty"`
	DetailEn         Optional[string] `json:"detailEn"`
	PlaceEs          *string          `json:"placeEs,omitempty"`
	PlaceEn          Optional[string] `json:"placeEn"`
	MapURL           Optional[string] `json:"mapUrl"`
	OrganizerEs      Optional[string] `json:"organizerEs"`
	OrganizerEn      Optional[string] `json:"organizerEn"`
	CategoryID       Optional[string] `json:"categoryId"`
	ImageFile        Optional[string] `json:"imageFile"`
	ImageAltEs       Optional[string] `json:"imageAltEs"`
	ImageAltEn       Optional[string] `json:"imageAltEn"`
	PublicationState *string          `json:"publicationState,omitempty"`
	Dates            *[]FechaInput    `json:"dates,omitempty"`
}

// hasChanges indica si el PATCH trae al menos un campo (minProperties: 1). Un
// `null` explícito cuenta como cambio: limpiar un campo opcional lo es.
func (p EventoPatch) hasChanges() bool {
	return p.TitleEs != nil || p.TitleEn.Set() || p.FrequencyEs.Set() || p.FrequencyEn.Set() ||
		p.DetailEs != nil || p.DetailEn.Set() || p.PlaceEs != nil || p.PlaceEn.Set() ||
		p.MapURL.Set() || p.OrganizerEs.Set() || p.OrganizerEn.Set() || p.CategoryID.Set() ||
		p.ImageFile.Set() || p.ImageAltEs.Set() || p.ImageAltEn.Set() ||
		p.PublicationState != nil || p.Dates != nil
}

// --- Eventos (FR-010) ---

// CreateEvento da de alta un evento con su lista de fechas (FR-010/FR-018):
// valida FR-015 completo, normaliza los `*En` vacíos a NULL (analyze I6), rechaza
// el duplicado del módulo con 409 (R4-9) y registra `event.create` —aunque el
// evento nazca publicado (R4-10)— en la misma transacción que el alta (FR-017).
func (s *service) CreateEvento(ctx context.Context, actorID uuid.UUID, in EventoInput) (EventoAdmin, error) {
	fail := func(err error) (EventoAdmin, error) {
		s.recordContentFailure(ctx, actorID, ModuleEvents.CreateCode(), contentLabel(ModuleEvents, ""))
		return EventoAdmin{}, err
	}

	if err := validate.Struct(in); err != nil {
		return fail(err)
	}
	dates, err := normalizeEventDates(in.Dates)
	if err != nil {
		return fail(err)
	}
	if err := validateImage(in.ImageFile, in.ImageAltEs); err != nil {
		return fail(err)
	}
	categoryID, err := s.resolveCategoryID(ctx, in.CategoryID)
	if err != nil {
		return fail(err)
	}

	event := Event{
		TitleEs:          normalizeText(in.TitleEs),
		TitleEn:          normalizeOptional(in.TitleEn),
		FrequencyEs:      normalizeOptional(in.FrequencyEs),
		FrequencyEn:      normalizeOptional(in.FrequencyEn),
		DetailEs:         normalizeText(in.DetailEs),
		DetailEn:         normalizeOptional(in.DetailEn),
		PlaceEs:          normalizeText(in.PlaceEs),
		PlaceEn:          normalizeOptional(in.PlaceEn),
		MapURL:           normalizeOptional(in.MapURL),
		OrganizerEs:      normalizeOptional(in.OrganizerEs),
		OrganizerEn:      normalizeOptional(in.OrganizerEn),
		CategoryID:       categoryID,
		ImageFile:        normalizeOptional(in.ImageFile),
		ImageAltEs:       normalizeOptional(in.ImageAltEs),
		ImageAltEn:       normalizeOptional(in.ImageAltEn),
		PublicationState: publicationStateOr(in.PublicationState),
	}

	duplicate, err := s.repository.ExistsEventDuplicate(ctx, event.TitleEs, event.PlaceEs, datesStrings(dates), nil)
	if err != nil {
		return EventoAdmin{}, fmt.Errorf("comprobar el duplicado del evento: %w", err)
	}
	if duplicate {
		return fail(apperr.Conflict(
			"Ya existe un evento con el mismo título, lugar y una fecha en común",
			apperr.WithDetails(map[string]any{"dates": "cambia el título, el lugar o alguna fecha"}),
		))
	}

	label := contentLabel(ModuleEvents, event.TitleEs)
	actions := []audit.Action{contentAction(actorID, ModuleEvents.CreateCode(), label)}
	saved, err := s.repository.InsertEvent(ctx, event, dates, actions...)
	if err != nil {
		return EventoAdmin{}, fmt.Errorf("crear el evento: %w", err)
	}
	return s.eventoAdmin(ctx, saved)
}

// UpdateEvento edita parcialmente un evento (FR-010). Registra `event.update` si
// cambian los datos y `event.publish`/`event.unpublish` si cambia el estado
// —dos filas cuando ocurren ambas (R4-10)— en la misma transacción. Un id
// inexistente → 404; el duplicado del módulo (sin contarse a sí mismo) → 409.
func (s *service) UpdateEvento(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch EventoPatch) (EventoAdmin, error) {
	fail := func(err error) (EventoAdmin, error) {
		s.recordContentFailure(ctx, actorID, ModuleEvents.UpdateCode(), contentLabel(ModuleEvents, id.String()))
		return EventoAdmin{}, err
	}

	if !patch.hasChanges() {
		return fail(apperr.Invalid("No hay cambios que guardar"))
	}

	current, found, err := s.repository.GetEventByID(ctx, id)
	if err != nil {
		return EventoAdmin{}, fmt.Errorf("leer el evento: %w", err)
	}
	if !found {
		return fail(apperr.NotFound("El evento no existe"))
	}

	updated, dates, dataChanged, stateChanged, err := s.mergeEvento(ctx, current, patch)
	if err != nil {
		return fail(err)
	}
	if !dataChanged && !stateChanged {
		return s.eventoAdmin(ctx, current)
	}

	if dataChanged {
		// Fechas efectivas: las del PATCH o las que ya tenía el evento.
		effective := dates
		if effective == nil {
			stored, err := s.repository.ListEventDates(ctx, current.ID)
			if err != nil {
				return EventoAdmin{}, fmt.Errorf("leer las fechas del evento: %w", err)
			}
			effective = stored
		}
		duplicate, err := s.repository.ExistsEventDuplicate(ctx, updated.TitleEs, updated.PlaceEs, datesStrings(effective), &current.ID)
		if err != nil {
			return EventoAdmin{}, fmt.Errorf("comprobar el duplicado del evento: %w", err)
		}
		if duplicate {
			return fail(apperr.Conflict(
				"Ya existe un evento con el mismo título, lugar y una fecha en común",
				apperr.WithDetails(map[string]any{"dates": "cambia el título, el lugar o alguna fecha"}),
			))
		}
	}

	label := contentLabel(ModuleEvents, updated.TitleEs)
	actions := make([]audit.Action, 0, 2)
	if dataChanged {
		actions = append(actions, contentAction(actorID, ModuleEvents.UpdateCode(), label))
	}
	if stateChanged {
		actions = append(actions, contentAction(actorID, ModuleEvents.StateCode(updated.PublicationState), label))
	}

	saved, err := s.repository.UpdateEvent(ctx, updated, dates, actions...)
	if err != nil {
		return EventoAdmin{}, fmt.Errorf("editar el evento: %w", err)
	}
	return s.eventoAdmin(ctx, saved)
}

// DeleteEvento elimina físicamente un evento (con sus fechas) y registra
// `event.delete` con su `targetLabel` conservado (R4-10) en la misma
// transacción. Un id inexistente → 404.
func (s *service) DeleteEvento(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	current, found, err := s.repository.GetEventByID(ctx, id)
	if err != nil {
		return fmt.Errorf("leer el evento: %w", err)
	}
	if !found {
		return apperr.NotFound("El evento no existe")
	}
	actions := []audit.Action{contentAction(actorID, ModuleEvents.DeleteCode(), contentLabel(ModuleEvents, current.TitleEs))}
	deleted, err := s.repository.DeleteEvent(ctx, id, actions...)
	if err != nil {
		return fmt.Errorf("eliminar el evento: %w", err)
	}
	if !deleted {
		return apperr.NotFound("El evento no existe")
	}
	return nil
}

// GetEvento devuelve la ficha completa de un evento para el formulario (FR-010):
// ambos idiomas, todas sus fechas y su estado, con `nextDate`/`isPast`. Un id
// inexistente → 404.
func (s *service) GetEvento(ctx context.Context, id uuid.UUID) (EventoAdmin, error) {
	event, found, err := s.repository.GetEventByID(ctx, id)
	if err != nil {
		return EventoAdmin{}, fmt.Errorf("leer el evento: %w", err)
	}
	if !found {
		return EventoAdmin{}, apperr.NotFound("El evento no existe")
	}
	return s.eventoAdmin(ctx, event)
}

// ListEventos devuelve el listado paginado del panel (FR-010): todos los eventos
// (borradores y con cualquier fecha incluidos), ordenados por
// `created_at DESC, id DESC` y con filtro opcional por estado.
func (s *service) ListEventos(ctx context.Context, state *PublicationState, params paginate.Params) (EventoAdminList, error) {
	total, err := s.repository.CountEvents(ctx, state)
	if err != nil {
		return EventoAdminList{}, fmt.Errorf("contar los eventos: %w", err)
	}
	events, err := s.repository.ListEvents(ctx, state, params)
	if err != nil {
		return EventoAdminList{}, fmt.Errorf("listar los eventos: %w", err)
	}

	today, nowTime, err := s.cutoff()
	if err != nil {
		return EventoAdminList{}, err
	}
	// Las fechas de TODA la página se leen en UNA sola consulta (por lote): el
	// listado no puede hacer una consulta por fila.
	datesByEvent, err := s.datesByEvent(ctx, events)
	if err != nil {
		return EventoAdminList{}, err
	}

	items := make([]EventoAdmin, 0, len(events))
	for _, event := range events {
		items = append(items, eventoAdminFrom(event, datesByEvent[event.ID], today, nowTime))
	}
	return EventoAdminList{Items: items, Total: total, Limit: params.Limit, Offset: params.Offset}, nil
}

// datesByEvent lee en UNA sola consulta las fechas (todas, próximas y pasadas)
// de los eventos de la página y las agrupa por evento. Sin eventos no se
// consulta la base.
func (s *service) datesByEvent(ctx context.Context, events []Event) (map[uuid.UUID][]EventDate, error) {
	if len(events) == 0 {
		return nil, nil
	}
	dates, err := s.repository.ListEventDatesForEvents(ctx, eventIDs(events))
	if err != nil {
		return nil, fmt.Errorf("leer las fechas de los eventos: %w", err)
	}
	return groupDatesByEvent(dates), nil
}

// UploadEventoImage guarda una imagen subida desde el panel para el módulo de
// eventos (FR-020/R4-6) y la audita con `event.image.upload` (R4-10) de forma
// **fail-closed**: si el registro falla, se elimina el archivo y se devuelve
// error, de modo que nunca queda una subida sin traza.
func (s *service) UploadEventoImage(ctx context.Context, actorID uuid.UUID, data []byte) (ImageUploadResult, error) {
	return s.uploadImage(ctx, actorID, ModuleEvents, data)
}

// eventoAdmin arma la ficha del panel de un evento leyendo sus fechas.
func (s *service) eventoAdmin(ctx context.Context, event Event) (EventoAdmin, error) {
	dates, err := s.repository.ListEventDates(ctx, event.ID)
	if err != nil {
		return EventoAdmin{}, fmt.Errorf("leer las fechas del evento: %w", err)
	}
	today, nowTime, err := s.cutoff()
	if err != nil {
		return EventoAdmin{}, err
	}
	return eventoAdminFrom(event, dates, today, nowTime), nil
}

// eventoAdminFrom traduce el evento y sus fechas al DTO del panel: marca cada
// fecha con `isPast` y calcula `nextDate` (la primera próxima) sin tocar el
// estado del elemento (FR-006/Q4).
func eventoAdminFrom(event Event, dates []EventDate, today, nowTime string) EventoAdmin {
	admin := EventoAdmin{
		ID:               event.ID.String(),
		TitleEs:          event.TitleEs,
		TitleEn:          event.TitleEn,
		FrequencyEs:      event.FrequencyEs,
		FrequencyEn:      event.FrequencyEn,
		DetailEs:         event.DetailEs,
		DetailEn:         event.DetailEn,
		PlaceEs:          event.PlaceEs,
		PlaceEn:          event.PlaceEn,
		MapURL:           event.MapURL,
		OrganizerEs:      event.OrganizerEs,
		OrganizerEn:      event.OrganizerEn,
		ImageFile:        event.ImageFile,
		ImageAltEs:       event.ImageAltEs,
		ImageAltEn:       event.ImageAltEn,
		PublicationState: string(event.PublicationState),
		Dates:            make([]FechaAdmin, 0, len(dates)),
		CreatedAt:        event.CreatedAt,
		UpdatedAt:        event.UpdatedAt,
	}
	if event.CategoryID != nil {
		id := event.CategoryID.String()
		admin.CategoryID = &id
	}
	for _, date := range dates {
		upcoming := isUpcoming(date.Date, today, nowTime, date.EndTime)
		admin.Dates = append(admin.Dates, FechaAdmin{
			ID:        date.ID.String(),
			Date:      date.Date,
			StartTime: date.StartTime,
			EndTime:   date.EndTime,
			IsPast:    !upcoming,
		})
		if upcoming && (admin.NextDate == nil || date.Date < *admin.NextDate) {
			next := date.Date
			admin.NextDate = &next
		}
	}
	admin.IsPast = admin.NextDate == nil
	return admin
}

// --- Mezcla y validación del PATCH de eventos ---

// mergeEvento combina el evento actual con los campos presentes del PATCH y
// valida el resultado (FR-015). Devuelve la lista de fechas nueva (nil si el
// PATCH no traía `dates`: no se toca) y si cambiaron los datos y/o el estado.
func (s *service) mergeEvento(ctx context.Context, current Event, patch EventoPatch) (Event, []EventDate, bool, bool, error) {
	updated := current
	dataChanged := false
	invalid := func(field, message string) error {
		return apperr.Invalid("Revisa los datos del formulario", apperr.WithDetails(map[string]any{field: message}))
	}

	if patch.TitleEs != nil {
		title := normalizeText(*patch.TitleEs)
		if title == "" {
			return Event{}, nil, false, false, invalid("titleEs", "Este campo es obligatorio.")
		}
		if len([]rune(title)) > MaxTitleLength {
			return Event{}, nil, false, false, invalid("titleEs", fmt.Sprintf("No puede tener más de %d caracteres.", MaxTitleLength))
		}
		updated.TitleEs = title
		dataChanged = true
	}
	if patch.TitleEn.Set() {
		updated.TitleEn = normalizeOptional(patch.TitleEn.Value())
		dataChanged = true
	}
	if patch.FrequencyEs.Set() {
		updated.FrequencyEs = normalizeOptional(patch.FrequencyEs.Value())
		dataChanged = true
	}
	if patch.FrequencyEn.Set() {
		updated.FrequencyEn = normalizeOptional(patch.FrequencyEn.Value())
		dataChanged = true
	}
	if patch.DetailEs != nil {
		detail := normalizeText(*patch.DetailEs)
		if detail == "" {
			return Event{}, nil, false, false, invalid("detailEs", "Este campo es obligatorio.")
		}
		if len([]rune(detail)) > MaxDetailLength {
			return Event{}, nil, false, false, invalid("detailEs", fmt.Sprintf("No puede tener más de %d caracteres.", MaxDetailLength))
		}
		updated.DetailEs = detail
		dataChanged = true
	}
	if patch.DetailEn.Set() {
		updated.DetailEn = normalizeOptional(patch.DetailEn.Value())
		dataChanged = true
	}
	if patch.PlaceEs != nil {
		place := normalizeText(*patch.PlaceEs)
		if place == "" {
			return Event{}, nil, false, false, invalid("placeEs", "Este campo es obligatorio.")
		}
		if len([]rune(place)) > MaxPlaceLength {
			return Event{}, nil, false, false, invalid("placeEs", fmt.Sprintf("No puede tener más de %d caracteres.", MaxPlaceLength))
		}
		updated.PlaceEs = place
		dataChanged = true
	}
	if patch.PlaceEn.Set() {
		updated.PlaceEn = normalizeOptional(patch.PlaceEn.Value())
		dataChanged = true
	}
	if patch.MapURL.Set() {
		mapURL := normalizeOptional(patch.MapURL.Value())
		if mapURL != nil {
			rule := urlRule{URL: *mapURL}
			if err := validate.Struct(rule); err != nil {
				return Event{}, nil, false, false, apperr.Invalid(
					"El enlace del mapa no es válido",
					apperr.WithDetails(map[string]any{"mapUrl": "usa un enlace https"}),
				)
			}
		}
		updated.MapURL = mapURL
		dataChanged = true
	}
	if patch.OrganizerEs.Set() {
		updated.OrganizerEs = normalizeOptional(patch.OrganizerEs.Value())
		dataChanged = true
	}
	if patch.OrganizerEn.Set() {
		updated.OrganizerEn = normalizeOptional(patch.OrganizerEn.Value())
		dataChanged = true
	}
	if patch.CategoryID.Set() {
		categoryID, err := s.resolveCategoryID(ctx, patch.CategoryID.Value())
		if err != nil {
			return Event{}, nil, false, false, err
		}
		updated.CategoryID = categoryID
		dataChanged = true
	}
	if patch.ImageFile.Set() {
		updated.ImageFile = normalizeOptional(patch.ImageFile.Value())
		dataChanged = true
	}
	if patch.ImageAltEs.Set() {
		updated.ImageAltEs = normalizeOptional(patch.ImageAltEs.Value())
		dataChanged = true
	}
	if patch.ImageAltEn.Set() {
		updated.ImageAltEn = normalizeOptional(patch.ImageAltEn.Value())
		dataChanged = true
	}

	if err := validateImage(derefString(updated.ImageFile), derefString(updated.ImageAltEs)); err != nil {
		return Event{}, nil, false, false, err
	}

	var dates []EventDate
	if patch.Dates != nil {
		normalized, err := normalizeEventDates(*patch.Dates)
		if err != nil {
			return Event{}, nil, false, false, err
		}
		dates = normalized
		dataChanged = true
	}

	stateChanged := false
	if patch.PublicationState != nil {
		state := PublicationState(*patch.PublicationState)
		if !state.Valid() {
			return Event{}, nil, false, false, invalid("publicationState", "solo se admite draft o published")
		}
		updated.PublicationState = state
		stateChanged = state != current.PublicationState
	}
	return updated, dates, dataChanged, stateChanged, nil
}

// --- Categorías y subidas compartidas ---

// resolveCategoryID valida la categoría indicada (Q6): vacío → sin categoría;
// un identificador inválido o desconocido → 400 invalid con details.categoryId.
func (s *service) resolveCategoryID(ctx context.Context, raw string) (*uuid.UUID, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil, apperr.Invalid(
			"La categoría no es válida",
			apperr.WithDetails(map[string]any{"categoryId": "no es un identificador válido"}),
		)
	}
	_, found, err := s.repository.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("leer la categoría: %w", err)
	}
	if !found {
		return nil, apperr.Invalid(
			"La categoría no existe",
			apperr.WithDetails(map[string]any{"categoryId": "elige una categoría del catálogo"}),
		)
	}
	return &id, nil
}

// uploadImage guarda una imagen subida desde el panel para el módulo indicado
// (FR-020): el Store valida la firma binaria (JPEG/PNG/WebP) y el tamaño, y
// genera el nombre `img_<uuid>.<ext>`. La subida se audita con el código del
// módulo (`*.image.upload`) de forma **fail-closed** (analyze I8/R4-10): si el
// registro falla, se elimina el archivo y se devuelve error.
func (s *service) uploadImage(ctx context.Context, actorID uuid.UUID, module Module, data []byte) (ImageUploadResult, error) {
	if s.store == nil {
		return ImageUploadResult{}, apperr.Internal(errors.New("agenda: almacén de imágenes no configurado"))
	}
	if len(data) == 0 {
		return ImageUploadResult{}, apperr.Invalid(
			"No se recibió ninguna imagen",
			apperr.WithDetails(map[string]any{"file": "adjunta un archivo de imagen"}),
		)
	}

	file, err := s.store.Save(ctx, data)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrTooLarge):
			return ImageUploadResult{}, apperr.Invalid(
				"La imagen es demasiado grande",
				apperr.WithDetails(map[string]any{"fileSize": "la imagen supera el tamaño máximo permitido"}),
			)
		case errors.Is(err, storage.ErrInvalidFormat):
			return ImageUploadResult{}, apperr.Invalid(
				"El formato de la imagen no es válido",
				apperr.WithDetails(map[string]any{"file": "solo se admiten imágenes JPEG, PNG o WebP"}),
			)
		default:
			return ImageUploadResult{}, fmt.Errorf("guardar la imagen: %w", err)
		}
	}

	action := contentAction(actorID, module.ImageCode(), imageLabel(module, file.Name))
	if err := s.repository.RecordAction(ctx, action); err != nil {
		// Fail-closed: sin registro no hay subida (analyze I8).
		s.deleteFile(ctx, file.Name)
		return ImageUploadResult{}, fmt.Errorf("registrar la subida de la imagen: %w", err)
	}

	return ImageUploadResult{
		FileName:  file.Name,
		URL:       MediaPathPrefix + file.Name,
		MimeType:  file.ContentType,
		SizeBytes: file.SizeBytes,
	}, nil
}

// deleteFile borra best-effort una imagen (R3-8/R4-6): un fallo se queda en el
// log y no cambia la respuesta.
func (s *service) deleteFile(ctx context.Context, name string) {
	if s.store == nil || name == "" {
		return
	}
	if err := s.store.Delete(ctx, name); err != nil {
		s.logError(ctx, "no se pudo borrar la imagen", err)
	}
}

// datesStrings devuelve las fechas de una lista como texto (para el chequeo de
// duplicado del módulo, R4-9).
func datesStrings(dates []EventDate) []string {
	out := make([]string, 0, len(dates))
	for _, date := range dates {
		out = append(out, date.Date)
	}
	return out
}

// urlRule envuelve la etiqueta `url` de platform/validate para reutilizarla
// desde el service (la forma del enlace de mapa, FR-015).
type urlRule struct {
	URL string `validate:"required,url,max=500"`
}

// --- Actividades (FR-011, T420) ---

// ActividadAdmin es la actividad completa del panel: ambos idiomas, su fecha
// única con horario opcional, su estado de publicación y su marca de pasado
// (FR-011/Q4). No tiene `nextDate`: la actividad es de fecha única (Decisión 3).
type ActividadAdmin struct {
	ID               string    `json:"id"`
	TitleEs          string    `json:"titleEs"`
	TitleEn          *string   `json:"titleEn"`
	DetailEs         string    `json:"detailEs"`
	DetailEn         *string   `json:"detailEn"`
	PlaceEs          string    `json:"placeEs"`
	PlaceEn          *string   `json:"placeEn"`
	MapURL           *string   `json:"mapUrl"`
	OrganizerEs      *string   `json:"organizerEs"`
	OrganizerEn      *string   `json:"organizerEn"`
	CategoryID       *string   `json:"categoryId"`
	ImageFile        *string   `json:"imageFile"`
	ImageAltEs       *string   `json:"imageAltEs"`
	ImageAltEn       *string   `json:"imageAltEn"`
	Date             string    `json:"date"`
	StartTime        *string   `json:"startTime,omitempty"`
	EndTime          *string   `json:"endTime,omitempty"`
	PublicationState string    `json:"publicationState"`
	IsPast           bool      `json:"isPast"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// ActividadAdminList es el listado paginado de actividades (§8.1.2/§8.1.3).
type ActividadAdminList struct {
	Items  []ActividadAdmin `json:"items"`
	Total  int64            `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

// ActividadPatch es el cuerpo de PATCH /api/v1/admin/actividades/{id}: solo los
// campos presentes se tocan (con `null` explícito para vaciar los opcionales).
type ActividadPatch struct {
	TitleEs          *string          `json:"titleEs,omitempty"`
	TitleEn          Optional[string] `json:"titleEn"`
	DetailEs         *string          `json:"detailEs,omitempty"`
	DetailEn         Optional[string] `json:"detailEn"`
	PlaceEs          *string          `json:"placeEs,omitempty"`
	PlaceEn          Optional[string] `json:"placeEn"`
	MapURL           Optional[string] `json:"mapUrl"`
	OrganizerEs      Optional[string] `json:"organizerEs"`
	OrganizerEn      Optional[string] `json:"organizerEn"`
	CategoryID       Optional[string] `json:"categoryId"`
	ImageFile        Optional[string] `json:"imageFile"`
	ImageAltEs       Optional[string] `json:"imageAltEs"`
	ImageAltEn       Optional[string] `json:"imageAltEn"`
	Date             *string          `json:"date,omitempty"`
	StartTime        Optional[string] `json:"startTime"`
	EndTime          Optional[string] `json:"endTime"`
	PublicationState *string          `json:"publicationState,omitempty"`
}

// hasChanges indica si el PATCH trae al menos un campo.
func (p ActividadPatch) hasChanges() bool {
	return p.TitleEs != nil || p.TitleEn.Set() || p.DetailEs != nil || p.DetailEn.Set() ||
		p.PlaceEs != nil || p.PlaceEn.Set() || p.MapURL.Set() || p.OrganizerEs.Set() ||
		p.OrganizerEn.Set() || p.CategoryID.Set() || p.ImageFile.Set() || p.ImageAltEs.Set() ||
		p.ImageAltEn.Set() || p.Date != nil || p.StartTime.Set() || p.EndTime.Set() ||
		p.PublicationState != nil
}

// CreateActividad da de alta una actividad puntual (FR-011/FR-019): una única
// fecha con horario opcional, validación FR-015, normalización analyze I6 y
// registro `activity.create` —aunque nazca publicada (R4-10)— en la misma
// transacción. El duplicado exacto del módulo → 409 (R4-9); un EVENTO con los
// mismos datos nunca es duplicado (Decisión 3).
func (s *service) CreateActividad(ctx context.Context, actorID uuid.UUID, in ActividadInput) (ActividadAdmin, error) {
	fail := func(err error) (ActividadAdmin, error) {
		s.recordContentFailure(ctx, actorID, ModuleActivities.CreateCode(), contentLabel(ModuleActivities, ""))
		return ActividadAdmin{}, err
	}

	if err := validate.Struct(in); err != nil {
		return fail(err)
	}
	date, start, end, err := normalizeActivityDate(in.Date, in.StartTime, in.EndTime)
	if err != nil {
		return fail(err)
	}
	if err := validateImage(in.ImageFile, in.ImageAltEs); err != nil {
		return fail(err)
	}
	categoryID, err := s.resolveCategoryID(ctx, in.CategoryID)
	if err != nil {
		return fail(err)
	}

	activity := Activity{
		TitleEs:          normalizeText(in.TitleEs),
		TitleEn:          normalizeOptional(in.TitleEn),
		DetailEs:         normalizeText(in.DetailEs),
		DetailEn:         normalizeOptional(in.DetailEn),
		PlaceEs:          normalizeText(in.PlaceEs),
		PlaceEn:          normalizeOptional(in.PlaceEn),
		MapURL:           normalizeOptional(in.MapURL),
		OrganizerEs:      normalizeOptional(in.OrganizerEs),
		OrganizerEn:      normalizeOptional(in.OrganizerEn),
		CategoryID:       categoryID,
		ImageFile:        normalizeOptional(in.ImageFile),
		ImageAltEs:       normalizeOptional(in.ImageAltEs),
		ImageAltEn:       normalizeOptional(in.ImageAltEn),
		Date:             date,
		StartTime:        start,
		EndTime:          end,
		PublicationState: publicationStateOr(in.PublicationState),
	}

	label := contentLabel(ModuleActivities, activity.TitleEs)
	actions := []audit.Action{contentAction(actorID, ModuleActivities.CreateCode(), label)}
	saved, err := s.repository.InsertActivity(ctx, activity, actions...)
	if err != nil {
		// El duplicado exacto del módulo lo detecta la base (UNIQUE): su rechazo
		// también queda en la auditoría (FR-017/R4-10).
		if isConflict(err) {
			s.recordContentFailure(ctx, actorID, ModuleActivities.CreateCode(), label)
		}
		return ActividadAdmin{}, fmt.Errorf("crear la actividad: %w", err)
	}
	return s.actividadAdmin(ctx, saved)
}

// UpdateActividad edita parcialmente una actividad (FR-011) con la misma regla
// de publicación que los eventos (FR-013/FR-014) y la regla de códigos de
// R4-10 (`activity.update` y/o `activity.publish`/`activity.unpublish`).
func (s *service) UpdateActividad(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch ActividadPatch) (ActividadAdmin, error) {
	fail := func(err error) (ActividadAdmin, error) {
		s.recordContentFailure(ctx, actorID, ModuleActivities.UpdateCode(), contentLabel(ModuleActivities, id.String()))
		return ActividadAdmin{}, err
	}

	if !patch.hasChanges() {
		return fail(apperr.Invalid("No hay cambios que guardar"))
	}

	current, found, err := s.repository.GetActivityByID(ctx, id)
	if err != nil {
		return ActividadAdmin{}, fmt.Errorf("leer la actividad: %w", err)
	}
	if !found {
		return fail(apperr.NotFound("La actividad no existe"))
	}

	updated, dataChanged, stateChanged, err := s.mergeActividad(ctx, current, patch)
	if err != nil {
		return fail(err)
	}
	if !dataChanged && !stateChanged {
		return s.actividadAdmin(ctx, current)
	}

	label := contentLabel(ModuleActivities, updated.TitleEs)
	actions := make([]audit.Action, 0, 2)
	if dataChanged {
		actions = append(actions, contentAction(actorID, ModuleActivities.UpdateCode(), label))
	}
	if stateChanged {
		actions = append(actions, contentAction(actorID, ModuleActivities.StateCode(updated.PublicationState), label))
	}

	saved, err := s.repository.UpdateActivity(ctx, updated, actions...)
	if err != nil {
		if isConflict(err) {
			s.recordContentFailure(ctx, actorID, ModuleActivities.UpdateCode(), label)
		}
		return ActividadAdmin{}, fmt.Errorf("editar la actividad: %w", err)
	}
	return s.actividadAdmin(ctx, saved)
}

// DeleteActividad elimina físicamente una actividad y registra `activity.delete`
// con su `targetLabel` conservado (R4-10). Un id inexistente → 404.
func (s *service) DeleteActividad(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	current, found, err := s.repository.GetActivityByID(ctx, id)
	if err != nil {
		return fmt.Errorf("leer la actividad: %w", err)
	}
	if !found {
		return apperr.NotFound("La actividad no existe")
	}
	actions := []audit.Action{contentAction(actorID, ModuleActivities.DeleteCode(), contentLabel(ModuleActivities, current.TitleEs))}
	deleted, err := s.repository.DeleteActivity(ctx, id, actions...)
	if err != nil {
		return fmt.Errorf("eliminar la actividad: %w", err)
	}
	if !deleted {
		return apperr.NotFound("La actividad no existe")
	}
	return nil
}

// GetActividad devuelve la ficha completa de una actividad (FR-011).
func (s *service) GetActividad(ctx context.Context, id uuid.UUID) (ActividadAdmin, error) {
	activity, found, err := s.repository.GetActivityByID(ctx, id)
	if err != nil {
		return ActividadAdmin{}, fmt.Errorf("leer la actividad: %w", err)
	}
	if !found {
		return ActividadAdmin{}, apperr.NotFound("La actividad no existe")
	}
	return s.actividadAdmin(ctx, activity)
}

// ListActividades devuelve el listado paginado del panel (FR-011): todas las
// actividades (borradoras y pasadas incluidas), ordenadas por
// `created_at DESC, id DESC`, con filtro opcional por estado.
func (s *service) ListActividades(ctx context.Context, state *PublicationState, params paginate.Params) (ActividadAdminList, error) {
	total, err := s.repository.CountActivities(ctx, state)
	if err != nil {
		return ActividadAdminList{}, fmt.Errorf("contar las actividades: %w", err)
	}
	activities, err := s.repository.ListActivities(ctx, state, params)
	if err != nil {
		return ActividadAdminList{}, fmt.Errorf("listar las actividades: %w", err)
	}

	// El corte de la regla R4-2 se calcula UNA vez para toda la página (la
	// actividad no tiene consultas por fila: su fecha vive en su propia fila).
	today, nowTime, err := s.cutoff()
	if err != nil {
		return ActividadAdminList{}, err
	}

	items := make([]ActividadAdmin, 0, len(activities))
	for _, activity := range activities {
		items = append(items, actividadAdminFrom(activity, today, nowTime))
	}
	return ActividadAdminList{Items: items, Total: total, Limit: params.Limit, Offset: params.Offset}, nil
}

// UploadActividadImage guarda una imagen subida para el módulo de actividades
// (FR-020) y la audita con `activity.image.upload` (fail-closed, R4-10).
func (s *service) UploadActividadImage(ctx context.Context, actorID uuid.UUID, data []byte) (ImageUploadResult, error) {
	return s.uploadImage(ctx, actorID, ModuleActivities, data)
}

// actividadAdmin arma la ficha del panel de una actividad con su marca de
// pasado (sin tocar su estado, FR-006/Q4).
func (s *service) actividadAdmin(ctx context.Context, activity Activity) (ActividadAdmin, error) {
	today, nowTime, err := s.cutoff()
	if err != nil {
		return ActividadAdmin{}, err
	}
	return actividadAdminFrom(activity, today, nowTime), nil
}

// actividadAdminFrom traduce la actividad al DTO del panel.
func actividadAdminFrom(activity Activity, today, nowTime string) ActividadAdmin {
	admin := ActividadAdmin{
		ID:               activity.ID.String(),
		TitleEs:          activity.TitleEs,
		TitleEn:          activity.TitleEn,
		DetailEs:         activity.DetailEs,
		DetailEn:         activity.DetailEn,
		PlaceEs:          activity.PlaceEs,
		PlaceEn:          activity.PlaceEn,
		MapURL:           activity.MapURL,
		OrganizerEs:      activity.OrganizerEs,
		OrganizerEn:      activity.OrganizerEn,
		ImageFile:        activity.ImageFile,
		ImageAltEs:       activity.ImageAltEs,
		ImageAltEn:       activity.ImageAltEn,
		Date:             activity.Date,
		StartTime:        activity.StartTime,
		EndTime:          activity.EndTime,
		PublicationState: string(activity.PublicationState),
		IsPast:           !isUpcoming(activity.Date, today, nowTime, activity.EndTime),
		CreatedAt:        activity.CreatedAt,
		UpdatedAt:        activity.UpdatedAt,
	}
	if activity.CategoryID != nil {
		id := activity.CategoryID.String()
		admin.CategoryID = &id
	}
	return admin
}

// normalizeActivityDate valida la fecha única de la actividad y su horario
// opcional (FR-015/Q2).
func normalizeActivityDate(dateRaw, startRaw, endRaw string) (string, *string, *string, error) {
	date, err := parseDate(dateRaw)
	if err != nil {
		return "", nil, nil, err
	}
	start, end, err := normalizeRange(startRaw, endRaw)
	if err != nil {
		return "", nil, nil, err
	}
	return date, start, end, nil
}

// mergeActividad combina la actividad actual con los campos presentes del PATCH
// y valida el resultado (FR-015).
func (s *service) mergeActividad(ctx context.Context, current Activity, patch ActividadPatch) (Activity, bool, bool, error) {
	updated := current
	dataChanged := false
	invalid := func(field, message string) error {
		return apperr.Invalid("Revisa los datos del formulario", apperr.WithDetails(map[string]any{field: message}))
	}

	if patch.TitleEs != nil {
		title := normalizeText(*patch.TitleEs)
		if title == "" {
			return Activity{}, false, false, invalid("titleEs", "Este campo es obligatorio.")
		}
		if len([]rune(title)) > MaxTitleLength {
			return Activity{}, false, false, invalid("titleEs", fmt.Sprintf("No puede tener más de %d caracteres.", MaxTitleLength))
		}
		updated.TitleEs = title
		dataChanged = true
	}
	if patch.TitleEn.Set() {
		updated.TitleEn = normalizeOptional(patch.TitleEn.Value())
		dataChanged = true
	}
	if patch.DetailEs != nil {
		detail := normalizeText(*patch.DetailEs)
		if detail == "" {
			return Activity{}, false, false, invalid("detailEs", "Este campo es obligatorio.")
		}
		if len([]rune(detail)) > MaxDetailLength {
			return Activity{}, false, false, invalid("detailEs", fmt.Sprintf("No puede tener más de %d caracteres.", MaxDetailLength))
		}
		updated.DetailEs = detail
		dataChanged = true
	}
	if patch.DetailEn.Set() {
		updated.DetailEn = normalizeOptional(patch.DetailEn.Value())
		dataChanged = true
	}
	if patch.PlaceEs != nil {
		place := normalizeText(*patch.PlaceEs)
		if place == "" {
			return Activity{}, false, false, invalid("placeEs", "Este campo es obligatorio.")
		}
		if len([]rune(place)) > MaxPlaceLength {
			return Activity{}, false, false, invalid("placeEs", fmt.Sprintf("No puede tener más de %d caracteres.", MaxPlaceLength))
		}
		updated.PlaceEs = place
		dataChanged = true
	}
	if patch.PlaceEn.Set() {
		updated.PlaceEn = normalizeOptional(patch.PlaceEn.Value())
		dataChanged = true
	}
	if patch.MapURL.Set() {
		mapURL := normalizeOptional(patch.MapURL.Value())
		if mapURL != nil {
			if err := validate.Struct(urlRule{URL: *mapURL}); err != nil {
				return Activity{}, false, false, apperr.Invalid(
					"El enlace del mapa no es válido",
					apperr.WithDetails(map[string]any{"mapUrl": "usa un enlace https"}),
				)
			}
		}
		updated.MapURL = mapURL
		dataChanged = true
	}
	if patch.OrganizerEs.Set() {
		updated.OrganizerEs = normalizeOptional(patch.OrganizerEs.Value())
		dataChanged = true
	}
	if patch.OrganizerEn.Set() {
		updated.OrganizerEn = normalizeOptional(patch.OrganizerEn.Value())
		dataChanged = true
	}
	if patch.CategoryID.Set() {
		categoryID, err := s.resolveCategoryID(ctx, patch.CategoryID.Value())
		if err != nil {
			return Activity{}, false, false, err
		}
		updated.CategoryID = categoryID
		dataChanged = true
	}
	if patch.ImageFile.Set() {
		updated.ImageFile = normalizeOptional(patch.ImageFile.Value())
		dataChanged = true
	}
	if patch.ImageAltEs.Set() {
		updated.ImageAltEs = normalizeOptional(patch.ImageAltEs.Value())
		dataChanged = true
	}
	if patch.ImageAltEn.Set() {
		updated.ImageAltEn = normalizeOptional(patch.ImageAltEn.Value())
		dataChanged = true
	}

	// Fecha y horario: se revalidan juntos con los valores resultantes (una
	// fecha nueva sin horario es válida; un fin sin inicio no lo es).
	dateRaw := updated.Date
	if patch.Date != nil {
		dateRaw = *patch.Date
	}
	startRaw, endRaw := "", ""
	if patch.StartTime.Set() {
		startRaw = patch.StartTime.Value()
	} else if updated.StartTime != nil {
		startRaw = *updated.StartTime
	}
	if patch.EndTime.Set() {
		endRaw = patch.EndTime.Value()
	} else if updated.EndTime != nil {
		endRaw = *updated.EndTime
	}
	if patch.Date != nil || patch.StartTime.Set() || patch.EndTime.Set() {
		date, start, end, err := normalizeActivityDate(dateRaw, startRaw, endRaw)
		if err != nil {
			return Activity{}, false, false, err
		}
		updated.Date, updated.StartTime, updated.EndTime = date, start, end
		dataChanged = true
	}

	if err := validateImage(derefString(updated.ImageFile), derefString(updated.ImageAltEs)); err != nil {
		return Activity{}, false, false, err
	}

	stateChanged := false
	if patch.PublicationState != nil {
		state := PublicationState(*patch.PublicationState)
		if !state.Valid() {
			return Activity{}, false, false, invalid("publicationState", "solo se admite draft o published")
		}
		updated.PublicationState = state
		stateChanged = state != current.PublicationState
	}
	return updated, dataChanged, stateChanged, nil
}

// --- Catálogo de categorías (Q6/Q7, T421) ---

// CategoriaAdmin es una categoría del catálogo en el panel (Q6): nombre en
// español obligatorio, en inglés opcional y su orden de presentación. El
// catálogo NO tiene estado de publicación (son etiquetas, no contenido).
type CategoriaAdmin struct {
	ID        string    `json:"id"`
	NameEs    string    `json:"nameEs"`
	NameEn    *string   `json:"nameEn"`
	SortOrder int       `json:"sortOrder"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CategoriaAdminList es el catálogo completo (acotado en el servidor a
// `MaxCategories`, sin paginación: es un catálogo pequeño, §8.1.3).
type CategoriaAdminList struct {
	Items []CategoriaAdmin `json:"items"`
}

// CategoriaPatch es el cuerpo de PATCH /api/v1/admin/categorias/{id}: solo los
// campos presentes se tocan (`null` explícito vacía el nombre en inglés).
type CategoriaPatch struct {
	NameEs    *string          `json:"nameEs,omitempty"`
	NameEn    Optional[string] `json:"nameEn"`
	SortOrder *int             `json:"sortOrder,omitempty"`
}

// hasChanges indica si el PATCH trae al menos un campo.
func (p CategoriaPatch) hasChanges() bool {
	return p.NameEs != nil || p.NameEn.Set() || p.SortOrder != nil
}

// ListCategorias devuelve el catálogo completo en su orden de presentación
// (`sort_order, id`): lo usan los selectores del panel (Q6).
func (s *service) ListCategorias(ctx context.Context) (CategoriaAdminList, error) {
	categories, err := s.repository.ListCategories(ctx)
	if err != nil {
		return CategoriaAdminList{}, fmt.Errorf("listar las categorías: %w", err)
	}
	items := make([]CategoriaAdmin, 0, len(categories))
	for _, category := range categories {
		items = append(items, categoriaAdminFrom(category))
	}
	return CategoriaAdminList{Items: items}, nil
}

// CreateCategoria da de alta una categoría del catálogo (Q6): nombre en español
// obligatorio y en inglés opcional, con `sortOrder` opcional; el catálogo está
// acotado a `MaxCategories`. Nombre duplicado → 409 (`UNIQUE (name_es)`).
// Registra `category.create` (R4-10) en la misma transacción.
func (s *service) CreateCategoria(ctx context.Context, actorID uuid.UUID, in CategoriaInput) (CategoriaAdmin, error) {
	fail := func(err error) (CategoriaAdmin, error) {
		s.recordContentFailure(ctx, actorID, ModuleCategories.CreateCode(), contentLabel(ModuleCategories, ""))
		return CategoriaAdmin{}, err
	}

	if err := validate.Struct(in); err != nil {
		return fail(err)
	}

	// El tope del catálogo se comprueba con un conteo (no cargando el catálogo
	// entero): es un tope blando, dos altas concurrentes podrían superarlo en una
	// fila (Q6).
	total, err := s.repository.CountCategories(ctx)
	if err != nil {
		return CategoriaAdmin{}, fmt.Errorf("contar las categorías: %w", err)
	}
	if total >= MaxCategories {
		return fail(apperr.Invalid(
			fmt.Sprintf("No se pueden guardar más de %d categorías", MaxCategories),
			apperr.WithDetails(map[string]any{"categorias": "límite alcanzado"}),
		))
	}

	category := Category{
		NameEs:    normalizeText(in.NameEs),
		NameEn:    normalizeOptional(in.NameEn),
		SortOrder: in.SortOrder,
	}
	label := contentLabel(ModuleCategories, category.NameEs)
	actions := []audit.Action{contentAction(actorID, ModuleCategories.CreateCode(), label)}
	saved, err := s.repository.InsertCategory(ctx, category, actions...)
	if err != nil {
		if isConflict(err) {
			s.recordContentFailure(ctx, actorID, ModuleCategories.CreateCode(), label)
		}
		return CategoriaAdmin{}, fmt.Errorf("crear la categoría: %w", err)
	}
	return categoriaAdminFrom(saved), nil
}

// UpdateCategoria edita una categoría (Q6). Renombrar está permitido: el cambio
// se ve de inmediato en los elementos publicados que la usan (es un dato del
// elemento, como cualquier otro). Nombre duplicado → 409. Registra
// `category.update` (R4-10) en la misma transacción.
func (s *service) UpdateCategoria(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch CategoriaPatch) (CategoriaAdmin, error) {
	fail := func(err error) (CategoriaAdmin, error) {
		s.recordContentFailure(ctx, actorID, ModuleCategories.UpdateCode(), contentLabel(ModuleCategories, id.String()))
		return CategoriaAdmin{}, err
	}

	if !patch.hasChanges() {
		return fail(apperr.Invalid("No hay cambios que guardar"))
	}

	current, found, err := s.repository.GetCategoryByID(ctx, id)
	if err != nil {
		return CategoriaAdmin{}, fmt.Errorf("leer la categoría: %w", err)
	}
	if !found {
		return fail(apperr.NotFound("La categoría no existe"))
	}

	updated, dataChanged, err := mergeCategoria(current, patch)
	if err != nil {
		return fail(err)
	}
	if !dataChanged {
		return categoriaAdminFrom(current), nil
	}

	label := contentLabel(ModuleCategories, updated.NameEs)
	actions := []audit.Action{contentAction(actorID, ModuleCategories.UpdateCode(), label)}
	saved, err := s.repository.UpdateCategory(ctx, updated, actions...)
	if err != nil {
		if isConflict(err) {
			s.recordContentFailure(ctx, actorID, ModuleCategories.UpdateCode(), label)
		}
		return CategoriaAdmin{}, fmt.Errorf("editar la categoría: %w", err)
	}
	return categoriaAdminFrom(saved), nil
}

// DeleteCategoria borra una categoría del catálogo y registra `category.delete`
// (R4-10) en la misma transacción. Si algún evento o actividad la usa responde
// `409 conflict` con el recuento ("está en uso por N elementos", P4-6, decisión
// firme de la puerta 2): la categoría en uso NO se desliga automáticamente, hay
// que quitarla primero de esos elementos. Un id inexistente → 404.
func (s *service) DeleteCategoria(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	current, found, err := s.repository.GetCategoryByID(ctx, id)
	if err != nil {
		return fmt.Errorf("leer la categoría: %w", err)
	}
	if !found {
		return apperr.NotFound("La categoría no existe")
	}

	usage, err := s.repository.CountCategoryUsage(ctx, id)
	if err != nil {
		return fmt.Errorf("contar el uso de la categoría: %w", err)
	}
	if usage > 0 {
		s.recordContentFailure(ctx, actorID, ModuleCategories.DeleteCode(), contentLabel(ModuleCategories, current.NameEs))
		return apperr.Conflict(
			categoryUsageMessage(usage),
			apperr.WithDetails(map[string]any{"categoryId": "quítala primero de esos elementos"}),
		)
	}

	actions := []audit.Action{contentAction(actorID, ModuleCategories.DeleteCode(), contentLabel(ModuleCategories, current.NameEs))}
	deleted, err := s.repository.DeleteCategory(ctx, id, actions...)
	if err != nil {
		// Carrera: alguien la usó entre el recuento y el borrado (la FK la
		// protege igualmente).
		if isConflict(err) {
			s.recordContentFailure(ctx, actorID, ModuleCategories.DeleteCode(), contentLabel(ModuleCategories, current.NameEs))
		}
		return fmt.Errorf("eliminar la categoría: %w", err)
	}
	if !deleted {
		return apperr.NotFound("La categoría no existe")
	}
	return nil
}

// categoryUsageMessage arma el mensaje del 409 del borrado con el recuento
// (P4-6): "está en uso por N elementos".
func categoryUsageMessage(usage int64) string {
	if usage == 1 {
		return "La categoría está en uso por 1 elemento"
	}
	return fmt.Sprintf("La categoría está en uso por %d elementos", usage)
}

// mergeCategoria combina la categoría actual con los campos presentes del PATCH
// y valida el resultado (Q6).
func mergeCategoria(current Category, patch CategoriaPatch) (Category, bool, error) {
	updated := current
	dataChanged := false

	if patch.NameEs != nil {
		name := normalizeText(*patch.NameEs)
		if name == "" {
			return Category{}, false, apperr.Invalid(
				"Revisa los datos del formulario",
				apperr.WithDetails(map[string]any{"nameEs": "Este campo es obligatorio."}),
			)
		}
		if len([]rune(name)) > MaxCategoryNameLength {
			return Category{}, false, apperr.Invalid(
				"Revisa los datos del formulario",
				apperr.WithDetails(map[string]any{"nameEs": fmt.Sprintf("No puede tener más de %d caracteres.", MaxCategoryNameLength)}),
			)
		}
		updated.NameEs = name
		dataChanged = true
	}
	if patch.NameEn.Set() {
		updated.NameEn = normalizeOptional(patch.NameEn.Value())
		dataChanged = true
	}
	if patch.SortOrder != nil {
		updated.SortOrder = *patch.SortOrder
		dataChanged = true
	}
	return updated, dataChanged, nil
}

// categoriaAdminFrom traduce la categoría guardada a su DTO del panel.
func categoriaAdminFrom(category Category) CategoriaAdmin {
	return CategoriaAdmin{
		ID:        category.ID.String(),
		NameEs:    category.NameEs,
		NameEn:    category.NameEn,
		SortOrder: category.SortOrder,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}
