package medios

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/validate"
)

// service_admin.go implementa la gestión del módulo desde el panel (T918/T919):
// el listado (con el registro de la última búsqueda), la edición de la
// **descripción corta** y la **visibilidad**, la **unión por puntero** de dos
// entradas y la **resolución de marcas** de los dos ejes.
//
// No conoce HTTP ni SQL: valida (FR-019/FR-023), normaliza (analyze I6), fusiona
// el `PATCH` sobre la fila actual y escribe la fila completa ya fusionada
// (patrón `UpdateDonationMethod` de F7), y compone las filas de auditoría con la
// regla exacta de códigos de FR-022, que el repository inserta en la MISMA
// transacción.

// errProveedorRepetido es la causa interna de una unión imposible (las dos
// fichas traen fuente del mismo proveedor): el mensaje seguro lo arma el
// service y esta causa nunca se serializa.
var errProveedorRepetido = errors.New("medios: las dos fichas traen fuente del mismo proveedor")

// --- Etiquetas y helpers de auditoría del dominio (R7-10/000015) ---

const (
	// moduleLabel es la etiqueta legible del módulo en `targetLabel`.
	moduleLabel = "Medios"
)

// contenidoAction construye una acción de auditoría de contenido (`result`
// success), la unidad que el repository inserta dentro de la transacción de la
// mutación (FR-022).
func contenidoAction(actorID uuid.UUID, code, label string) audit.Action {
	return audit.Action{
		ActorUserID: &actorID,
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultSuccess,
	}
}

// etiquetaContenido arma `Medios · <elemento>` (R7-10); con un elemento vacío
// queda el módulo solo (`Medios`). El recorte a 254 es la última defensa para que
// ninguna etiqueta pueda romper el `CHECK` de `admin_actions` (el nombre de una
// ficha admite 300 caracteres).
func etiquetaContenido(elemento string) string {
	elemento = colapsarEspacios(elemento)
	if elemento == "" {
		return moduleLabel
	}
	label := moduleLabel + " · " + elemento
	if runes := []rune(label); len(runes) > maxTargetLabel {
		label = string(runes[:maxTargetLabel])
	}
	return label
}

// etiquetaEpisodio arma `Medios · <nombre del episodio>` (R7-10): el objetivo de
// la mayoría de las operaciones del módulo.
func etiquetaEpisodio(episodio Episodio) string {
	return etiquetaContenido(episodio.Titulo)
}

// visibilidadCode traduce un cambio de visibilidad a su código (FR-022): solo el
// cambio de `visibility` emite `media.hide`/`media.show`.
func visibilidadCode(visibilidad Visibilidad) string {
	if visibilidad == VisibilidadVisible {
		return audit.ActionMediaShow
	}
	return audit.ActionMediaHide
}

// codigoIntentoMedio traduce el `PATCH` recibido a la acción que se **intentaba**
// (FR-022): un cuerpo que solo trae visibilidad es un ocultar/mostrar —también
// cuando la operación acaba fallando— y cualquier otro (o una visibilidad fuera
// del catálogo, que no identifica ninguna operación) es la edición de la
// descripción. Es el mismo criterio del camino de éxito.
func codigoIntentoMedio(patch MedioPatch) string {
	if patch.DescriptionEs == nil && patch.DescriptionEn == nil && patch.Visibility != nil {
		if visibilidad := Visibilidad(strings.TrimSpace(*patch.Visibility)); visibilidad.Valid() {
			return visibilidadCode(visibilidad)
		}
	}
	return audit.ActionMediaDescriptionUpdate
}

// recordActionBestEffort persiste una acción que ya va a fallar (rechazo o
// denegación) sin cambiar la respuesta: su fallo se queda en el log (R7-10).
func (s *service) recordActionBestEffort(ctx context.Context, action audit.Action) {
	if err := s.repository.RecordAction(ctx, action); err != nil {
		s.logError(ctx, "no se pudo registrar la acción administrativa", err)
	}
}

// recordContentFailure registra best-effort el rechazo de una mutación del
// módulo (`result='failure'`), con el mismo objetivo que habría tenido la acción
// (R7-10).
func (s *service) recordContentFailure(ctx context.Context, actorID uuid.UUID, code, label string) {
	s.recordActionBestEffort(ctx, audit.Action{
		ActorUserID: &actorID,
		Code:        code,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultFailure,
	})
}

