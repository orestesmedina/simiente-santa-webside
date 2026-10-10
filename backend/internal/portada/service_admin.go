package portada

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/storage"
	"simiente-santa/backend/internal/platform/validate"
)

// service_admin.go implementa la gestión del módulo desde el panel (T319/T320):
// identidad, «quiénes somos» y contacto (singletons) y horario, WhatsApp y redes
// (colecciones). No conoce HTTP ni SQL: valida (FR-015), normaliza (analyze I6),
// aplica los límites de colección (R3-14), publica al guardar (FR-014) y compone
// las filas de auditoría con la regla exacta de códigos (analyze M5), que el
// repository inserta en la MISMA transacción (FR-017).

// --- DTOs de salida del panel (contrato `PortadaAdmin`) ---

// IdentityAdmin es la identidad con sus dos idiomas y su estado (FR-002/FR-008).
type IdentityAdmin struct {
	NameEs           string    `json:"nameEs"`
	NameEn           *string   `json:"nameEn,omitempty"`
	TaglineEs        *string   `json:"taglineEs,omitempty"`
	TaglineEn        *string   `json:"taglineEn,omitempty"`
	MissionEs        *string   `json:"missionEs,omitempty"`
	MissionEn        *string   `json:"missionEn,omitempty"`
	VisionEs         *string   `json:"visionEs,omitempty"`
	VisionEn         *string   `json:"visionEn,omitempty"`
	LogoFile         *string   `json:"logoFile,omitempty"`
	LogoAltEs        *string   `json:"logoAltEs,omitempty"`
	LogoAltEn        *string   `json:"logoAltEn,omitempty"`
	CoverImageFile   *string   `json:"coverImageFile,omitempty"`
	CoverImageAltEs  *string   `json:"coverImageAltEs,omitempty"`
	CoverImageAltEn  *string   `json:"coverImageAltEn,omitempty"`
	PublicationState string    `json:"publicationState"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// AboutAdmin es «quiénes somos» en sus dos idiomas (FR-003).
type AboutAdmin struct {
	TextEs           string    `json:"textEs"`
	TextEn           *string   `json:"textEn,omitempty"`
	PublicationState string    `json:"publicationState"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// ContactAdmin son los datos de contacto del panel (FR-007).
type ContactAdmin struct {
	AddressEs        string    `json:"addressEs"`
	AddressEn        *string   `json:"addressEn,omitempty"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone"`
	PublicationState string    `json:"publicationState"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// ScheduleItemAdmin es un servicio del horario en el panel (FR-004).
type ScheduleItemAdmin struct {
	ID               string    `json:"id"`
	DayOfWeek        int       `json:"dayOfWeek"`
	StartTime        string    `json:"startTime"`
	EndTime          *string   `json:"endTime,omitempty"`
	NameEs           string    `json:"nameEs"`
	NameEn           *string   `json:"nameEn,omitempty"`
	DescriptionEs    *string   `json:"descriptionEs,omitempty"`
	DescriptionEn    *string   `json:"descriptionEn,omitempty"`
	PlaceEs          string    `json:"placeEs"`
	PlaceEn          *string   `json:"placeEn,omitempty"`
	PublicationState string    `json:"publicationState"`
	SortOrder        int       `json:"sortOrder"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// WhatsappChannelAdmin es un canal de WhatsApp en el panel (FR-005).
type WhatsappChannelAdmin struct {
	ID               string    `json:"id"`
	NameEs           string    `json:"nameEs"`
	NameEn           *string   `json:"nameEn,omitempty"`
	Kind             string    `json:"kind"`
	Destination      string    `json:"destination"`
	PublicationState string    `json:"publicationState"`
	SortOrder        int       `json:"sortOrder"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// SocialLinkAdmin es un enlace de red en el panel (FR-006).
type SocialLinkAdmin struct {
	ID               string    `json:"id"`
	Network          string    `json:"network"`
	URL              string    `json:"url"`
	PublicationState string    `json:"publicationState"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// ImageUploadResult es el resultado de una subida (contrato `ImageUploadResult`,
// research R3-8): el nombre generado por el servidor y su URL pública relativa.
type ImageUploadResult struct {
	FileName  string `json:"fileName"`
	URL       string `json:"url"`
	MimeType  string `json:"mimeType"`
	SizeBytes int64  `json:"sizeBytes"`
}

// PortadaAdmin es el estado completo del módulo para precargar el panel
// (FR-011, analyze C1): los singletons son null hasta el primer guardado y cada
// colección va en su sobre `{items}` (nunca null, §8.1.2). A diferencia de la
// vista pública, aquí se devuelven los pares Es/En crudos y los borradores.
type PortadaAdmin struct {
	Identity *IdentityAdmin        `json:"identity"`
	About    *AboutAdmin           `json:"about"`
	Contact  *ContactAdmin         `json:"contact"`
	Schedule ScheduleItemsAdmin    `json:"schedule"`
	Whatsapp WhatsappChannelsAdmin `json:"whatsapp"`
	Socials  SocialLinksAdmin      `json:"socials"`
}

// ScheduleItemsAdmin es el sobre de la colección del horario.
type ScheduleItemsAdmin struct {
	Items []ScheduleItemAdmin `json:"items"`
}

// WhatsappChannelsAdmin es el sobre de la colección de canales de WhatsApp.
type WhatsappChannelsAdmin struct {
	Items []WhatsappChannelAdmin `json:"items"`
}

// SocialLinksAdmin es el sobre de la colección de redes sociales.
type SocialLinksAdmin struct {
	Items []SocialLinkAdmin `json:"items"`
}

// --- DTOs de entrada de las ediciones parciales (PATCH) ---
//
// Un campo ausente no se toca. Los campos anulables del contrato (`endTime`,
// `nameEn`, `descriptionEs`, `descriptionEn`, `placeEn`) usan Optional[T] para
// distinguir además el `null` explícito, que vacía el campo (analyze I6/B1). Los
// campos no anulables son punteros: `null` no es un valor válido para ellos, así
// que se comporta como ausente. La validación manual del service cubre lo que el
// validador de structs no ve a través de punteros.

// ScheduleItemPatch es el cuerpo de PATCH /api/v1/admin/portada/horario/{id}.
type ScheduleItemPatch struct {
	DayOfWeek        *int             `json:"dayOfWeek,omitempty"`
	StartTime        *string          `json:"startTime,omitempty"`
	EndTime          Optional[string] `json:"endTime"`
	NameEs           *string          `json:"nameEs,omitempty"`
	NameEn           Optional[string] `json:"nameEn"`
	DescriptionEs    Optional[string] `json:"descriptionEs"`
	DescriptionEn    Optional[string] `json:"descriptionEn"`
	PlaceEs          *string          `json:"placeEs,omitempty"`
	PlaceEn          Optional[string] `json:"placeEn"`
	PublicationState *string          `json:"publicationState,omitempty"`
	SortOrder        *int             `json:"sortOrder,omitempty"`
}

// WhatsappChannelPatch es el cuerpo de PATCH /api/v1/admin/portada/whatsapp/{id}.
type WhatsappChannelPatch struct {
	NameEs           *string          `json:"nameEs,omitempty"`
	NameEn           Optional[string] `json:"nameEn"`
	Kind             *string          `json:"kind,omitempty"`
	Destination      *string          `json:"destination,omitempty"`
	PublicationState *string          `json:"publicationState,omitempty"`
	SortOrder        *int             `json:"sortOrder,omitempty"`
}

// SocialLinkPatch es el cuerpo de PATCH /api/v1/admin/portada/redes/{id}.
type SocialLinkPatch struct {
	Network          *string `json:"network,omitempty"`
	URL              *string `json:"url,omitempty"`
	PublicationState *string `json:"publicationState,omitempty"`
}

// --- Identidad, «quiénes somos» y contacto (T319) ---

// SaveIdentity reemplaza el singleton de identidad (FR-002/FR-011): valida
// FR-015 (nombre en español, correo/alt de imagen cuando hay imagen), normaliza
// los `*En` vacíos a NULL (analyze I6), publica al guardar (FR-014) y registra
// `home.identity.update` —más `home.publish`/`home.unpublish` si cambia el
// estado (analyze M5)— en la misma transacción (FR-017). Al reemplazar una
// imagen, la anterior se borra best-effort (R3-8).
func (s *service) SaveIdentity(ctx context.Context, actorID uuid.UUID, in IdentityInput) (IdentityAdmin, error) {
	if err := validate.Struct(in); err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeIdentityUpdate, contentLabel(sectionIdentity, ""))
		return IdentityAdmin{}, err
	}
	if err := identityImageRules(in); err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeIdentityUpdate, contentLabel(sectionIdentity, ""))
		return IdentityAdmin{}, err
	}

	identity := Identity{
		NameEs:           normalizeText(in.NameEs),
		NameEn:           normalizeOptional(in.NameEn),
		TaglineEs:        normalizeOptional(in.TaglineEs),
		TaglineEn:        normalizeOptional(in.TaglineEn),
		MissionEs:        normalizeOptional(in.MissionEs),
		MissionEn:        normalizeOptional(in.MissionEn),
		VisionEs:         normalizeOptional(in.VisionEs),
		VisionEn:         normalizeOptional(in.VisionEn),
		LogoFile:         normalizeOptional(in.LogoFile),
		LogoAltEs:        normalizeOptional(in.LogoAltEs),
		LogoAltEn:        normalizeOptional(in.LogoAltEn),
		CoverImageFile:   normalizeOptional(in.CoverImageFile),
		CoverImageAltEs:  normalizeOptional(in.CoverImageAltEs),
		CoverImageAltEn:  normalizeOptional(in.CoverImageAltEn),
		PublicationState: PublicationState(in.PublicationState),
	}

	previous, found, err := s.repository.GetHomeIdentity(ctx)
	if err != nil {
		return IdentityAdmin{}, fmt.Errorf("leer la identidad: %w", err)
	}

	label := contentLabel(sectionIdentity, identity.NameEs)
	actions := []audit.Action{contentAction(actorID, audit.ActionHomeIdentityUpdate, label)}
	if found && previous.PublicationState != identity.PublicationState {
		actions = append(actions, contentAction(actorID, stateActionCode(identity.PublicationState), label))
	}

	saved, err := s.repository.UpsertHomeIdentity(ctx, identity, actions...)
	if err != nil {
		return IdentityAdmin{}, fmt.Errorf("guardar la identidad: %w", err)
	}

	if found {
		s.deleteReplacedImage(ctx, previous.LogoFile, saved.LogoFile)
		s.deleteReplacedImage(ctx, previous.CoverImageFile, saved.CoverImageFile)
	}
	return identityAdminFrom(saved), nil
}

