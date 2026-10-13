package ministerios

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
)

// fakeRepository es el repositorio falso de los tests del service: un almacén en
// memoria que imita el comportamiento observable del repository real (solo
// publicado en la vía pública, orden `sort_order, id`, tope, duplicado exacto,
// borradores incluidos en el panel) y recuerda lo que se le pidió, para poder
// comprobar los parámetros (tope del bloque) y las filas de auditoría.
//
// La vía pública **no enmascara** la ficha del encargado: entrega la fila tal y
// como está guardada (como hace la consulta del panel). Así la prueba del
// service demuestra que la privacidad de FR-026/SC-014 se sostiene **también**
// en el service, sin depender de la máscara del SQL —que se verifica contra
// PostgreSQL en repository_public_integration_test.go (T609)—.
type fakeRepository struct {
	ministries []Ministry

	// Auditoría: `recorded` son las acciones best-effort (denegaciones y
	// rechazos) y `committed` las que acompañan a una mutación en su transacción.
	recorded  []audit.Action
	committed []audit.Action

	// Guionización de errores y de la detección de duplicados. `readErr` afecta a
	// las lecturas públicas y a los listados del panel; `getErr` solo al detalle
	// (para poder fallar al releer DESPUÉS de escribir); `duplicateCheckErr` solo
	// a la comprobación del duplicado.
	recordErr         error
	mutationErr       error
	readErr           error
	getErr            error
	duplicateCheckErr error
	duplicate         bool

	// Observaciones de las consultas.
	publicLimits []int
}

// compile-time check: el falso satisface el puerto del service.
var _ Repository = (*fakeRepository)(nil)

// newFakeRepository construye el almacén en memoria vacío.
func newFakeRepository() *fakeRepository { return &fakeRepository{} }

// seed agrega ministerios al almacén sin pasar por el service (los datos ya
// guardados de las pruebas).
func (f *fakeRepository) seed(ministries ...Ministry) {
	f.ministries = append(f.ministries, ministries...)
}

// published devuelve los ministerios publicados ordenados como el SQL
// (`sort_order`, `id`) y acotados al tope pedido.
func (f *fakeRepository) published(limit int) []Ministry {
	items := make([]Ministry, 0, len(f.ministries))
	for _, ministry := range f.ministries {
		if ministry.PublicationState == StatePublished {
			items = append(items, ministry)
		}
	}
	sortMinistries(items)
	if limit >= 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

// sortMinistries ordena por `sort_order, id` (el orden manual del módulo, que en
// producción aplica el `ORDER BY`).
func sortMinistries(items []Ministry) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].SortOrder != items[j].SortOrder {
			return items[i].SortOrder < items[j].SortOrder
		}
		return items[i].ID.String() < items[j].ID.String()
	})
}

// findIndex devuelve la posición de un ministerio por id (-1 si no existe).
func (f *fakeRepository) findIndex(id uuid.UUID) int {
	for i := range f.ministries {
		if f.ministries[i].ID == id {
			return i
		}
	}
	return -1
}