// logError registra un fallo best-effort con el logger de la petición (que lleva
// el request_id) o, si no hay, con el logger base del service.
func (s *service) logError(ctx context.Context, message string, err error) {
	logger := s.logger
	if requestLogger := httpserver.RequestLoggerFromContext(ctx); requestLogger != nil {
		logger = requestLogger
	}
	if logger == nil {
		logger = slog.Default()
	}
	logger.Error(message, slog.Any("error", err))
}

// --- Listado del panel (FR-018/FR-019) ---

// List devuelve el listado paginado del módulo (FR-019): **todo** —ocultas y
// marcas de los dos ejes incluidas— menos las fichas absorbidas por una unión
// (salvo `IncluirAbsorbidas`), con el **mismo orden que ve el visitante**
// (FR-006) y los filtros opcionales por visibilidad y por marca **pendiente** de
// cada eje (P9-19: `possible_duplicate` es un eje y `unavailable`/
// `missing_from_feed` el otro).
//
// Cada fila lleva la procedencia (`sources`), que es lo que el equipo necesita
// para decidir una unión, y el sobre trae `lastSync` —el registro de la última
// búsqueda— solo cuando ya se buscó alguna vez (FR-018).
//
// Los filtros fuera de su catálogo son `400 invalid` con su campo: el contrato
// solo admite los valores del registro cerrado.
func (s *service) List(ctx context.Context, filtro FiltroPanel, params paginate.Params) (MedioAdminList, error) {
	if filtro.Visibilidad != nil && !filtro.Visibilidad.Valid() {
		return MedioAdminList{}, apperr.Invalid(
			validate.MessageInvalid,
			apperr.WithDetails(map[string]any{"visibility": "solo se admite visible u oculto"}),
		)
	}
	if filtro.MarcaRevision != nil && !filtro.MarcaRevision.Valid() {
		return MedioAdminList{}, apperr.Invalid(
			validate.MessageInvalid,
			apperr.WithDetails(map[string]any{
				"reviewMark": "solo se admite possible_duplicate, unavailable o missing_from_feed",
			}),
		)
	}

	total, err := s.repository.CountMediaItems(ctx, filtro)
	if err != nil {
		return MedioAdminList{}, fmt.Errorf("contar los episodios del panel: %w", err)
	}
	episodios, err := s.repository.ListMediaItems(ctx, filtro, params)
	if err != nil {
		return MedioAdminList{}, fmt.Errorf("listar los episodios del panel: %w", err)
	}

	fuentesPorFicha, err := s.fuentesDeLaPagina(ctx, episodios)
	if err != nil {
		return MedioAdminList{}, err
	}

	items := make([]MedioAdmin, 0, len(episodios))
	for _, episodio := range episodios {
		items = append(items, medioAdminFrom(episodio, fuentesPorFicha[episodio.ID]))
	}

	out := MedioAdminList{Items: items, Total: total, Limit: params.Limit, Offset: params.Offset}
	run, found, err := s.repository.GetLastMediaSyncRun(ctx)
	if err != nil {
		return MedioAdminList{}, fmt.Errorf("leer la última búsqueda: %w", err)
	}
	if found {
		last := syncRunAdminFrom(run)
		out.LastSync = &last
	}
	return out, nil
}

// fuentesDeLaPagina lee la procedencia de **todas** las fichas de la página en
// UNA sola consulta y la agrupa por ficha (hallazgo B3): sin el lote, el listado
// hacía una consulta por fila y el panel pide el tope del contrato (100), así que
// una carga eran 100 consultas. El orden dentro de cada ficha es el de
// importación (FIFO), el mismo de `ListMediaItemSourcesByItem`.
func (s *service) fuentesDeLaPagina(ctx context.Context, episodios []Episodio) (map[uuid.UUID][]FuenteEpisodio, error) {
	if len(episodios) == 0 {
		return nil, nil
	}
	ids := make([]uuid.UUID, 0, len(episodios))
	for _, episodio := range episodios {
		ids = append(ids, episodio.ID)
	}
	fuentes, err := s.repository.ListMediaItemSourcesByItems(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("leer las fuentes de la página: %w", err)
	}
	porFicha := make(map[uuid.UUID][]FuenteEpisodio, len(episodios))
	for _, fuente := range fuentes {
		porFicha[fuente.EpisodioID] = append(porFicha[fuente.EpisodioID], fuente)
	}
	return porFicha, nil
}