// SaveAbout reemplaza «quiénes somos» (FR-003): español obligatorio y ≤1.000
// caracteres, inglés opcional (analyze I6), publicación al guardar (FR-014) y
// auditoría transaccional `home.about.update` (+ publish/unpublish si cambia el
// estado).
func (s *service) SaveAbout(ctx context.Context, actorID uuid.UUID, in AboutInput) (AboutAdmin, error) {
	if err := validate.Struct(in); err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeAboutUpdate, contentLabel(sectionAbout, ""))
		return AboutAdmin{}, err
	}

	about := About{
		TextEs:           normalizeText(in.TextEs),
		TextEn:           normalizeOptional(in.TextEn),
		PublicationState: PublicationState(in.PublicationState),
	}

	previous, found, err := s.repository.GetHomeAbout(ctx)
	if err != nil {
		return AboutAdmin{}, fmt.Errorf("leer quiénes somos: %w", err)
	}

	label := contentLabel(sectionAbout, "")
	actions := []audit.Action{contentAction(actorID, audit.ActionHomeAboutUpdate, label)}
	if found && previous.PublicationState != about.PublicationState {
		actions = append(actions, contentAction(actorID, stateActionCode(about.PublicationState), label))
	}

	saved, err := s.repository.UpsertHomeAbout(ctx, about, actions...)
	if err != nil {
		return AboutAdmin{}, fmt.Errorf("guardar quiénes somos: %w", err)
	}
	return aboutAdminFrom(saved), nil
}

