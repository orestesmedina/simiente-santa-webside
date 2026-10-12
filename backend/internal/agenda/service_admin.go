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

	items := make([]EventoAdmin, 0, len(events))
	for _, event := range events {
		admin, err := s.eventoAdmin(ctx, event)
		if err != nil {
			return EventoAdminList{}, err
		}
		items = append(items, admin)
	}
	return EventoAdminList{Items: items, Total: total, Limit: params.Limit, Offset: params.Offset}, nil
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