// filaAdmin compone la fila del panel con su procedencia: lo usan las respuestas
// de **una** ficha (editar, unir y resolver una marca). El listado usa la
// consulta por lote (`fuentesDeLaPagina`), no esta.
func (s *service) filaAdmin(ctx context.Context, episodio Episodio) (MedioAdmin, error) {
	fuentes, err := s.repository.ListMediaItemSourcesByItem(ctx, episodio.ID)
	if err != nil {
		return MedioAdmin{}, fmt.Errorf("leer las fuentes del episodio: %w", err)
	}
	return medioAdminFrom(episodio, fuentes), nil
}

// --- Edición del panel (FR-019/FR-020) ---

// Update edita parcialmente una ficha (FR-019/FR-020): **solo** la descripción
// corta en sus dos idiomas y la visibilidad. Fusiona los campos presentes sobre
// la fila actual, revalida el resultado con las mismas reglas del dominio
// (FR-023: sin guardar nada a medias) y escribe la fila completa ya fusionada.
//
// La descripción guardada pasa a ser **del equipo** (`descriptionEdited`):
// FR-015/SC-007 garantiza que la búsqueda jamás la regenera ni la pisa. El
// cambio se publica al guardar (FR-015) y la auditoría sigue la regla exacta de
// FR-022: `media.description.update` por los datos (incluido el primer ajuste),
// `media.hide`/`media.show` **solo** por el cambio de visibilidad, y un guardado
// que cambia las dos cosas deja **dos** filas.
//
// Un `PATCH` sin cambios no es una operación: devuelve la fila actual sin
// escribir ni registrar (patrón de F7). Un id inexistente es `404`.
func (s *service) Update(ctx context.Context, actorID uuid.UUID, id uuid.UUID, patch MedioPatch) (MedioAdmin, error) {
	label := etiquetaContenido(id.String())
	fail := func(err error) (MedioAdmin, error) {
		s.recordContentFailure(ctx, actorID, codigoIntentoMedio(patch), label)
		return MedioAdmin{}, err
	}

	entrada, err := patchMedioFromInput(patch)
	if err != nil {
		return fail(err)
	}

	current, found, err := s.repository.GetMediaItemByID(ctx, id)
	if err != nil {
		return MedioAdmin{}, fmt.Errorf("leer el episodio: %w", err)
	}
	if !found {
		return fail(apperr.NotFound(messageEpisodioNoExiste))
	}
	label = etiquetaEpisodio(current)

	updated := current
	dataChanged := false
	if entrada.tocaEs {
		dataChanged = dataChanged || !mismoOpcional(current.DescripcionEs, entrada.descripcionEs)
		updated.DescripcionEs = entrada.descripcionEs
	}
	if entrada.tocaEn {
		dataChanged = dataChanged || !mismoOpcional(current.DescripcionEn, entrada.descripcionEn)
		updated.DescripcionEn = entrada.descripcionEn
	}
	visibilityChanged := false
	if entrada.visibilidad != nil {
		updated.Visibilidad = *entrada.visibilidad
		visibilityChanged = updated.Visibilidad != current.Visibilidad
	}

	if err := validarEpisodio(updated); err != nil {
		return fail(err)
	}
	if !dataChanged && !visibilityChanged {
		// Un `PATCH` idéntico no es una operación: no se escribe (`updated_at`
		// incluido) ni se registra auditoría (FR-022).
		return s.filaAdmin(ctx, current)
	}
	if dataChanged {
		// El texto guardado es del equipo: la sincronización nunca lo pisa.
		updated.DescripcionEditada = true
	}

	actions := make([]audit.Action, 0, 2)
	if dataChanged {
		actions = append(actions, contenidoAction(actorID, audit.ActionMediaDescriptionUpdate, label))
	}
	if visibilityChanged {
		actions = append(actions, contenidoAction(actorID, visibilidadCode(updated.Visibilidad), label))
	}

	saved, err := s.repository.UpdateMediaItem(ctx, updated, actions...)
	if err != nil {
		return MedioAdmin{}, fmt.Errorf("editar el episodio: %w", err)
	}
	return s.filaAdmin(ctx, saved)
}