// SaveContact reemplaza el contacto (FR-007): dirección/correo/teléfono
// obligatorios (FR-015, con los criterios de F2), inglés opcional (analyze I6),
// publicación al guardar (FR-014) y auditoría transaccional
// `home.contact.update` (+ publish/unpublish si cambia el estado).
func (s *service) SaveContact(ctx context.Context, actorID uuid.UUID, in ContactInput) (ContactAdmin, error) {
	if err := validate.Struct(in); err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeContactUpdate, contentLabel(sectionContact, ""))
		return ContactAdmin{}, err
	}

	contact := Contact{
		AddressEs:        normalizeText(in.AddressEs),
		AddressEn:        normalizeOptional(in.AddressEn),
		Email:            strings.TrimSpace(in.Email),
		Phone:            normalizePhone(in.Phone),
		PublicationState: PublicationState(in.PublicationState),
	}

	previous, found, err := s.repository.GetHomeContact(ctx)
	if err != nil {
		return ContactAdmin{}, fmt.Errorf("leer el contacto: %w", err)
	}

	label := contentLabel(sectionContact, "")
	actions := []audit.Action{contentAction(actorID, audit.ActionHomeContactUpdate, label)}
	if found && previous.PublicationState != contact.PublicationState {
		actions = append(actions, contentAction(actorID, stateActionCode(contact.PublicationState), label))
	}

	saved, err := s.repository.UpsertHomeContact(ctx, contact, actions...)
	if err != nil {
		return ContactAdmin{}, fmt.Errorf("guardar el contacto: %w", err)
	}
	return contactAdminFrom(saved), nil
}

// --- Imágenes (T323) ---

// UploadImage guarda una imagen subida desde el panel (FR-002/FR-011, R3-8):
// el Store valida la firma binaria (JPEG/PNG/WebP) y el tamaño, y genera el
// nombre `img_<uuid>.<ext>`. La subida se audita con `home.image.upload` en la
// misma operación (analyze I8); es **fail-closed**: si el registro falla, se
// elimina el archivo y se devuelve error, de modo que nunca queda una subida sin
// traza. La subida NO cambia contenido visible: la referencia se registra al
// guardar la identidad (`home.identity.update`).
func (s *service) UploadImage(ctx context.Context, actorID uuid.UUID, data []byte) (ImageUploadResult, error) {
	if s.store == nil {
		return ImageUploadResult{}, apperr.Internal(errors.New("portada: almacén de imágenes no configurado"))
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

	action := contentAction(actorID, audit.ActionHomeImageUpload, contentLabel(sectionImage, file.Name))
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

// --- Horario (T320) ---

// CreateService da de alta un servicio del horario (FR-004): día 0–6, inicio
// "HH:MM" y fin opcional posterior (analyze C2), textos en español obligatorios,
// límite de colección (≤50) y `home.schedule.create` —aunque nazca publicada
// (analyze M5)— en la misma transacción (FR-017).
func (s *service) CreateService(ctx context.Context, actorID uuid.UUID, in ScheduleItemInput) (ScheduleItemAdmin, error) {
	if err := validate.Struct(in); err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeScheduleCreate, contentLabel(sectionSchedule, ""))
		return ScheduleItemAdmin{}, err
	}
	start, end, err := normalizeScheduleTimes(in.StartTime, in.EndTime)
	if err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeScheduleCreate, contentLabel(sectionSchedule, ""))
		return ScheduleItemAdmin{}, err
	}

	existing, err := s.repository.ListHomeServices(ctx)
	if err != nil {
		return ScheduleItemAdmin{}, fmt.Errorf("contar el horario: %w", err)
	}
	if len(existing) >= MaxScheduleItems {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeScheduleCreate, contentLabel(sectionSchedule, ""))
		return ScheduleItemAdmin{}, apperr.Invalid(
			fmt.Sprintf("No se pueden guardar más de %d servicios", MaxScheduleItems),
			apperr.WithDetails(map[string]any{"schedule": "límite alcanzado"}),
		)
	}

	service := Service{
		DayOfWeek:        in.DayOfWeek,
		StartTime:        start,
		EndTime:          end,
		NameEs:           normalizeText(in.NameEs),
		NameEn:           normalizeOptional(in.NameEn),
		DescriptionEs:    normalizeOptional(in.DescriptionEs),
		DescriptionEn:    normalizeOptional(in.DescriptionEn),
		PlaceEs:          normalizeText(in.PlaceEs),
		PlaceEn:          normalizeOptional(in.PlaceEn),
		PublicationState: publicationStateOr(in.PublicationState),
		SortOrder:        in.SortOrder,
	}

	actions := []audit.Action{contentAction(actorID, audit.ActionHomeScheduleCreate, contentLabel(sectionSchedule, service.NameEs))}
	saved, err := s.repository.InsertHomeService(ctx, service, actions...)
	if err != nil {
		return ScheduleItemAdmin{}, fmt.Errorf("crear el servicio: %w", err)
	}
	return scheduleItemAdminFrom(saved), nil
}

// UpdateService edita parcialmente un servicio (FR-004). Registra
// `home.schedule.update` si cambian los datos y `home.publish`/`home.unpublish`
// si cambia el estado —dos filas cuando ocurren ambas (analyze M5)— en la misma
// transacción. Un id inexistente → 404.
func (s *service) UpdateService(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch ScheduleItemPatch) (ScheduleItemAdmin, error) {
	if !patch.hasChanges() {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeScheduleUpdate, contentLabel(sectionSchedule, id.String()))
		return ScheduleItemAdmin{}, apperr.Invalid("No hay cambios que guardar")
	}

	current, found, err := s.repository.GetHomeServiceByID(ctx, id)
	if err != nil {
		return ScheduleItemAdmin{}, fmt.Errorf("leer el servicio: %w", err)
	}
	if !found {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeScheduleUpdate, contentLabel(sectionSchedule, id.String()))
		return ScheduleItemAdmin{}, apperr.NotFound("El servicio no existe")
	}

	updated, dataChanged, stateChanged, err := mergeSchedule(current, patch)
	if err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeScheduleUpdate, contentLabel(sectionSchedule, current.NameEs))
		return ScheduleItemAdmin{}, err
	}
	if !dataChanged && !stateChanged {
		return scheduleItemAdminFrom(current), nil
	}

	label := contentLabel(sectionSchedule, updated.NameEs)
	actions := make([]audit.Action, 0, 2)
	if dataChanged {
		actions = append(actions, contentAction(actorID, audit.ActionHomeScheduleUpdate, label))
	}
	if stateChanged {
		actions = append(actions, contentAction(actorID, stateActionCode(updated.PublicationState), label))
	}

	saved, err := s.repository.UpdateHomeService(ctx, updated, actions...)
	if err != nil {
		return ScheduleItemAdmin{}, fmt.Errorf("editar el servicio: %w", err)
	}
	return scheduleItemAdminFrom(saved), nil
}

