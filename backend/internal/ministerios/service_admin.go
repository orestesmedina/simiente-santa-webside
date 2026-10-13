package ministerios

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/storage"
	"simiente-santa/backend/internal/platform/validate"
)

// service_admin.go implementa la gestión del módulo desde el panel (T615):
// alta, edición parcial, borrado, detalle y listado paginado. No conoce HTTP ni
// SQL: valida (FR-014), normaliza (analyze I6), «publica al guardar» (FR-013),
// aplica el **`PATCH` que fusiona los campos presentes** sobre la fila actual y
// escribe la fila completa ya fusionada (A6-3/P6-20), y compone las filas de
// auditoría con la regla exacta de códigos (R6-12/A6-10), que el repository
// inserta en la MISMA transacción (FR-017).

// --- DTOs del panel (contrato `MinisterioAdmin`, `MinisterioAdminList`) ---

// MinisterioAdmin es la ficha completa del panel (FR-010/FR-019): ambos idiomas,
// la ficha del encargado **íntegra aunque no esté publicada** (dato guardado ≠
// dato publicado), el estado de publicación y el aviso `whatsappUrlShared`
// —`true` si **otro** ministerio usa el mismo enlace: un aviso no bloqueante que
// se recalcula en cada lectura (R6-11)—. No tiene `isPast`: en este módulo nada
// vence (sin fechas ni recurrencias, FR-024).
type MinisterioAdmin struct {
	ID                string    `json:"id"`
	NameEs            string    `json:"nameEs"`
	NameEn            *string   `json:"nameEn"`
	DescriptionEs     string    `json:"descriptionEs"`
	DescriptionEn     *string   `json:"descriptionEn"`
	LeaderNameEs      *string   `json:"leaderNameEs"`
	LeaderNameEn      *string   `json:"leaderNameEn"`
	LeaderPhone       *string   `json:"leaderPhone"`
	LeaderEmail       *string   `json:"leaderEmail"`
	LeaderVisible     bool      `json:"leaderVisible"`
	WhatsappURL       *string   `json:"whatsappUrl"`
	WhatsappURLShared bool      `json:"whatsappUrlShared"`
	ScheduleEs        *string   `json:"scheduleEs"`
	ScheduleEn        *string   `json:"scheduleEn"`
	PlaceEs           *string   `json:"placeEs"`
	PlaceEn           *string   `json:"placeEn"`
	ImageFile         *string   `json:"imageFile"`
	ImageAltEs        *string   `json:"imageAltEs"`
	ImageAltEn        *string   `json:"imageAltEn"`
	PublicationState  string    `json:"publicationState"`
	SortOrder         int       `json:"sortOrder"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// MinisterioAdminList es el listado paginado del panel (§8.1.2/§8.1.3).
type MinisterioAdminList struct {
	Items  []MinisterioAdmin `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

// ImageUploadResult es el resultado de una subida (contrato `ImageUploadResult`,
// R6-7): el nombre generado por el servidor y su URL pública relativa.
type ImageUploadResult struct {
	FileName  string `json:"fileName"`
	URL       string `json:"url"`
	MimeType  string `json:"mimeType"`
	SizeBytes int64  `json:"sizeBytes"`
}

// MinisterioPatch es el cuerpo de PATCH /api/v1/admin/ministerios/{id}: solo los
// campos **presentes** se tocan y un `null` explícito vacía el opcional
// (A6-3/P6-20). Los campos no anulables (`leaderVisible`, `sortOrder`,
// `publicationState`) son punteros: `null` no es un valor válido para ellos y se
// comporta como ausente.
type MinisterioPatch struct {
	NameEs           *string          `json:"nameEs,omitempty"`
	NameEn           Optional[string] `json:"nameEn"`
	DescriptionEs    *string          `json:"descriptionEs,omitempty"`
	DescriptionEn    Optional[string] `json:"descriptionEn"`
	LeaderNameEs     Optional[string] `json:"leaderNameEs"`
	LeaderNameEn     Optional[string] `json:"leaderNameEn"`
	LeaderPhone      Optional[string] `json:"leaderPhone"`
	LeaderEmail      Optional[string] `json:"leaderEmail"`
	LeaderVisible    *bool            `json:"leaderVisible,omitempty"`
	WhatsappURL      Optional[string] `json:"whatsappUrl"`
	ScheduleEs       Optional[string] `json:"scheduleEs"`
	ScheduleEn       Optional[string] `json:"scheduleEn"`
	PlaceEs          Optional[string] `json:"placeEs"`
	PlaceEn          Optional[string] `json:"placeEn"`
	ImageFile        Optional[string] `json:"imageFile"`
	ImageAltEs       Optional[string] `json:"imageAltEs"`
	ImageAltEn       Optional[string] `json:"imageAltEn"`
	PublicationState *string          `json:"publicationState,omitempty"`
	SortOrder        *int             `json:"sortOrder,omitempty"`
}

// hasChanges indica si el PATCH trae al menos un campo (minProperties: 1). Un
// `null` explícito cuenta como cambio: limpiar un campo opcional lo es.
func (p MinisterioPatch) hasChanges() bool {
	return p.NameEs != nil || p.NameEn.Set() || p.DescriptionEs != nil || p.DescriptionEn.Set() ||
		p.LeaderNameEs.Set() || p.LeaderNameEn.Set() || p.LeaderPhone.Set() || p.LeaderEmail.Set() ||
		p.LeaderVisible != nil || p.WhatsappURL.Set() || p.ScheduleEs.Set() || p.ScheduleEn.Set() ||
		p.PlaceEs.Set() || p.PlaceEn.Set() || p.ImageFile.Set() || p.ImageAltEs.Set() ||
		p.ImageAltEn.Set() || p.PublicationState != nil || p.SortOrder != nil
}

// --- Gestión del panel (FR-010) ---

// CreateMinisterio da de alta un ministerio (FR-010/FR-014): validación completa
// —el `400` lleva `details` por campo y **no guarda nada a medias**—, duplicado
// exacto → `409 conflict` (P6-12) y `leaderVisible` con default `false` («Oculto
// por defecto», P6-3: si la clave falta, la ficha del encargado nace NO
// publicada). El enlace compartido con otro ministerio se permite (aviso, nunca
// bloqueo, RG6-11).
//
// Auditoría (A6-10/R6-12): el alta registra **solo `ministry.create`** —una sola
// fila— también cuando nace publicada o con la ficha del encargado visible.
func (s *service) CreateMinisterio(ctx context.Context, actorID uuid.UUID, in MinisterioInput) (MinisterioAdmin, error) {
	label := contentLabel("")
	fail := func(err error) (MinisterioAdmin, error) {
		s.recordContentFailure(ctx, actorID, audit.ActionMinistryCreate, label)
		return MinisterioAdmin{}, err
	}

	if err := validate.Struct(in); err != nil {
		return fail(err)
	}
	ministry, err := ministryFromInput(in)
	if err != nil {
		return fail(err)
	}
	label = contentLabel(ministry.NameEs)

	duplicate, err := s.repository.ExistsMinistryDuplicate(
		ctx, ministry.NameEs, ministry.DescriptionEs, ministry.LeaderNameEs, ministry.WhatsappURL, nil,
	)
	if err != nil {
		return MinisterioAdmin{}, fmt.Errorf("comprobar el duplicado: %w", err)
	}
	if duplicate {
		return fail(apperr.Conflict(messageDuplicate))
	}

	actions := []audit.Action{contentAction(actorID, audit.ActionMinistryCreate, label)}
	saved, err := s.repository.InsertMinistry(ctx, ministry, actions...)
	if err != nil {
		// La base es la última línea de defensa del duplicado: su rechazo también
		// queda registrado (FR-017/R6-12).
		if isConflict(err) {
			s.recordContentFailure(ctx, actorID, audit.ActionMinistryCreate, label)
		}
		return MinisterioAdmin{}, fmt.Errorf("crear el ministerio: %w", err)
	}
	return s.ministerioAdmin(ctx, saved.ID)
}

// UpdateMinisterio edita parcialmente un ministerio (FR-010): fusiona los campos
// presentes sobre la fila actual y escribe la fila completa ya fusionada
// (A6-3/P6-20), revalidando el resultado con las mismas reglas del alta
// (FR-014). Un id inexistente → `404`; un duplicado exacto → `409`.
//
// Auditoría (R6-12/A6-10): `ministry.update` si cambian datos, `ministry.publish`
// o `ministry.unpublish` si cambia el estado y `ministry.leader.show` o
// `ministry.leader.hide` si cambia la visibilidad de la ficha —una fila por
// operación, todas en la misma transacción—. Apagar el interruptor **conserva**
// el nombre y el contacto guardados (US2 esc. 10–11).
func (s *service) UpdateMinisterio(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch MinisterioPatch) (MinisterioAdmin, error) {
	label := contentLabel(id.String())
	fail := func(err error) (MinisterioAdmin, error) {
		s.recordContentFailure(ctx, actorID, audit.ActionMinistryUpdate, label)
		return MinisterioAdmin{}, err
	}

	if !patch.hasChanges() {
		return fail(apperr.Invalid("No hay cambios que guardar"))
	}

	current, found, err := s.repository.GetMinistryByID(ctx, id)
	if err != nil {
		return MinisterioAdmin{}, fmt.Errorf("leer el ministerio: %w", err)
	}
	if !found {
		return fail(apperr.NotFound(messageNotFound))
	}
	label = contentLabel(current.NameEs)

	updated, dataChanged, stateChanged, leaderChanged, err := mergeMinisterio(current.Ministry, patch)
	if err != nil {
		return fail(err)
	}
	if !dataChanged && !stateChanged && !leaderChanged {
		return ministerioAdminFrom(current), nil
	}
	label = contentLabel(updated.NameEs)

	if dataChanged {
		duplicate, err := s.repository.ExistsMinistryDuplicate(
			ctx, updated.NameEs, updated.DescriptionEs, updated.LeaderNameEs, updated.WhatsappURL, &id,
		)
		if err != nil {
			return MinisterioAdmin{}, fmt.Errorf("comprobar el duplicado: %w", err)
		}
		if duplicate {
			return fail(apperr.Conflict(messageDuplicate))
		}
	}

	actions := make([]audit.Action, 0, 3)
	if dataChanged {
		actions = append(actions, contentAction(actorID, audit.ActionMinistryUpdate, label))
	}
	if stateChanged {
		actions = append(actions, contentAction(actorID, stateCode(updated.PublicationState), label))
	}
	if leaderChanged {
		actions = append(actions, contentAction(actorID, leaderCode(updated.LeaderVisible), label))
	}

	saved, err := s.repository.UpdateMinistry(ctx, updated, actions...)
	if err != nil {
		if isConflict(err) {
			s.recordContentFailure(ctx, actorID, audit.ActionMinistryUpdate, label)
		}
		return MinisterioAdmin{}, fmt.Errorf("editar el ministerio: %w", err)
	}

	// A6-11: la imagen reemplazada o quitada se limpia del disco best-effort (un
	// fallo de limpieza no cambia la respuesta). Un archivo que deja de
	// referenciarse simplemente deja de servirse (`IsFilePublished`).
	s.deleteReplacedImage(ctx, current.ImageFile, saved.ImageFile)

	return s.ministerioAdmin(ctx, saved.ID)
}

// DeleteMinisterio elimina definitivamente un ministerio (FR-010) y registra
// `ministry.delete` con su etiqueta conservada (R6-12) en la misma transacción.
// Un id inexistente → `404`.
//
// **No borra el archivo de la imagen** (A6-11): la fila deja de referenciarlo y
// `IsFilePublished` deja de servirlo; si volviera a referenciarse, el archivo
// sigue ahí.
func (s *service) DeleteMinisterio(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	current, found, err := s.repository.GetMinistryByID(ctx, id)
	if err != nil {
		return fmt.Errorf("leer el ministerio: %w", err)
	}
	if !found {
		return apperr.NotFound(messageNotFound)
	}

	actions := []audit.Action{contentAction(actorID, audit.ActionMinistryDelete, contentLabel(current.NameEs))}
	deleted, err := s.repository.DeleteMinistry(ctx, id, actions...)
	if err != nil {
		return fmt.Errorf("eliminar el ministerio: %w", err)
	}
	if !deleted {
		return apperr.NotFound(messageNotFound)
	}
	return nil
}

// GetMinisterio devuelve la ficha completa del panel (FR-010), con la ficha del
// encargado aunque no esté publicada. Un id inexistente → `404`.
func (s *service) GetMinisterio(ctx context.Context, id uuid.UUID) (MinisterioAdmin, error) {
	item, found, err := s.repository.GetMinistryByID(ctx, id)
	if err != nil {
		return MinisterioAdmin{}, fmt.Errorf("leer el ministerio: %w", err)
	}
	if !found {
		return MinisterioAdmin{}, apperr.NotFound(messageNotFound)
	}
	return ministerioAdminFrom(item), nil
}

// ListMinisterios devuelve el listado paginado del panel (FR-010): todos los
// ministerios —borradores incluidos— en el orden del módulo (`sort_order`, `id`,
// el mismo que ve el visitante: P6-19/US2 esc. 9), con filtro opcional por
// estado de publicación y el aviso de enlace compartido por fila.
func (s *service) ListMinisterios(ctx context.Context, state *PublicationState, params paginate.Params) (MinisterioAdminList, error) {
	total, err := s.repository.CountMinistries(ctx, state)
	if err != nil {
		return MinisterioAdminList{}, fmt.Errorf("contar los ministerios: %w", err)
	}
	items, err := s.repository.ListMinistries(ctx, state, params)
	if err != nil {
		return MinisterioAdminList{}, fmt.Errorf("listar los ministerios: %w", err)
	}

	out := make([]MinisterioAdmin, 0, len(items))
	for _, item := range items {
		out = append(out, ministerioAdminFrom(item))
	}
	return MinisterioAdminList{Items: out, Total: total, Limit: params.Limit, Offset: params.Offset}, nil
}

// --- Imágenes (FR-020/R6-7) ---

// UploadMinisterioImage guarda una imagen subida desde el panel: el Store valida
// la firma binaria (JPEG/PNG/WebP, sin SVG ni GIF) y el tamaño, y genera el
// nombre `img_<uuid>.<ext>`. La subida se audita con `ministry.image.upload` de
// forma **fail-closed**: si el registro falla, el archivo se elimina y la subida
// se aborta (nunca queda una subida sin traza). La subida no cambia contenido
// visible: la referencia se registra al guardar el ministerio (`imageFile`).
func (s *service) UploadMinisterioImage(ctx context.Context, actorID uuid.UUID, data []byte) (ImageUploadResult, error) {
	if s.store == nil {
		return ImageUploadResult{}, apperr.Internal(errors.New("ministerios: almacén de imágenes no configurado"))
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

	action := contentAction(actorID, audit.ActionMinistryImageUpload, imageLabel(file.Name))
	if err := s.repository.RecordAction(ctx, action); err != nil {
		// Fail-closed: sin registro no hay subida (R6-7).
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

// --- Mezcla del PATCH (A6-3/P6-20) ---

// mergeMinisterio fusiona los campos presentes del PATCH sobre la fila actual y
// devuelve el ministerio resultante, si cambiaron los datos, si cambió el estado
// de publicación y si cambió la visibilidad de la ficha del encargado. El
// resultado se revalida con las mismas reglas del alta (FR-014) ANTES de
// escribir: la transacción guarda la fila completa ya fusionada.
func mergeMinisterio(current Ministry, patch MinisterioPatch) (Ministry, bool, bool, bool, error) {
	updated := current
	dataChanged := false

	if patch.NameEs != nil {
		updated.NameEs = normalizeText(*patch.NameEs)
		dataChanged = true
	}
	if patch.NameEn.Set() {
		updated.NameEn = normalizeOptional(patch.NameEn.Value())
		dataChanged = true
	}
	if patch.DescriptionEs != nil {
		updated.DescriptionEs = normalizeText(*patch.DescriptionEs)
		dataChanged = true
	}
	if patch.DescriptionEn.Set() {
		updated.DescriptionEn = normalizeOptional(patch.DescriptionEn.Value())
		dataChanged = true
	}
	if patch.LeaderNameEs.Set() {
		updated.LeaderNameEs = normalizeOptional(patch.LeaderNameEs.Value())
		dataChanged = true
	}
	if patch.LeaderNameEn.Set() {
		updated.LeaderNameEn = normalizeOptional(patch.LeaderNameEn.Value())
		dataChanged = true
	}
	if patch.LeaderPhone.Set() {
		updated.LeaderPhone = normalizeOptional(patch.LeaderPhone.Value())
		dataChanged = true
	}
	if patch.LeaderEmail.Set() {
		updated.LeaderEmail = normalizeOptional(patch.LeaderEmail.Value())
		dataChanged = true
	}
	if patch.WhatsappURL.Set() {
		updated.WhatsappURL = normalizeOptional(patch.WhatsappURL.Value())
		dataChanged = true
	}
	if patch.ScheduleEs.Set() {
		updated.ScheduleEs = normalizeOptional(patch.ScheduleEs.Value())
		dataChanged = true
	}
	if patch.ScheduleEn.Set() {
		updated.ScheduleEn = normalizeOptional(patch.ScheduleEn.Value())
		dataChanged = true
	}
	if patch.PlaceEs.Set() {
		updated.PlaceEs = normalizeOptional(patch.PlaceEs.Value())
		dataChanged = true
	}
	if patch.PlaceEn.Set() {
		updated.PlaceEn = normalizeOptional(patch.PlaceEn.Value())
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
	if patch.SortOrder != nil {
		updated.SortOrder = *patch.SortOrder
		dataChanged = true
	}

	stateChanged := false
	if patch.PublicationState != nil {
		state := PublicationState(*patch.PublicationState)
		if !state.Valid() {
			return Ministry{}, false, false, false, apperr.Invalid(
				validate.MessageInvalid,
				apperr.WithDetails(map[string]any{"publicationState": "solo se admite draft o published"}),
			)
		}
		updated.PublicationState = state
		stateChanged = state != current.PublicationState
	}

	// La visibilidad de la ficha no toca los datos guardados: solo decide si se
	// publican (FR-025/FR-026).
	leaderChanged := false
	if patch.LeaderVisible != nil {
		updated.LeaderVisible = *patch.LeaderVisible
		leaderChanged = *patch.LeaderVisible != current.LeaderVisible
	}

	if err := validateMinistry(updated); err != nil {
		return Ministry{}, false, false, false, err
	}
	return updated, dataChanged, stateChanged, leaderChanged, nil
}

// --- Helpers ---

// messageNotFound es el mensaje del 404 del módulo.
const messageNotFound = "El ministerio no existe"

// isConflict indica si el error es un conflicto de dominio (409): el rechazo de
// un duplicado detectado por la BASE también queda en la auditoría
// (FR-017/R6-12).
func isConflict(err error) bool {
	var domainErr *apperr.Error
	return errors.As(err, &domainErr) && domainErr.Kind == apperr.KindConflict
}

// ministerioAdmin vuelve a leer el ministerio guardado para responder el DTO del
// panel con el aviso `whatsappUrlShared` recalculado en esta lectura (R6-11).
func (s *service) ministerioAdmin(ctx context.Context, id uuid.UUID) (MinisterioAdmin, error) {
	item, found, err := s.repository.GetMinistryByID(ctx, id)
	if err != nil {
		return MinisterioAdmin{}, fmt.Errorf("leer el ministerio: %w", err)
	}
	if !found {
		return MinisterioAdmin{}, apperr.NotFound(messageNotFound)
	}
	return ministerioAdminFrom(item), nil
}

// deleteReplacedImage borra best-effort la imagen anterior cuando la referencia
// cambió (A6-11): reemplazarla o quitarla limpia el archivo; conservarla, no.
func (s *service) deleteReplacedImage(ctx context.Context, previous, current *string) {
	if previous == nil {
		return
	}
	if current != nil && *previous == *current {
		return
	}
	s.deleteFile(ctx, *previous)
}

// ministerioAdminFrom traduce la fila del panel a su DTO (ambos idiomas, la
// ficha del encargado íntegra y el aviso de enlace compartido).
func ministerioAdminFrom(item MinistryListItem) MinisterioAdmin {
	ministry := item.Ministry
	return MinisterioAdmin{
		ID:                ministry.ID.String(),
		NameEs:            ministry.NameEs,
		NameEn:            ministry.NameEn,
		DescriptionEs:     ministry.DescriptionEs,
		DescriptionEn:     ministry.DescriptionEn,
		LeaderNameEs:      ministry.LeaderNameEs,
		LeaderNameEn:      ministry.LeaderNameEn,
		LeaderPhone:       ministry.LeaderPhone,
		LeaderEmail:       ministry.LeaderEmail,
		LeaderVisible:     ministry.LeaderVisible,
		WhatsappURL:       ministry.WhatsappURL,
		WhatsappURLShared: item.WhatsappURLShared,
		ScheduleEs:        ministry.ScheduleEs,
		ScheduleEn:        ministry.ScheduleEn,
		PlaceEs:           ministry.PlaceEs,
		PlaceEn:           ministry.PlaceEn,
		ImageFile:         ministry.ImageFile,
		ImageAltEs:        ministry.ImageAltEs,
		ImageAltEn:        ministry.ImageAltEn,
		PublicationState:  string(ministry.PublicationState),
		SortOrder:         ministry.SortOrder,
		CreatedAt:         ministry.CreatedAt,
		UpdatedAt:         ministry.UpdatedAt,
	}
}