// --- Unión por puntero y resolución de marcas (FR-013/FR-014/FR-017/FR-019) ---

// Merge une dos entradas en **una sola ficha** (FR-013/FR-014/P9-10): `{id}` es
// la ficha que se **conserva** —adopta el enlace y la procedencia que le
// faltaban— y `otherId` la que queda **absorbida**.
//
// Reglas de la unión:
//
//   - **Base del resultado**: el nombre, la carátula y la fecha son los de la
//     entrada del **video** (precisión 5 de la puerta 2/default #7), sin importar
//     cuál de las dos sea `{id}`; el título del podcast se conserva como dato de
//     origen en `sources`.
//   - **Descripción**: si solo una de las dos tiene la descripción ajustada por
//     el equipo, **manda esa** (FR-015/SC-007); si ninguna la ajustó, se conserva
//     la que exista.
//   - **Eje de duplicados**: se **despeja** en las dos (el par deja de existir y
//     su resolución también). El **eje de disponibilidad no se toca**: es
//     independiente (P9-19) y una marca abierta de la base sigue.
//   - **Cero DELETE** (FR-019): la absorbida queda con `merged_into_id` y oculta,
//     fuera de todas las consultas, y sus fuentes se **repuntan** a la resultante
//     —que queda `origin='both'`— en **una sola transacción** junto con la
//     auditoría `media.merge` (la absorbida se cita en `target_label`).
//   - La carátula propia de la absorbida, si queda sin uso, se borra del almacén
//     **en mejor esfuerzo** (`storage.Delete` es idempotente): si falla, la
//     unión no se deshace (el archivo queda huérfano, nunca publicado).
//
// Errores: `otherId` igual a `{id}` → `invalid`; una ficha inexistente o ya
// absorbida → `not_found`; dos fichas con fuente del **mismo** proveedor (no
// caben en una ficha: una entrada por sermón) → `conflict`.
func (s *service) Merge(ctx context.Context, actorID, id, otherID uuid.UUID) (MedioAdmin, error) {
	label := etiquetaContenido(id.String())
	fail := func(err error) (MedioAdmin, error) {
		s.recordContentFailure(ctx, actorID, audit.ActionMediaMerge, label)
		return MedioAdmin{}, err
	}

	if err := validarUnion(id, otherID); err != nil {
		return fail(err)
	}

	base, found, err := s.repository.GetMediaItemByID(ctx, id)
	if err != nil {
		return MedioAdmin{}, fmt.Errorf("leer el episodio: %w", err)
	}
	if !found || base.UnidoAID != nil {
		return fail(apperr.NotFound(messageEpisodioNoExiste))
	}
	absorbida, found, err := s.repository.GetMediaItemByID(ctx, otherID)
	if err != nil {
		return MedioAdmin{}, fmt.Errorf("leer el episodio que se une: %w", err)
	}
	if !found || absorbida.UnidoAID != nil {
		return fail(apperr.NotFound(messageEpisodioNoExiste))
	}

	fuentesBase, err := s.repository.ListMediaItemSourcesByItem(ctx, base.ID)
	if err != nil {
		return MedioAdmin{}, fmt.Errorf("leer las fuentes del episodio: %w", err)
	}
	fuentesAbsorbida, err := s.repository.ListMediaItemSourcesByItem(ctx, absorbida.ID)
	if err != nil {
		return MedioAdmin{}, fmt.Errorf("leer las fuentes del episodio que se une: %w", err)
	}
	if proveedorRepetido(fuentesBase, fuentesAbsorbida) {
		return fail(apperr.Conflict(
			"No se pueden unir dos episodios del mismo servicio",
			apperr.WithCause(errProveedorRepetido),
		))
	}

	resultante := fusionarEpisodios(base, absorbida, fuentesBase, fuentesAbsorbida)
	absorbida.UnidoAID = &resultante.ID
	absorbida.Visibilidad = VisibilidadOculta
	// El eje de duplicados se despeja en las dos (el par deja de existir).
	absorbida.DuplicadoDeID = nil
	absorbida.MarcaDuplicadoResueltaEn = nil

	label = etiquetaContenido(resultante.Titulo + " · " + otherID.String())
	actions := []audit.Action{contenidoAction(actorID, audit.ActionMediaMerge, label)}

	movidas, err := s.repository.RepointMediaItemSources(ctx, resultante, absorbida, actions...)
	if err != nil {
		return MedioAdmin{}, fmt.Errorf("unir los episodios: %w", err)
	}

	// Las carátulas que la unión deja sin uso (la de la absorbida y la que la
	// base tuviera si el video impuso la suya) se retiran en mejor esfuerzo: el
	// archivo deja de ser publicable por la rama de IsFilePublished.
	s.borrarCaratulasSinUso(ctx, resultante.Caratula, base.Caratula, absorbida.Caratula)

	fuentes := make([]FuenteEpisodio, 0, len(fuentesBase)+len(movidas))
	fuentes = append(fuentes, fuentesBase...)
	for _, fuente := range movidas {
		fuente.EpisodioID = resultante.ID
		fuentes = append(fuentes, fuente)
	}
	// Los enlaces no son columnas: los resuelve el `LEFT JOIN` por proveedor de
	// las consultas. Aquí se completan con las fuentes ya repuntadas, que es
	// exactamente lo que devolverá la próxima lectura.
	resultante.VideoURL = enlaceDeProveedor(fuentes, ProveedorYouTube)
	resultante.AudioURL = enlaceDeProveedor(fuentes, ProveedorPodcast)
	return medioAdminFrom(resultante, fuentes), nil
}