// DeleteService elimina físicamente un servicio y registra `home.schedule.delete`
// con su `targetLabel` conservado (R3-11) en la misma transacción. Un id
// inexistente → 404.
func (s *service) DeleteService(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	current, found, err := s.repository.GetHomeServiceByID(ctx, id)
	if err != nil {
		return fmt.Errorf("leer el servicio: %w", err)
	}
	if !found {
		return apperr.NotFound("El servicio no existe")
	}
	actions := []audit.Action{contentAction(actorID, audit.ActionHomeScheduleDelete, contentLabel(sectionSchedule, current.NameEs))}
	deleted, err := s.repository.DeleteHomeService(ctx, id, actions...)
	if err != nil {
		return fmt.Errorf("eliminar el servicio: %w", err)
	}
	if !deleted {
		return apperr.NotFound("El servicio no existe")
	}
	return nil
}

// --- WhatsApp (T320) ---

// CreateWhatsappChannel da de alta un canal (FR-005): `direct` valida teléfono
// y normaliza a dígitos; `group` valida URL https de WhatsApp; límite de
// colección (≤20); duplicado exacto → 409 (UNIQUE de la BD). El alta registra
// `home.whatsapp.create` aunque nazca publicada (analyze M5).
func (s *service) CreateWhatsappChannel(ctx context.Context, actorID uuid.UUID, in WhatsappChannelInput) (WhatsappChannelAdmin, error) {
	if err := validate.Struct(in); err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeWhatsappCreate, contentLabel(sectionWhatsapp, ""))
		return WhatsappChannelAdmin{}, err
	}
	destination, err := normalizeWhatsappDestination(in.Kind, in.Destination)
	if err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeWhatsappCreate, contentLabel(sectionWhatsapp, ""))
		return WhatsappChannelAdmin{}, err
	}

	existing, err := s.repository.ListHomeWhatsappChannels(ctx)
	if err != nil {
		return WhatsappChannelAdmin{}, fmt.Errorf("contar los canales: %w", err)
	}
	if len(existing) >= MaxWhatsappChannels {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeWhatsappCreate, contentLabel(sectionWhatsapp, ""))
		return WhatsappChannelAdmin{}, apperr.Invalid(
			fmt.Sprintf("No se pueden guardar más de %d canales", MaxWhatsappChannels),
			apperr.WithDetails(map[string]any{"whatsapp": "límite alcanzado"}),
		)
	}

	channel := WhatsappChannel{
		Kind:             in.Kind,
		Destination:      destination,
		NameEs:           normalizeText(in.NameEs),
		NameEn:           normalizeOptional(in.NameEn),
		PublicationState: publicationStateOr(in.PublicationState),
		SortOrder:        in.SortOrder,
	}

	actions := []audit.Action{contentAction(actorID, audit.ActionHomeWhatsappCreate, contentLabel(sectionWhatsapp, channel.NameEs))}
	saved, err := s.repository.InsertHomeWhatsappChannel(ctx, channel, actions...)
	if err != nil {
		return WhatsappChannelAdmin{}, fmt.Errorf("crear el canal: %w", err)
	}
	return whatsappChannelAdminFrom(saved), nil
}

// UpdateWhatsappChannel edita parcialmente un canal (FR-005). Revalida el
// destino si cambia el tipo; duplicado exacto → 409. Registra
// `home.whatsapp.update` y/o `home.publish`/`home.unpublish` (analyze M5).
func (s *service) UpdateWhatsappChannel(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch WhatsappChannelPatch) (WhatsappChannelAdmin, error) {
	if !patch.hasChanges() {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeWhatsappUpdate, contentLabel(sectionWhatsapp, id.String()))
		return WhatsappChannelAdmin{}, apperr.Invalid("No hay cambios que guardar")
	}

	current, found, err := s.repository.GetHomeWhatsappChannelByID(ctx, id)
	if err != nil {
		return WhatsappChannelAdmin{}, fmt.Errorf("leer el canal: %w", err)
	}
	if !found {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeWhatsappUpdate, contentLabel(sectionWhatsapp, id.String()))
		return WhatsappChannelAdmin{}, apperr.NotFound("El canal no existe")
	}

	updated, dataChanged, stateChanged, err := mergeWhatsappChannel(current, patch)
	if err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeWhatsappUpdate, contentLabel(sectionWhatsapp, current.NameEs))
		return WhatsappChannelAdmin{}, err
	}
	if !dataChanged && !stateChanged {
		return whatsappChannelAdminFrom(current), nil
	}

	label := contentLabel(sectionWhatsapp, updated.NameEs)
	actions := make([]audit.Action, 0, 2)
	if dataChanged {
		actions = append(actions, contentAction(actorID, audit.ActionHomeWhatsappUpdate, label))
	}
	if stateChanged {
		actions = append(actions, contentAction(actorID, stateActionCode(updated.PublicationState), label))
	}

	saved, err := s.repository.UpdateHomeWhatsappChannel(ctx, updated, actions...)
	if err != nil {
		return WhatsappChannelAdmin{}, fmt.Errorf("editar el canal: %w", err)
	}
	return whatsappChannelAdminFrom(saved), nil
}