// sharedWith indica si OTRO ministerio usa el mismo enlace (R6-11): el aviso
// que el panel recalcula en cada lectura.
func (f *fakeRepository) sharedWith(id uuid.UUID, whatsappURL *string) bool {
	if whatsappURL == nil {
		return false
	}
	for _, other := range f.ministries {
		if other.ID == id || other.WhatsappURL == nil {
			continue
		}
		if *other.WhatsappURL == *whatsappURL {
			return true
		}
	}
	return false
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

func (f *fakeRepository) ListPublishedMinistries(_ context.Context, limit int) ([]Ministry, error) {
	f.publicLimits = append(f.publicLimits, limit)
	if f.readErr != nil {
		return nil, f.readErr
	}
	return f.published(limit), nil
}

// --- Panel ---

func (f *fakeRepository) InsertMinistry(_ context.Context, ministry Ministry, actions ...audit.Action) (Ministry, error) {
	if f.mutationErr != nil {
		return Ministry{}, f.mutationErr
	}
	if f.duplicate {
		return Ministry{}, apperr.Conflict(messageDuplicate)
	}
	ministry.ID = uuid.New()
	ministry.CreatedAt, ministry.UpdatedAt = time.Now(), time.Now()
	f.ministries = append(f.ministries, ministry)
	f.committed = append(f.committed, actions...)
	return ministry, nil
}

func (f *fakeRepository) UpdateMinistry(_ context.Context, ministry Ministry, actions ...audit.Action) (Ministry, error) {
	if f.mutationErr != nil {
		return Ministry{}, f.mutationErr
	}
	if f.duplicate {
		return Ministry{}, apperr.Conflict(messageDuplicate)
	}
	index := f.findIndex(ministry.ID)
	if index < 0 {
		return Ministry{}, apperr.NotFound("El ministerio no existe")
	}
	ministry.CreatedAt = f.ministries[index].CreatedAt
	ministry.UpdatedAt = time.Now()
	f.ministries[index] = ministry
	f.committed = append(f.committed, actions...)
	return ministry, nil
}

func (f *fakeRepository) DeleteMinistry(_ context.Context, id uuid.UUID, actions ...audit.Action) (bool, error) {
	if f.mutationErr != nil {
		return false, f.mutationErr
	}
	index := f.findIndex(id)
	if index < 0 {
		return false, nil
	}
	f.ministries = append(f.ministries[:index], f.ministries[index+1:]...)
	f.committed = append(f.committed, actions...)
	return true, nil
}

func (f *fakeRepository) GetMinistryByID(_ context.Context, id uuid.UUID) (MinistryListItem, bool, error) {
	if f.getErr != nil {
		return MinistryListItem{}, false, f.getErr
	}
	if f.readErr != nil {
		return MinistryListItem{}, false, f.readErr
	}
	index := f.findIndex(id)
	if index < 0 {
		return MinistryListItem{}, false, nil
	}
	ministry := f.ministries[index]
	return MinistryListItem{
		Ministry:          ministry,
		WhatsappURLShared: f.sharedWith(id, ministry.WhatsappURL),
	}, true, nil
}

func (f *fakeRepository) ListMinistries(_ context.Context, state *PublicationState, params paginate.Params) ([]MinistryListItem, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	items := make([]Ministry, 0, len(f.ministries))
	for _, ministry := range f.ministries {
		if state != nil && ministry.PublicationState != *state {
			continue
		}
		items = append(items, ministry)
	}
	sortMinistries(items)
	if params.Offset < len(items) {
		items = items[params.Offset:]
	} else {
		items = nil
	}
	if len(items) > params.Limit {
		items = items[:params.Limit]
	}

	out := make([]MinistryListItem, 0, len(items))
	for _, ministry := range items {
		out = append(out, MinistryListItem{
			Ministry:          ministry,
			WhatsappURLShared: f.sharedWith(ministry.ID, ministry.WhatsappURL),
		})
	}
	return out, nil
}

func (f *fakeRepository) CountMinistries(_ context.Context, state *PublicationState) (int64, error) {
	if f.readErr != nil {
		return 0, f.readErr
	}
	var total int64
	for _, ministry := range f.ministries {
		if state != nil && ministry.PublicationState != *state {
			continue
		}
		total++
	}
	return total, nil
}

func (f *fakeRepository) ExistsMinistryDuplicate(_ context.Context, nameEs, descriptionEs string, leaderNameEs, whatsappURL *string, excludeID *uuid.UUID) (bool, error) {
	if f.duplicateCheckErr != nil {
		return false, f.duplicateCheckErr
	}
	if f.readErr != nil {
		return false, f.readErr
	}
	if f.duplicate {
		return true, nil
	}
	for _, ministry := range f.ministries {
		if excludeID != nil && ministry.ID == *excludeID {
			continue
		}
		if sameText(ministry.NameEs, &nameEs) &&
			sameText(ministry.DescriptionEs, &descriptionEs) &&
			sameTextPtr(ministry.LeaderNameEs, leaderNameEs) &&
			sameTextPtr(ministry.WhatsappURL, whatsappURL) {
			return true, nil
		}
	}
	return false, nil
}

// sameText compara dos valores ya normalizados (el duplicado exacto del módulo
// se detecta en SQL con `btrim` + minúsculas; el falso compara en minúsculas).
func sameText(stored string, value *string) bool {
	if value == nil {
		return strings.TrimSpace(stored) == ""
	}
	return strings.EqualFold(strings.TrimSpace(stored), strings.TrimSpace(*value))
}

// sameTextPtr compara dos opcionales (nil = sin dato).
func sameTextPtr(stored, value *string) bool {
	if stored == nil || value == nil {
		return stored == nil && value == nil
	}
	return strings.EqualFold(*stored, *value)
}