// ResolveMark resuelve **la marca indicada** (FR-017/US3 esc. 6, P9-19): los dos
// ejes pueden estar activos a la vez, así que el cuerpo trae `mark` y el service
// resuelve el eje que corresponde —`possible_duplicate` el de duplicados;
// `unavailable`/`missing_from_feed` el de disponibilidad—, fija la fecha de
// resolución de **ese** eje, aplica la visibilidad elegida por el equipo
// («Mostrar de nuevo» / «Mantener oculto» / «Mantener visible») y registra
// `media.mark.resolve` con la marca y la visibilidad en `target_label`.
//
// «No son el mismo sermón» (`mark='possible_duplicate'`) resuelve **las dos**
// fichas del par, que quedan visibles si el equipo así lo decide (US3 esc. 2) y
// conservan el puntero recíproco como índice de la no-reapertura (default #6).
//
// Es **idempotente**: sobre una marca ya resuelta y sin cambio de visibilidad no
// reescribe ni vuelve a registrar (repetir la operación deja el mismo estado).
// Resolver una marca que **no está activa** es `conflict` (estado inconsistente);
// un id inexistente es `not_found`. La indisponibilidad **confirmada** posterior
// manda sobre esta resolución (precisión A3 de la fase 5): eso lo aplica la
// búsqueda (service_sync.go), no esta operación.
func (s *service) ResolveMark(ctx context.Context, actorID, id uuid.UUID, mark MarcaRevision, visibility Visibilidad) (MedioAdmin, error) {
	label := etiquetaContenido(id.String())
	fail := func(err error) (MedioAdmin, error) {
		s.recordContentFailure(ctx, actorID, audit.ActionMediaMarkResolve, label)
		return MedioAdmin{}, err
	}

	if err := validarMarca(mark); err != nil {
		return fail(err)
	}
	if err := validarVisibilidad(visibility); err != nil {
		return fail(err)
	}

	current, found, err := s.repository.GetMediaItemByID(ctx, id)
	if err != nil {
		return MedioAdmin{}, fmt.Errorf("leer el episodio: %w", err)
	}
	if !found {
		return fail(apperr.NotFound(messageEpisodioNoExiste))
	}
	label = etiquetaContenido(current.Titulo + " · " + string(mark) + " · " + string(visibility))

	if mark == MarcaPosibleDuplicado {
		return s.resolverParDuplicado(ctx, actorID, current, visibility, label, fail)
	}
	return s.resolverMarcaDisponibilidad(ctx, actorID, current, mark, visibility, label, fail)
}