// DeleteWhatsappChannel elimina un canal y registra `home.whatsapp.delete` con
// su `targetLabel` conservado. Un id inexistente → 404.
func (s *service) DeleteWhatsappChannel(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	current, found, err := s.repository.GetHomeWhatsappChannelByID(ctx, id)
	if err != nil {
		return fmt.Errorf("leer el canal: %w", err)
	}
	if !found {
		return apperr.NotFound("El canal no existe")
	}
	actions := []audit.Action{contentAction(actorID, audit.ActionHomeWhatsappDelete, contentLabel(sectionWhatsapp, current.NameEs))}
	deleted, err := s.repository.DeleteHomeWhatsappChannel(ctx, id, actions...)
	if err != nil {
		return fmt.Errorf("eliminar el canal: %w", err)
	}
	if !deleted {
		return apperr.NotFound("El canal no existe")
	}
	return nil
}

// --- Redes (T320) ---

// CreateSocialLink da de alta el enlace de una red (FR-006): red del catálogo y
// URL https del dominio oficial; un segundo enlace para la misma red → 409
// (UNIQUE de la BD). El alta registra `home.social.create` (analyze M5).
func (s *service) CreateSocialLink(ctx context.Context, actorID uuid.UUID, in SocialLinkInput) (SocialLinkAdmin, error) {
	if err := validate.Struct(in); err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeSocialCreate, contentLabel(sectionSocial, ""))
		return SocialLinkAdmin{}, err
	}
	network, linkURL, err := normalizeSocial(in.Network, in.URL)
	if err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeSocialCreate, contentLabel(sectionSocial, ""))
		return SocialLinkAdmin{}, err
	}

	link := SocialLink{
		Network:          network,
		URL:              linkURL,
		PublicationState: publicationStateOr(in.PublicationState),
	}

	actions := []audit.Action{contentAction(actorID, audit.ActionHomeSocialCreate, contentLabel(sectionSocial, network))}
	saved, err := s.repository.InsertHomeSocialLink(ctx, link, actions...)
	if err != nil {
		return SocialLinkAdmin{}, fmt.Errorf("crear el enlace: %w", err)
	}
	return socialLinkAdminFrom(saved), nil
}

// UpdateSocialLink edita parcialmente el enlace de una red (FR-006). Revalida
// red y URL; duplicado → 409. Registra `home.social.update` y/o
// `home.publish`/`home.unpublish` (analyze M5).
func (s *service) UpdateSocialLink(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch SocialLinkPatch) (SocialLinkAdmin, error) {
	if !patch.hasChanges() {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeSocialUpdate, contentLabel(sectionSocial, id.String()))
		return SocialLinkAdmin{}, apperr.Invalid("No hay cambios que guardar")
	}

	current, found, err := s.repository.GetHomeSocialLinkByID(ctx, id)
	if err != nil {
		return SocialLinkAdmin{}, fmt.Errorf("leer el enlace: %w", err)
	}
	if !found {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeSocialUpdate, contentLabel(sectionSocial, id.String()))
		return SocialLinkAdmin{}, apperr.NotFound("El enlace no existe")
	}

	updated, dataChanged, stateChanged, err := mergeSocialLink(current, patch)
	if err != nil {
		s.recordContentFailure(ctx, actorID, audit.ActionHomeSocialUpdate, contentLabel(sectionSocial, current.Network))
		return SocialLinkAdmin{}, err
	}
	if !dataChanged && !stateChanged {
		return socialLinkAdminFrom(current), nil
	}

	label := contentLabel(sectionSocial, updated.Network)
	actions := make([]audit.Action, 0, 2)
	if dataChanged {
		actions = append(actions, contentAction(actorID, audit.ActionHomeSocialUpdate, label))
	}
	if stateChanged {
		actions = append(actions, contentAction(actorID, stateActionCode(updated.PublicationState), label))
	}

	saved, err := s.repository.UpdateHomeSocialLink(ctx, updated, actions...)
	if err != nil {
		return SocialLinkAdmin{}, fmt.Errorf("editar el enlace: %w", err)
	}
	return socialLinkAdminFrom(saved), nil
}

// DeleteSocialLink elimina un enlace y registra `home.social.delete` con su
// `targetLabel` (la red) conservado. Un id inexistente → 404.
func (s *service) DeleteSocialLink(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	current, found, err := s.repository.GetHomeSocialLinkByID(ctx, id)
	if err != nil {
		return fmt.Errorf("leer el enlace: %w", err)
	}
	if !found {
		return apperr.NotFound("El enlace no existe")
	}
	actions := []audit.Action{contentAction(actorID, audit.ActionHomeSocialDelete, contentLabel(sectionSocial, current.Network))}
	deleted, err := s.repository.DeleteHomeSocialLink(ctx, id, actions...)
	if err != nil {
		return fmt.Errorf("eliminar el enlace: %w", err)
	}
	if !deleted {
		return apperr.NotFound("El enlace no existe")
	}
	return nil
}

// --- Validación y mezcla de las ediciones parciales ---

// hasChanges indica si el PATCH trae al menos un campo (minProperties: 1). Un
// `null` explícito cuenta como cambio: limpiar un campo opcional lo es.
func (p ScheduleItemPatch) hasChanges() bool {
	return p.DayOfWeek != nil || p.StartTime != nil || p.EndTime.Set() ||
		p.NameEs != nil || p.NameEn.Set() || p.DescriptionEs.Set() || p.DescriptionEn.Set() ||
		p.PlaceEs != nil || p.PlaceEn.Set() || p.PublicationState != nil || p.SortOrder != nil
}

// hasChanges indica si el PATCH trae al menos un campo.
func (p WhatsappChannelPatch) hasChanges() bool {
	return p.NameEs != nil || p.NameEn.Set() || p.Kind != nil ||
		p.Destination != nil || p.PublicationState != nil || p.SortOrder != nil
}

// hasChanges indica si el PATCH trae al menos un campo.
func (p SocialLinkPatch) hasChanges() bool {
	return p.Network != nil || p.URL != nil || p.PublicationState != nil
}

