package donaciones

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
	"simiente-santa/backend/internal/platform/validate"
)

// service_admin.go implementa la gestión del módulo desde el panel (T715): alta,
// edición parcial, borrado, detalle, listado paginado y los dos textos. No
// conoce HTTP ni SQL: valida (FR-014), normaliza (analyze I6), «publica al
// guardar» (FR-013), aplica el **`PATCH` que fusiona los campos presentes** sobre
// la fila actual y escribe la fila completa ya fusionada (R7-15), y compone las
// filas de auditoría con la regla exacta de códigos (R7-10), que el repository
// inserta en la MISMA transacción (FR-017).

// --- DTOs del panel (contrato `DonacionAdmin`, `DonacionAdminList`) ---

// DonacionAdmin es la ficha completa del panel (FR-010/FR-019): ambos idiomas del
// titular (dato guardado ≠ dato publicado), la moneda (solo IBAN), el estado de
// publicación y el número **normalizado** (FR-015).
type DonacionAdmin struct {
	ID               string    `json:"id"`
	Kind             string    `json:"kind"`
	Number           string    `json:"number"`
	Institution      string    `json:"institution"`
	HolderEs         *string   `json:"holderEs"`
	HolderEn         *string   `json:"holderEn"`
	Currency         *string   `json:"currency"`
	PublicationState string    `json:"publicationState"`
	SortOrder        int       `json:"sortOrder"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// DonacionAdminList es el listado paginado del panel (§8.1.2/§8.1.3).
type DonacionAdminList struct {
	Items  []DonacionAdmin `json:"items"`
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// DonacionTextoAdmin es la ficha de un texto del panel (FR-021): ambos idiomas y
// su propio estado de publicación (Q10).
type DonacionTextoAdmin struct {
	TextKey          string    `json:"textKey"`
	TextEs           string    `json:"textEs"`
	TextEn           *string   `json:"textEn"`
	PublicationState string    `json:"publicationState"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// DonacionTextosAdmin agrupa los dos textos de la sección (FR-021/Q5/Q11):
// singletons por clave, en `null` hasta el primer guardado.
type DonacionTextosAdmin struct {
	Intro *DonacionTextoAdmin `json:"intro"`
	Usage *DonacionTextoAdmin `json:"usage"`
}

// --- Gestión del panel: métodos (FR-010) ---

// CreateDonacion da de alta un método (FR-010/FR-014): validación completa —el
// `400` lleva `details` por campo y **no guarda nada a medias**—, normalización
// del número (FR-015/FR-016) y duplicado exacto → `409 conflict` (P7-10).
//
// Auditoría (R7-10): el alta registra **solo `donation.create`** —una sola fila—
// también cuando nace publicado (cero filas `donation.publish`).
func (s *service) CreateDonacion(ctx context.Context, actorID uuid.UUID, in DonacionInput) (DonacionAdmin, error) {
	label := contentLabel("")
	fail := func(err error) (DonacionAdmin, error) {
		s.recordContentFailure(ctx, actorID, audit.ActionDonationCreate, label)
		return DonacionAdmin{}, err
	}

	if err := validate.Struct(in); err != nil {
		return fail(err)
	}
	method, err := methodFromInput(in)
	if err != nil {
		return fail(err)
	}
	label = methodLabel(method)

	duplicate, err := s.repository.ExistsDonationMethodDuplicate(ctx, method.Kind, method.Number, nil)
	if err != nil {
		return DonacionAdmin{}, fmt.Errorf("comprobar el duplicado: %w", err)
	}
	if duplicate {
		return fail(apperr.Conflict(messageDuplicate))
	}

	actions := []audit.Action{contentAction(actorID, audit.ActionDonationCreate, label)}
	saved, err := s.repository.InsertDonationMethod(ctx, method, actions...)
	if err != nil {
		// La base es la última línea de defensa del duplicado: su rechazo también
		// queda registrado (FR-017/R7-10).
		if isConflict(err) {
			s.recordContentFailure(ctx, actorID, audit.ActionDonationCreate, label)
		}
		return DonacionAdmin{}, fmt.Errorf("crear el método de donación: %w", err)
	}
	return donacionAdminFrom(saved), nil
}

// UpdateDonacion edita parcialmente un método (FR-010): fusiona los campos
// presentes sobre la fila actual y escribe la fila completa ya fusionada
// (R7-15), revalidando el resultado con las mismas reglas del alta (FR-014). Un
// id inexistente → `404`; un duplicado exacto con **otro** método → `409` (el
// chequeo excluye la fila que se edita: editar sin cambiar el número responde
// `200`).
//
// Auditoría (R7-10): `donation.update` si cambian datos (incluido `sortOrder`),
// `donation.publish`/`donation.unpublish` si cambia el estado —una fila por
// operación, ambas en la misma transacción—.
func (s *service) UpdateDonacion(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch DonacionPatch) (DonacionAdmin, error) {
	label := contentLabel(id.String())
	fail := func(err error) (DonacionAdmin, error) {
		s.recordContentFailure(ctx, actorID, audit.ActionDonationUpdate, label)
		return DonacionAdmin{}, err
	}

	current, found, err := s.repository.GetDonationMethodByID(ctx, id)
	if err != nil {
		return DonacionAdmin{}, fmt.Errorf("leer el método de donación: %w", err)
	}
	if !found {
		return fail(apperr.NotFound(messageNotFound))
	}
	label = methodLabel(current)

	updated, dataChanged, stateChanged, err := mergeDonacion(current, patch)
	if err != nil {
		return fail(err)
	}
	if !dataChanged && !stateChanged {
		return donacionAdminFrom(current), nil
	}
	label = methodLabel(updated)

	if dataChanged {
		duplicate, err := s.repository.ExistsDonationMethodDuplicate(ctx, updated.Kind, updated.Number, &id)
		if err != nil {
			return DonacionAdmin{}, fmt.Errorf("comprobar el duplicado: %w", err)
		}
		if duplicate {
			return fail(apperr.Conflict(messageDuplicate))
		}
	}

	actions := make([]audit.Action, 0, 2)
	if dataChanged {
		actions = append(actions, contentAction(actorID, audit.ActionDonationUpdate, label))
	}
	if stateChanged {
		actions = append(actions, contentAction(actorID, stateCode(updated.PublicationState), label))
	}

	saved, err := s.repository.UpdateDonationMethod(ctx, updated, actions...)
	if err != nil {
		if isConflict(err) {
			s.recordContentFailure(ctx, actorID, audit.ActionDonationUpdate, label)
		}
		return DonacionAdmin{}, fmt.Errorf("editar el método de donación: %w", err)
	}
	return donacionAdminFrom(saved), nil
}

// DeleteDonacion elimina definitivamente un método (FR-010) y registra
// `donation.delete` con su etiqueta (R7-10) en la misma transacción. Un id
// inexistente → `404`.
func (s *service) DeleteDonacion(ctx context.Context, actorID uuid.UUID, id uuid.UUID) error {
	current, found, err := s.repository.GetDonationMethodByID(ctx, id)
	if err != nil {
		return fmt.Errorf("leer el método de donación: %w", err)
	}
	if !found {
		return apperr.NotFound(messageNotFound)
	}

	actions := []audit.Action{contentAction(actorID, audit.ActionDonationDelete, methodLabel(current))}
	deleted, err := s.repository.DeleteDonationMethod(ctx, id, actions...)
	if err != nil {
		return fmt.Errorf("eliminar el método de donación: %w", err)
	}
	if !deleted {
		return apperr.NotFound(messageNotFound)
	}
	return nil
}

// GetDonacion devuelve la ficha completa del panel (FR-010), borradores
// incluidos. Un id inexistente → `404`.
func (s *service) GetDonacion(ctx context.Context, id uuid.UUID) (DonacionAdmin, error) {
	method, found, err := s.repository.GetDonationMethodByID(ctx, id)
	if err != nil {
		return DonacionAdmin{}, fmt.Errorf("leer el método de donación: %w", err)
	}
	if !found {
		return DonacionAdmin{}, apperr.NotFound(messageNotFound)
	}
	return donacionAdminFrom(method), nil
}

// ListDonaciones devuelve el listado paginado del panel (FR-010): todos los
// métodos —borradores incluidos— en el orden del módulo (`CASE kind`,
// `sort_order`, `id`: el mismo que ve el visitante, P7-20/US2 esc. 10), con
// filtro opcional por estado de publicación.
func (s *service) ListDonaciones(ctx context.Context, state *PublicationState, params paginate.Params) (DonacionAdminList, error) {
	total, err := s.repository.CountDonationMethods(ctx, state)
	if err != nil {
		return DonacionAdminList{}, fmt.Errorf("contar los métodos de donación: %w", err)
	}
	items, err := s.repository.ListDonationMethods(ctx, state, params)
	if err != nil {
		return DonacionAdminList{}, fmt.Errorf("listar los métodos de donación: %w", err)
	}

	out := make([]DonacionAdmin, 0, len(items))
	for _, item := range items {
		out = append(out, donacionAdminFrom(item))
	}
	return DonacionAdminList{Items: out, Total: total, Limit: params.Limit, Offset: params.Offset}, nil
}

// --- Textos del panel (FR-021, singleton por clave, sin borrado — P7-17) ---

// GetTextos devuelve los dos textos de la sección (FR-021/Q5/Q11) con su propio
// estado: cada clave viene en `null` hasta el primer guardado.
func (s *service) GetTextos(ctx context.Context) (DonacionTextosAdmin, error) {
	texts, err := s.repository.ListDonationTexts(ctx)
	if err != nil {
		return DonacionTextosAdmin{}, fmt.Errorf("listar los textos: %w", err)
	}

	var out DonacionTextosAdmin
	for _, text := range texts {
		admin := donacionTextoAdminFrom(text)
		switch text.TextKey {
		case TextKeyIntro:
			out.Intro = &admin
		case TextKeyUsage:
			out.Usage = &admin
		}
	}
	return out, nil
}

// PutTexto crea o reemplaza el singleton de la clave indicada (FR-021): valida
// (`textEs` obligatorio, `textEn` opcional, ≤2.000 por idioma), guarda íntegro y
// registra su auditoría (R7-10). No hay operación de borrado: un texto se retira
// poniéndolo en `draft`.
//
// Un `PUT` idéntico al texto ya guardado (mismo contenido y estado) no es una
// operación: devuelve la fila actual **sin escribir** —`updated_at` incluido— y
// **sin registrar** auditoría (FR-017/SC-011), igual que `UpdateDonacion` ante un
// `PATCH` sin cambios.
//
// Auditoría: data (alta o edición del singleton) → `donation.text.update`; el
// cambio de estado → `donation.text.publish`/`donation.text.unpublish` (solo
// cuando hay estado previo distinto); ambos a la vez → dos filas.
func (s *service) PutTexto(ctx context.Context, actorID uuid.UUID, key TextKey, in DonacionTextoInput) (DonacionTextoAdmin, error) {
	label := textLabel(key)
	fail := func(err error) (DonacionTextoAdmin, error) {
		s.recordContentFailure(ctx, actorID, audit.ActionDonationTextUpdate, label)
		return DonacionTextoAdmin{}, err
	}

	if !key.Valid() {
		return fail(apperr.Invalid(
			validate.MessageInvalid,
			apperr.WithDetails(map[string]any{"textKey": "la clave debe ser intro o usage"}),
		))
	}
	if err := validate.Struct(in); err != nil {
		return fail(err)
	}
	texto, err := textoFromInput(key, in)
	if err != nil {
		return fail(err)
	}

	current, found, err := s.repository.GetDonationTextByKey(ctx, key)
	if err != nil {
		return DonacionTextoAdmin{}, fmt.Errorf("leer el texto: %w", err)
	}

	dataChanged := true
	stateChanged := false
	if found {
		dataChanged = current.BodyEs != texto.BodyEs || !sameOptional(current.BodyEn, texto.BodyEn)
		stateChanged = current.PublicationState != texto.PublicationState
		if !dataChanged && !stateChanged {
			// Un PUT idéntico no es una operación: no se escribe (`updated_at`
			// incluido) ni se registra auditoría (FR-017/SC-011), igual que
			// `UpdateDonacion` ante un PATCH sin cambios.
			return donacionTextoAdminFrom(current), nil
		}
	}

	actions := make([]audit.Action, 0, 2)
	if dataChanged {
		actions = append(actions, contentAction(actorID, audit.ActionDonationTextUpdate, label))
	}
	if stateChanged {
		actions = append(actions, contentAction(actorID, textStateCode(texto.PublicationState), label))
	}

	saved, err := s.repository.UpsertDonationText(ctx, texto, actions...)
	if err != nil {
		return DonacionTextoAdmin{}, fmt.Errorf("guardar el texto: %w", err)
	}
	return donacionTextoAdminFrom(saved), nil
}

// --- Mezcla del PATCH (R7-15) ---

// mergeDonacion fusiona los campos presentes del PATCH sobre la fila actual y
// devuelve el método resultante, si cambiaron los datos y si cambió el estado de
// publicación. El resultado se revalida con las mismas reglas del alta (FR-014)
// ANTES de escribir: la transacción guarda la fila completa ya fusionada. Un
// `null` explícito borra el opcional (`holderEs`, `holderEn`, `currency`).
func mergeDonacion(current MetodoDonacion, patch DonacionPatch) (MetodoDonacion, bool, bool, error) {
	updated := current
	dataChanged := false

	if patch.Kind.Set() && !patch.Kind.Null() {
		updated.Kind = MetodoKind(strings.TrimSpace(patch.Kind.Value()))
		dataChanged = true
	}
	if patch.Number.Set() && !patch.Number.Null() {
		updated.Number = patch.Number.Value()
		dataChanged = true
	}
	if patch.Institution.Set() {
		updated.Institution = normalizeText(patch.Institution.Value())
		dataChanged = true
	}
	if patch.HolderEs.Set() {
		updated.HolderEs = normalizeOptional(patch.HolderEs.Value())
		dataChanged = true
	}
	if patch.HolderEn.Set() {
		updated.HolderEn = normalizeOptional(patch.HolderEn.Value())
		dataChanged = true
	}
	if patch.Currency.Set() {
		if patch.Currency.Null() {
			updated.Currency = nil
		} else {
			currency := Currency(strings.TrimSpace(patch.Currency.Value()))
			updated.Currency = &currency
		}
		dataChanged = true
	}
	if patch.SortOrder.Set() && !patch.SortOrder.Null() {
		updated.SortOrder = patch.SortOrder.Value()
		dataChanged = true
	}

	stateChanged := false
	if patch.PublicationState.Set() && !patch.PublicationState.Null() {
		state := PublicationState(strings.TrimSpace(patch.PublicationState.Value()))
		if !state.Valid() {
			return MetodoDonacion{}, false, false, apperr.Invalid(
				validate.MessageInvalid,
				apperr.WithDetails(map[string]any{"publicationState": "solo se admite draft o published"}),
			)
		}
		updated.PublicationState = state
		stateChanged = state != current.PublicationState
	}

	if err := validateMethod(&updated); err != nil {
		return MetodoDonacion{}, false, false, err
	}
	return updated, dataChanged, stateChanged, nil
}

// --- Helpers ---

// isConflict indica si el error es un conflicto de dominio (409): el rechazo de
// un duplicado detectado por la BASE también queda en la auditoría (FR-017).
func isConflict(err error) bool {
	var domainErr *apperr.Error
	return errors.As(err, &domainErr) && domainErr.Kind == apperr.KindConflict
}

// sameOptional compara dos opcionales (nil = sin dato).
func sameOptional(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// donacionAdminFrom traduce la fila del panel a su DTO (ambos idiomas del titular
// y la moneda solo en IBAN).
func donacionAdminFrom(method MetodoDonacion) DonacionAdmin {
	dto := DonacionAdmin{
		ID:               method.ID.String(),
		Kind:             string(method.Kind),
		Number:           method.Number,
		Institution:      method.Institution,
		HolderEs:         method.HolderEs,
		HolderEn:         method.HolderEn,
		PublicationState: string(method.PublicationState),
		SortOrder:        method.SortOrder,
		CreatedAt:        method.CreatedAt,
		UpdatedAt:        method.UpdatedAt,
	}
	if method.Currency != nil {
		currency := string(*method.Currency)
		dto.Currency = &currency
	}
	return dto
}

// donacionTextoAdminFrom traduce la fila de un texto a su DTO del panel.
func donacionTextoAdminFrom(text TextoDonacion) DonacionTextoAdmin {
	return DonacionTextoAdmin{
		TextKey:          string(text.TextKey),
		TextEs:           text.BodyEs,
		TextEn:           text.BodyEn,
		PublicationState: string(text.PublicationState),
		CreatedAt:        text.CreatedAt,
		UpdatedAt:        text.UpdatedAt,
	}
}