// resolverParDuplicado resuelve el eje de duplicados de las dos fichas del par
// (P9-19): la marca queda cerrada, la visibilidad elegida se aplica a las dos y
// el puntero recíproco se conserva. Las **dos** filas y la auditoría van en UNA
// sola transacción (M2/FR-022): o la pareja queda resuelta con su registro, o no
// cambia nada —nunca una pareja cambiada sin auditoría—. El registro cita la
// ficha que pidió la operación.
func (s *service) resolverParDuplicado(
	ctx context.Context,
	actorID uuid.UUID,
	current Episodio,
	visibility Visibilidad,
	label string,
	fail func(error) (MedioAdmin, error),
) (MedioAdmin, error) {
	if current.DuplicadoDeID == nil {
		return fail(apperr.Conflict("Ese episodio no tiene un posible duplicado que revisar"))
	}
	pareja, found, err := s.repository.GetMediaItemByID(ctx, *current.DuplicadoDeID)
	if err != nil {
		return MedioAdmin{}, fmt.Errorf("leer el par del posible duplicado: %w", err)
	}
	if !found {
		return fail(apperr.Conflict("El par del posible duplicado ya no existe"))
	}

	ahora := time.Now()
	actualizada := current
	if actualizada.MarcaDuplicadoResueltaEn == nil {
		actualizada.MarcaDuplicadoResueltaEn = &ahora
	}
	actualizada.Visibilidad = visibility
	otra := pareja
	if otra.MarcaDuplicadoResueltaEn == nil {
		otra.MarcaDuplicadoResueltaEn = &ahora
	}
	otra.Visibilidad = visibility

	if actualizada == current && otra == pareja {
		// Marca ya resuelta y misma visibilidad en las dos: repetir no es una
		// operación (no se reescribe ni se vuelve a registrar).
		return s.filaAdmin(ctx, current)
	}

	saved, err := s.repository.ResolveDuplicatePair(ctx, actualizada, otra,
		contenidoAction(actorID, audit.ActionMediaMarkResolve, label))
	if err != nil {
		return MedioAdmin{}, fmt.Errorf("resolver el par del posible duplicado: %w", err)
	}
	return s.filaAdmin(ctx, saved)
}

// resolverMarcaDisponibilidad resuelve la marca del eje de disponibilidad: fija
// su fecha de resolución (no la borra: es el índice de la no-reapertura) y aplica
// la visibilidad elegida. El resultado y la fecha de la comprobación se conservan
// como el dato que motivó la marca.
func (s *service) resolverMarcaDisponibilidad(
	ctx context.Context,
	actorID uuid.UUID,
	current Episodio,
	mark MarcaRevision,
	visibility Visibilidad,
	label string,
	fail func(error) (MedioAdmin, error),
) (MedioAdmin, error) {
	clase, ok := mark.MarcaDisponibilidad()
	if !ok || current.MarcaDisponibilidad == nil || *current.MarcaDisponibilidad != clase {
		return fail(apperr.Conflict("Ese episodio no tiene esa marca de disponibilidad"))
	}

	actualizada := current
	if actualizada.MarcaDisponibilidadResueltaEn == nil {
		ahora := time.Now()
		actualizada.MarcaDisponibilidadResueltaEn = &ahora
	}
	actualizada.Visibilidad = visibility
	if actualizada == current {
		// Marca ya resuelta y misma visibilidad: repetir no es una operación.
		return s.filaAdmin(ctx, current)
	}

	saved, err := s.repository.UpdateMediaItem(ctx, actualizada,
		contenidoAction(actorID, audit.ActionMediaMarkResolve, label))
	if err != nil {
		return MedioAdmin{}, fmt.Errorf("resolver la marca de disponibilidad: %w", err)
	}
	return s.filaAdmin(ctx, saved)
}