// mergeSchedule combina el servicio actual con los campos presentes del PATCH,
// valida el rango horario resultante y devuelve si cambiaron los datos y/o el
// estado.
func mergeSchedule(current Service, patch ScheduleItemPatch) (Service, bool, bool, error) {
	updated := current
	dataChanged := false

	if patch.DayOfWeek != nil {
		if *patch.DayOfWeek < 0 || *patch.DayOfWeek > 6 {
			return Service{}, false, false, apperr.Invalid(
				"El día de la semana no es válido",
				apperr.WithDetails(map[string]any{"dayOfWeek": "debe ser un valor entre 0 y 6"}),
			)
		}
		updated.DayOfWeek = *patch.DayOfWeek
		dataChanged = true
	}
	if patch.StartTime != nil {
		start, err := normalizeStartTime(*patch.StartTime)
		if err != nil {
			return Service{}, false, false, err
		}
		updated.StartTime = start
		dataChanged = true
	}
	if patch.EndTime.Set() {
		// `endTime: null` (o "") quita la hora de fin: Value() es "".
		end, err := normalizeEndTime(patch.EndTime.Value())
		if err != nil {
			return Service{}, false, false, err
		}
		updated.EndTime = end
		dataChanged = true
	}
	if patch.NameEs != nil {
		name := normalizeText(*patch.NameEs)
		if name == "" {
			return Service{}, false, false, apperr.Invalid(
				"Revisa los datos del formulario",
				apperr.WithDetails(map[string]any{"nameEs": "Este campo es obligatorio."}),
			)
		}
		updated.NameEs = name
		dataChanged = true
	}
	if patch.NameEn.Set() {
		updated.NameEn = normalizeOptional(patch.NameEn.Value())
		dataChanged = true
	}
	if patch.DescriptionEs.Set() {
		updated.DescriptionEs = normalizeOptional(patch.DescriptionEs.Value())
		dataChanged = true
	}
	if patch.DescriptionEn.Set() {
		updated.DescriptionEn = normalizeOptional(patch.DescriptionEn.Value())
		dataChanged = true
	}
	if patch.PlaceEs != nil {
		place := normalizeText(*patch.PlaceEs)
		if place == "" {
			return Service{}, false, false, apperr.Invalid(
				"Revisa los datos del formulario",
				apperr.WithDetails(map[string]any{"placeEs": "Este campo es obligatorio."}),
			)
		}
		updated.PlaceEs = place
		dataChanged = true
	}
	if patch.PlaceEn.Set() {
		updated.PlaceEn = normalizeOptional(patch.PlaceEn.Value())
		dataChanged = true
	}
	if patch.SortOrder != nil {
		updated.SortOrder = *patch.SortOrder
		dataChanged = true
	}

	if updated.EndTime != nil && *updated.EndTime <= updated.StartTime {
		return Service{}, false, false, apperr.Invalid(
			"La hora de fin no es válida",
			apperr.WithDetails(map[string]any{"endTime": "debe ser posterior a la hora de inicio"}),
		)
	}

	stateChanged := false
	if patch.PublicationState != nil {
		state := PublicationState(*patch.PublicationState)
		if !state.Valid() {
			return Service{}, false, false, apperr.Invalid(
				"El estado de publicación no es válido",
				apperr.WithDetails(map[string]any{"publicationState": "solo se admite draft o published"}),
			)
		}
		updated.PublicationState = state
		stateChanged = state != current.PublicationState
	}
	return updated, dataChanged, stateChanged, nil
}

// mergeWhatsappChannel combina el canal actual con los campos presentes del
// PATCH y revalida el destino con el tipo resultante.
func mergeWhatsappChannel(current WhatsappChannel, patch WhatsappChannelPatch) (WhatsappChannel, bool, bool, error) {
	updated := current
	dataChanged := false

	if patch.NameEs != nil {
		name := normalizeText(*patch.NameEs)
		if name == "" {
			return WhatsappChannel{}, false, false, apperr.Invalid(
				"Revisa los datos del formulario",
				apperr.WithDetails(map[string]any{"nameEs": "Este campo es obligatorio."}),
			)
		}
		updated.NameEs = name
		dataChanged = true
	}
	if patch.NameEn.Set() {
		updated.NameEn = normalizeOptional(patch.NameEn.Value())
		dataChanged = true
	}
	if patch.Kind != nil {
		updated.Kind = *patch.Kind
		dataChanged = true
	}
	if patch.Destination != nil {
		updated.Destination = *patch.Destination
		dataChanged = true
	}
	if patch.SortOrder != nil {
		updated.SortOrder = *patch.SortOrder
		dataChanged = true
	}

	if dataChanged {
		destination, err := normalizeWhatsappDestination(updated.Kind, updated.Destination)
		if err != nil {
			return WhatsappChannel{}, false, false, err
		}
		updated.Destination = destination
	}

	stateChanged := false
	if patch.PublicationState != nil {
		state := PublicationState(*patch.PublicationState)
		if !state.Valid() {
			return WhatsappChannel{}, false, false, apperr.Invalid(
				"El estado de publicación no es válido",
				apperr.WithDetails(map[string]any{"publicationState": "solo se admite draft o published"}),
			)
		}
		updated.PublicationState = state
		stateChanged = state != current.PublicationState
	}
	return updated, dataChanged, stateChanged, nil
}

// mergeSocialLink combina el enlace actual con los campos presentes del PATCH y
// revalida red y URL resultantes.
func mergeSocialLink(current SocialLink, patch SocialLinkPatch) (SocialLink, bool, bool, error) {
	updated := current
	dataChanged := false

	network := current.Network
	rawURL := current.URL
	if patch.Network != nil {
		network = *patch.Network
		dataChanged = true
	}
	if patch.URL != nil {
		rawURL = *patch.URL
		dataChanged = true
	}
	if dataChanged {
		normalizedNetwork, normalizedURL, err := normalizeSocial(network, rawURL)
		if err != nil {
			return SocialLink{}, false, false, err
		}
		updated.Network = normalizedNetwork
		updated.URL = normalizedURL
	}

	stateChanged := false
	if patch.PublicationState != nil {
		state := PublicationState(*patch.PublicationState)
		if !state.Valid() {
			return SocialLink{}, false, false, apperr.Invalid(
				"El estado de publicación no es válido",
				apperr.WithDetails(map[string]any{"publicationState": "solo se admite draft o published"}),
			)
		}
		updated.PublicationState = state
		stateChanged = state != current.PublicationState
	}
	return updated, dataChanged, stateChanged, nil
}

