package medios

import (
	"context"
	"fmt"
	"log/slog"

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

	items := make([]MedioAdmin, 0, len(episodios))
	for _, episodio := range episodios {
		fila, err := s.filaAdmin(ctx, episodio)
		if err != nil {
			return MedioAdminList{}, err
		}
		items = append(items, fila)
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

// filaAdmin compone la fila del panel con su procedencia. La procedencia se lee
// por ficha (una consulta por fila, `ListMediaItemSourcesByItem`): una ficha
// tiene 1 o 2 fuentes y el listado va paginado (20 por defecto), así que el
// coste es acotado y constante por página.
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
		s.recordContentFailure(ctx, actorID, audit.ActionMediaDescriptionUpdate, label)
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