// borrarCaratulasSinUso retira del almacén las carátulas que la unión dejó sin
// uso (mejor esfuerzo: un fallo solo se registra en el log, la unión ya está
// aplicada y el archivo huérfano nunca se publica). Nunca toca la que quedó en
// uso: es la que sirve el sitio.
func (s *service) borrarCaratulasSinUso(ctx context.Context, enUso *string, candidatas ...*string) {
	if s.store == nil {
		return
	}
	for _, candidata := range candidatas {
		if candidata == nil {
			continue
		}
		if enUso != nil && *candidata == *enUso {
			continue
		}
		if err := s.store.Delete(ctx, *candidata); err != nil {
			s.logError(ctx, "no se pudo retirar la carátula que quedó sin uso", err)
		}
	}
}

// fusionarEpisodios arma la fila completa de la ficha resultante (P9-10): base
// del video para el nombre, la carátula y la fecha; la descripción del equipo si
// solo una la ajustó; `origin` recalculado de las fuentes unidas; y el eje de
// duplicados despejado. El eje de disponibilidad de la base se conserva tal cual.
func fusionarEpisodios(base, absorbida Episodio, fuentesBase, fuentesAbsorbida []FuenteEpisodio) Episodio {
	resultante := base

	// Precisión 5 (default #7): el nombre, la carátula y la fecha son los de la
	// entrada del video, sea cual sea la ficha que se conserva.
	if !tieneProveedor(fuentesBase, ProveedorYouTube) && tieneProveedor(fuentesAbsorbida, ProveedorYouTube) {
		resultante.Titulo = absorbida.Titulo
		resultante.Caratula = absorbida.Caratula
		resultante.PublicadoEn = absorbida.PublicadoEn
	}

	// La descripción del equipo manda (FR-015/SC-007): si solo una está
	// ajustada, gana esa; si ninguna lo está, se conserva la que exista.
	switch {
	case base.DescripcionEditada && !absorbida.DescripcionEditada:
		// La base ya es la del equipo.
	case !base.DescripcionEditada && absorbida.DescripcionEditada:
		resultante.DescripcionEs = absorbida.DescripcionEs
		resultante.DescripcionEn = absorbida.DescripcionEn
		resultante.DescripcionEditada = true
	case base.DescripcionEs == nil && base.DescripcionEn == nil:
		resultante.DescripcionEs = absorbida.DescripcionEs
		resultante.DescripcionEn = absorbida.DescripcionEn
	}

	resultante.Origen = origenDeFuentes(fuentesBase, fuentesAbsorbida)
	resultante.DuplicadoDeID = nil
	resultante.MarcaDuplicadoResueltaEn = nil
	return resultante
}

// tieneProveedor indica si una lista de fuentes trae una del proveedor indicado.
func tieneProveedor(fuentes []FuenteEpisodio, proveedor Proveedor) bool {
	for _, fuente := range fuentes {
		if fuente.Proveedor == proveedor {
			return true
		}
	}
	return false
}

// proveedorRepetido indica si las dos fichas traen fuente del mismo proveedor:
// una ficha es **una entrada por sermón** con un video y/o un audio, así que dos
// fuentes del mismo proveedor no caben y la unión es `conflict` (invariante de
// service, data-model.md).
func proveedorRepetido(fuentesBase, fuentesAbsorbida []FuenteEpisodio) bool {
	for _, fuente := range fuentesBase {
		if tieneProveedor(fuentesAbsorbida, fuente.Proveedor) {
			return true
		}
	}
	return false
}

// enlaceDeProveedor devuelve el enlace de la fuente del proveedor indicado ("" si
// la ficha no tiene esa fuente).
func enlaceDeProveedor(fuentes []FuenteEpisodio, proveedor Proveedor) string {
	for _, fuente := range fuentes {
		if fuente.Proveedor == proveedor {
			return fuente.Enlace
		}
	}
	return ""
}

// origenDeFuentes deriva los formatos de la ficha de sus fuentes (FR-003): video
// (solo YouTube), audio (solo podcast) o both.
func origenDeFuentes(fuentesBase, fuentesAbsorbida []FuenteEpisodio) OrigenMedio {
	video := tieneProveedor(fuentesBase, ProveedorYouTube) || tieneProveedor(fuentesAbsorbida, ProveedorYouTube)
	audio := tieneProveedor(fuentesBase, ProveedorPodcast) || tieneProveedor(fuentesAbsorbida, ProveedorPodcast)
	switch {
	case video && audio:
		return OrigenAmbos
	case video:
		return OrigenVideo
	case audio:
		return OrigenAudio
	default:
		// Una ficha sin fuentes no debería existir; se conserva el catálogo con
		// el valor más conservador (solo audio no publica ningún enlace).
		return OrigenAudio
	}
}