// --- Validación de dominio ---

// identityImageRules exige el texto alternativo en español cuando hay imagen
// (FR-019) y acota el nombre del archivo al límite del contrato.
func identityImageRules(in IdentityInput) error {
	details := map[string]any{}
	if strings.TrimSpace(in.LogoFile) != "" && normalizeText(in.LogoAltEs) == "" {
		details["logoAltEs"] = "El texto alternativo es obligatorio cuando hay logotipo."
	}
	if strings.TrimSpace(in.CoverImageFile) != "" && normalizeText(in.CoverImageAltEs) == "" {
		details["coverImageAltEs"] = "El texto alternativo es obligatorio cuando hay imagen de portada."
	}
	if strings.TrimSpace(in.LogoFile) != "" && len(strings.TrimSpace(in.LogoFile)) > maxMediaFileName {
		details["logoFile"] = "El nombre del archivo no es válido."
	}
	if strings.TrimSpace(in.CoverImageFile) != "" && len(strings.TrimSpace(in.CoverImageFile)) > maxMediaFileName {
		details["coverImageFile"] = "El nombre del archivo no es válido."
	}
	if len(details) > 0 {
		return apperr.Invalid("Revisa los datos del formulario", apperr.WithDetails(details))
	}
	return nil
}

// normalizeScheduleTimes valida el inicio (obligatorio, "HH:MM") y el fin
// opcional (patrón y posterior al inicio, analyze C2).
func normalizeScheduleTimes(startRaw, endRaw string) (string, *string, error) {
	start, err := normalizeStartTime(startRaw)
	if err != nil {
		return "", nil, err
	}
	end, err := normalizeEndTime(endRaw)
	if err != nil {
		return "", nil, err
	}
	if end != nil && *end <= start {
		return "", nil, apperr.Invalid(
			"La hora de fin no es válida",
			apperr.WithDetails(map[string]any{"endTime": "debe ser posterior a la hora de inicio"}),
		)
	}
	return start, end, nil
}

// normalizeStartTime valida y normaliza una hora de inicio "HH:MM".
func normalizeStartTime(raw string) (string, error) {
	start := strings.TrimSpace(raw)
	if !timePattern.MatchString(start) {
		return "", apperr.Invalid(
			"La hora de inicio no es válida",
			apperr.WithDetails(map[string]any{"startTime": "usa el formato HH:MM"}),
		)
	}
	return start, nil
}

// normalizeEndTime valida el fin opcional: vacío → nil (sin hora de fin).
func normalizeEndTime(raw string) (*string, error) {
	end := strings.TrimSpace(raw)
	if end == "" {
		return nil, nil
	}
	if !timePattern.MatchString(end) {
		return nil, apperr.Invalid(
			"La hora de fin no es válida",
			apperr.WithDetails(map[string]any{"endTime": "usa el formato HH:MM"}),
		)
	}
	return &end, nil
}

// normalizeWhatsappDestination valida y normaliza el destino según el tipo:
// teléfono a dígitos para `direct`, URL https de WhatsApp para `group` (R3-6).
func normalizeWhatsappDestination(kind, destination string) (string, error) {
	switch kind {
	case KindDirect:
		if !validPhoneDestination(destination) {
			return "", apperr.Invalid(
				"El teléfono no es válido",
				apperr.WithDetails(map[string]any{"destination": "escribe un número de teléfono válido"}),
			)
		}
		return normalizePhone(destination), nil
	case KindGroup:
		trimmed := strings.TrimSpace(destination)
		if !validWhatsappGroupURL(trimmed) {
			return "", apperr.Invalid(
				"El enlace del grupo no es válido",
				apperr.WithDetails(map[string]any{"destination": "usa una URL https de chat.whatsapp.com o wa.me"}),
			)
		}
		return trimmed, nil
	default:
		return "", apperr.Invalid(
			"El tipo de canal no es válido",
			apperr.WithDetails(map[string]any{"kind": "solo se admite direct o group"}),
		)
	}
}

// normalizeSocial valida la red (catálogo) y la URL (host oficial) y devuelve
// sus valores normalizados (R3-7).
func normalizeSocial(network, rawURL string) (string, string, error) {
	normalizedNetwork := strings.ToLower(strings.TrimSpace(network))
	if !validSocialNetwork(normalizedNetwork) {
		return "", "", apperr.Invalid(
			"La red social no es válida",
			apperr.WithDetails(map[string]any{"network": "elige una red del catálogo"}),
		)
	}
	link := strings.TrimSpace(rawURL)
	if !validSocialURL(normalizedNetwork, link) {
		return "", "", apperr.Invalid(
			"El enlace no es válido",
			apperr.WithDetails(map[string]any{"url": "usa una URL https del dominio oficial de la red"}),
		)
	}
	return normalizedNetwork, link, nil
}

// deleteReplacedImage borra best-effort la imagen anterior si cambió (R3-8).
func (s *service) deleteReplacedImage(ctx context.Context, previous, current *string) {
	if previous == nil || current == nil {
		if previous != nil && current == nil {
			s.deleteFile(ctx, *previous)
		}
		return
	}
	if *previous != *current {
		s.deleteFile(ctx, *previous)
	}
}

// --- Agregado del panel (T340, analyze C1) ---

// GetPortadaAdmin devuelve el estado completo del módulo para precargar el panel
// (FR-011, US3 esc. 7): identidad, «quiénes somos» y contacto (null hasta el
// primer guardado), horario, WhatsApp y redes —en ambos idiomas y con su
// `publicationState` por elemento, borradores incluidos—. Las colecciones van
// acotadas a los límites de escritura y en su sobre `{items}` (nunca null). Es
// la única vista que expone borradores y exige el permiso `portada`.
func (s *service) GetPortadaAdmin(ctx context.Context) (PortadaAdmin, error) {
	out := PortadaAdmin{
		Schedule: ScheduleItemsAdmin{Items: []ScheduleItemAdmin{}},
		Whatsapp: WhatsappChannelsAdmin{Items: []WhatsappChannelAdmin{}},
		Socials:  SocialLinksAdmin{Items: []SocialLinkAdmin{}},
	}

	if identity, ok, err := s.repository.GetHomeIdentity(ctx); err != nil {
		return PortadaAdmin{}, fmt.Errorf("leer la identidad: %w", err)
	} else if ok {
		admin := identityAdminFrom(identity)
		out.Identity = &admin
	}

	if about, ok, err := s.repository.GetHomeAbout(ctx); err != nil {
		return PortadaAdmin{}, fmt.Errorf("leer quiénes somos: %w", err)
	} else if ok {
		admin := aboutAdminFrom(about)
		out.About = &admin
	}

	if contact, ok, err := s.repository.GetHomeContact(ctx); err != nil {
		return PortadaAdmin{}, fmt.Errorf("leer el contacto: %w", err)
	} else if ok {
		admin := contactAdminFrom(contact)
		out.Contact = &admin
	}

	services, err := s.repository.ListHomeServices(ctx)
	if err != nil {
		return PortadaAdmin{}, fmt.Errorf("leer el horario: %w", err)
	}
	for i, service := range services {
		if i >= MaxScheduleItems {
			break
		}
		out.Schedule.Items = append(out.Schedule.Items, scheduleItemAdminFrom(service))
	}

	channels, err := s.repository.ListHomeWhatsappChannels(ctx)
	if err != nil {
		return PortadaAdmin{}, fmt.Errorf("leer los canales de WhatsApp: %w", err)
	}
	for i, channel := range channels {
		if i >= MaxWhatsappChannels {
			break
		}
		out.Whatsapp.Items = append(out.Whatsapp.Items, whatsappChannelAdminFrom(channel))
	}

	links, err := s.repository.ListHomeSocialLinks(ctx)
	if err != nil {
		return PortadaAdmin{}, fmt.Errorf("leer las redes: %w", err)
	}
	for _, link := range links {
		out.Socials.Items = append(out.Socials.Items, socialLinkAdminFrom(link))
	}

	return out, nil
}

// --- Conversiones entidad → DTO del panel ---

// identityAdminFrom traduce la identidad guardada a su DTO del panel.
func identityAdminFrom(identity Identity) IdentityAdmin {
	return IdentityAdmin{
		NameEs:           identity.NameEs,
		NameEn:           identity.NameEn,
		TaglineEs:        identity.TaglineEs,
		TaglineEn:        identity.TaglineEn,
		MissionEs:        identity.MissionEs,
		MissionEn:        identity.MissionEn,
		VisionEs:         identity.VisionEs,
		VisionEn:         identity.VisionEn,
		LogoFile:         identity.LogoFile,
		LogoAltEs:        identity.LogoAltEs,
		LogoAltEn:        identity.LogoAltEn,
		CoverImageFile:   identity.CoverImageFile,
		CoverImageAltEs:  identity.CoverImageAltEs,
		CoverImageAltEn:  identity.CoverImageAltEn,
		PublicationState: string(identity.PublicationState),
		CreatedAt:        identity.CreatedAt,
		UpdatedAt:        identity.UpdatedAt,
	}
}

// aboutAdminFrom traduce «quiénes somos» a su DTO del panel.
func aboutAdminFrom(about About) AboutAdmin {
	return AboutAdmin{
		TextEs:           about.TextEs,
		TextEn:           about.TextEn,
		PublicationState: string(about.PublicationState),
		CreatedAt:        about.CreatedAt,
		UpdatedAt:        about.UpdatedAt,
	}
}

// contactAdminFrom traduce el contacto a su DTO del panel.
func contactAdminFrom(contact Contact) ContactAdmin {
	return ContactAdmin{
		AddressEs:        contact.AddressEs,
		AddressEn:        contact.AddressEn,
		Email:            contact.Email,
		Phone:            contact.Phone,
		PublicationState: string(contact.PublicationState),
		CreatedAt:        contact.CreatedAt,
		UpdatedAt:        contact.UpdatedAt,
	}
}

// scheduleItemAdminFrom traduce un servicio a su DTO del panel.
func scheduleItemAdminFrom(service Service) ScheduleItemAdmin {
	return ScheduleItemAdmin{
		ID:               service.ID.String(),
		DayOfWeek:        service.DayOfWeek,
		StartTime:        service.StartTime,
		EndTime:          service.EndTime,
		NameEs:           service.NameEs,
		NameEn:           service.NameEn,
		DescriptionEs:    service.DescriptionEs,
		DescriptionEn:    service.DescriptionEn,
		PlaceEs:          service.PlaceEs,
		PlaceEn:          service.PlaceEn,
		PublicationState: string(service.PublicationState),
		SortOrder:        service.SortOrder,
		CreatedAt:        service.CreatedAt,
		UpdatedAt:        service.UpdatedAt,
	}
}

// whatsappChannelAdminFrom traduce un canal a su DTO del panel.
func whatsappChannelAdminFrom(channel WhatsappChannel) WhatsappChannelAdmin {
	return WhatsappChannelAdmin{
		ID:               channel.ID.String(),
		NameEs:           channel.NameEs,
		NameEn:           channel.NameEn,
		Kind:             channel.Kind,
		Destination:      channel.Destination,
		PublicationState: string(channel.PublicationState),
		SortOrder:        channel.SortOrder,
		CreatedAt:        channel.CreatedAt,
		UpdatedAt:        channel.UpdatedAt,
	}
}

// socialLinkAdminFrom traduce un enlace a su DTO del panel.
func socialLinkAdminFrom(link SocialLink) SocialLinkAdmin {
	return SocialLinkAdmin{
		ID:               link.ID.String(),
		Network:          link.Network,
		URL:              link.URL,
		PublicationState: string(link.PublicationState),
		CreatedAt:        link.CreatedAt,
		UpdatedAt:        link.UpdatedAt,
	}
}