// --- Mapeo del dominio a los DTOs del panel ---

// medioAdminFrom traduce una ficha del panel a su DTO (FR-019): datos del feed
// **no editables**, la descripción corta en ambos idiomas con su marca de
// edición, la visibilidad, las marcas **por eje** (solo viajan aquí), la
// comprobación de disponibilidad, el origen y la procedencia. Las claves
// opcionales se omiten cuando no aplican (nunca `null`).
func medioAdminFrom(episodio Episodio, fuentes []FuenteEpisodio) MedioAdmin {
	dto := MedioAdmin{
		ID:                 episodio.ID.String(),
		Titulo:             episodio.Titulo,
		DescripcionEs:      episodio.DescripcionEs,
		DescripcionEn:      episodio.DescripcionEn,
		DescripcionEditada: episodio.DescripcionEditada,
		PublicadoEn:        episodio.PublicadoEn,
		Origen:             string(episodio.Origen),
		Visibilidad:        string(episodio.Visibilidad),
		Fuentes:            fuentesAdminFrom(fuentes),
		CreadoEn:           episodio.CreadoEn,
		ActualizadoEn:      episodio.ActualizadoEn,
	}
	if episodio.Caratula != nil {
		dto.CaratulaURL = rutaCaratula(*episodio.Caratula)
	}
	dto.VideoURL = episodio.VideoURL
	dto.AudioURL = episodio.AudioURL

	if episodio.DuplicadoDeID != nil {
		dto.DuplicadoDeID = episodio.DuplicadoDeID.String()
	}
	dto.MarcaDuplicadoResueltaEn = episodio.MarcaDuplicadoResueltaEn
	if episodio.UnidoAID != nil {
		dto.UnidoAID = episodio.UnidoAID.String()
	}
	dto.DisponibilidadComprobadaEn = episodio.DisponibilidadComprobadaEn
	if episodio.ResultadoDisponibilidad != nil {
		dto.ResultadoDisponibilidad = string(*episodio.ResultadoDisponibilidad)
	}
	if episodio.MarcaDisponibilidad != nil {
		dto.MarcaDisponibilidad = string(*episodio.MarcaDisponibilidad)
	}
	dto.MarcaDisponibilidadResueltaEn = episodio.MarcaDisponibilidadResueltaEn
	return dto
}

// fuentesAdminFrom traduce la procedencia de una ficha: siempre un arreglo (el
// contrato exige al menos una fuente; una ficha sin fuentes no existe).
func fuentesAdminFrom(fuentes []FuenteEpisodio) []FuenteAdmin {
	out := make([]FuenteAdmin, 0, len(fuentes))
	for _, fuente := range fuentes {
		out = append(out, FuenteAdmin{
			Proveedor:  string(fuente.Proveedor),
			IDExterno:  fuente.IDExterno,
			TituloFeed: fuente.TituloFeed,
			Enlace:     fuente.Enlace,
		})
	}
	return out
}

// syncRunAdminFrom traduce el registro de una búsqueda al DTO del panel (FR-018).
func syncRunAdminFrom(run SyncRun) SyncRunAdmin {
	dto := SyncRunAdmin{
		ID:            run.ID.String(),
		IniciadoEn:    run.IniciadoEn,
		TerminadoEn:   run.TerminadoEn,
		Importados:    run.Importados,
		Unidos:        run.Unidos,
		Marcados:      run.Marcados,
		Ocultos:       run.Ocultos,
		Comprobados:   run.Comprobados,
		Inconclusos:   run.Inconclusos,
		EstadoYouTube: string(run.EstadoYouTube),
		EstadoPodcast: string(run.EstadoPodcast),
	}
	if run.ActorUserID != nil {
		dto.ActorUserID = run.ActorUserID.String()
	}
	return dto
}

// mismoOpcional compara dos opcionales ya normalizados (nil = sin dato).
func mismoOpcional(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
