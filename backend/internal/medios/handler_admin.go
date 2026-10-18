package medios

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/httpserver"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/session"
	"simiente-santa/backend/internal/platform/validate"
)

// handler_admin.go publica la gestión del módulo desde el panel bajo
// /api/v1/admin/medios (T925–T927). Es solo HTTP: decodifica y valida los DTOs
// del contrato, delega en el service y responde el DTO del panel —que sí lleva la
// visibilidad, las marcas de los dos ejes y la procedencia—. Nada de SQL, de
// credenciales ni de texto técnico de red en las respuestas (FR-018).
//
// La cadena (authn → guard → authz(`medios`) → CSRF) la monta routes.go (T928):
// los handlers **no** comprueban permisos por su cuenta; el 401 de aquí es solo la
// red de seguridad de composición.
//
// Aquí viven también los helpers HTTP compartidos del panel (patrón de
// F2/F3/F4/F6/F7): decodificar y validar el cuerpo, resolver la identidad y los
// preludios de las operaciones.

// messageUnauthenticated es el 401 genérico del dominio. Coincide a propósito con
// el de la cadena de sesión (middleware.authn).
const messageUnauthenticated = "Necesitas iniciar sesión para continuar"

// maxRequestBytes acota el cuerpo JSON de una petición (1 MiB): evita leer
// entradas ilimitadas antes de decodificarlas (CWE-400).
const maxRequestBytes = 1 << 20

// --- Listado del panel (T925, FR-018/FR-019) ---

// ListMedios responde GET /api/v1/admin/medios (FR-019): listado paginado del
// módulo con los filtros del contrato —`visibility`, `reviewMark` (la marca de
// esa clase **pendiente**) e `includeMerged` (por defecto las absorbidas no se
// listan)—, el mismo orden que ve el visitante y `lastSync` cuando ya hubo alguna
// búsqueda (FR-018).
//
// El catálogo cerrado de `visibility`/`reviewMark` lo valida el service (es su
// regla, T918): un valor fuera del catálogo es `400 invalid` con el campo en
// `details`. Aquí solo se traducen los query params a los tipos del dominio y se
// comprueba que `includeMerged` sea booleano.
//
// Códigos: 200 con MedioAdminList; parámetros inválidos → 400 con details; sin
// sesión → 401.
func (h *Handler) ListMedios(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, ok := h.adminIdentity(ctx, w); !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}

	filtro, params, err := h.listParams(r)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}

	list, err := h.service.List(ctx, filtro, params)
	if err != nil {
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, list)
}

// --- Edición de la ficha (T925, FR-019/FR-020) ---

// UpdateMedio responde PATCH /api/v1/admin/medios/{id} (FR-019/FR-020): la única
// edición que F9 permite —la descripción corta en sus dos idiomas y la
// visibilidad—. El handler transmite el DTO **parcial** tal cual y el service
// fusiona los campos presentes sobre la fila actual: los ausentes no cambian.
// Cualquier otra clave del cuerpo se rechaza con `400 invalid` (el DTO es
// `additionalProperties: false`) y el campo se nombra en `details`.
//
// Códigos: 200 con MedioAdmin; DTO inválido → 400 con details; id no UUID → 400
// con details.id; inexistente → 404.
func (h *Handler) UpdateMedio(w http.ResponseWriter, r *http.Request) {
	id, err := medioIDFromRequest(r)
	if err != nil {
		httpserver.WriteError(r.Context(), w, h.logger, err)
		return
	}
	h.updateItem(w, r, func(ctx context.Context, actorID uuid.UUID) (any, error) {
		var patch MedioPatch
		if !h.decodeAndValidate(w, r, &patch) {
			return nil, errAlreadyHandled
		}
		return h.service.Update(ctx, actorID, id, patch)
	})
}

// --- Helpers HTTP compartidos del panel (patrón de F2/F3/F4/F6/F7) ---

// decodeAndValidate decodifica el cuerpo JSON de la petición en dto y lo valida
// con platform/validate (CWE-20). Devuelve true si el DTO es utilizable; si no, ya
// escribió el 400 invalid (con `details` por campo cuando la validación lo
// produce) y devuelve false. Rechaza campos desconocidos
// (additionalProperties: false del contrato) y más de un valor JSON.
//
// Todo rechazo pasa por recordRejected: deja la traza y el registro best-effort en
// admin_actions con `result='failure'` (R7-10).
func (h *Handler) decodeAndValidate(w http.ResponseWriter, r *http.Request, dto any) bool {
	ctx := r.Context()
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dto); err != nil {
		h.rejectInvalid(ctx, w, r, err)
		return false
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		h.rejectInvalid(ctx, w, r, errors.New("el cuerpo debe contener un único objeto JSON"))
		return false
	}

	if err := validate.Struct(dto); err != nil {
		h.recordRejected(ctx, r)
		httpserver.WriteError(ctx, w, h.logger, err)
		return false
	}
	return true
}

// rejectInvalid responde el 400 del contrato por un cuerpo que no se pudo
// decodificar o que traía basura tras el objeto JSON. Cuando el decoder señala un
// campo concreto —una clave desconocida (el `PATCH` es `additionalProperties:
// false`) o un tipo equivocado— el campo viaja en `details` para que el panel lo
// marque (el contrato: «`details` indica qué campo corregir»); el texto del error
// de Go nunca sale en la respuesta.
func (h *Handler) rejectInvalid(ctx context.Context, w http.ResponseWriter, r *http.Request, cause error) {
	h.recordRejected(ctx, r)
	httpserver.WriteError(ctx, w, h.logger, apperr.Invalid(validate.MessageInvalid, opcionesRechazo(cause)...))
}

// opcionesRechazo traduce el error del decoder a las opciones del 400: el campo
// desconocido o el de tipo equivocado, con un motivo seguro.
func opcionesRechazo(cause error) []apperr.Option {
	if campo := campoDesconocido(cause); campo != "" {
		return []apperr.Option{apperr.WithDetails(map[string]any{
			campo: "Este campo no forma parte del formulario.",
		})}
	}
	var tipo *json.UnmarshalTypeError
	if errors.As(cause, &tipo) && tipo.Field != "" {
		return []apperr.Option{apperr.WithDetails(map[string]any{
			tipo.Field: "El tipo del valor no es válido.",
		})}
	}
	return nil
}

// campoDesconocido extrae el nombre del campo de un error de clave desconocida
// del decoder (`json: unknown field "title"`); "" si el error es de otra clase.
func campoDesconocido(err error) string {
	const prefijo = "json: unknown field "
	msg := err.Error()
	if !strings.HasPrefix(msg, prefijo) {
		return ""
	}
	return strings.Trim(strings.TrimPrefix(msg, prefijo), `"`)
}

// recordRejected deja constancia de una petición rechazada por JSON o DTO
// inválido. Es best-effort y nunca cambia la respuesta; usa el logger por petición
// (con request_id) cuando está en el contexto y no transporta el cuerpo de la
// petición (regla de F2).
func (h *Handler) recordRejected(ctx context.Context, r *http.Request) {
	logger := httpserver.RequestLoggerFromContext(ctx)
	if logger == nil {
		logger = h.logger
	}
	if logger == nil {
		logger = slog.Default()
	}
	logger.Warn("petición rechazada: cuerpo o datos inválidos",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)

	if h.service == nil {
		return
	}
	h.service.RecordRejectedBestEffort(ctx, Rejection{
		ActorUserID: identityUserID(ctx),
		Method:      r.Method,
		Path:        r.URL.Path,
	})
}

// adminIdentity recupera la identidad resuelta por authn. Las rutas del panel van
// dentro de la cadena (authn → …), así que siempre debería estar; sin ella
// responde 401 como red de seguridad.
func (h *Handler) adminIdentity(ctx context.Context, w http.ResponseWriter) (session.Identity, bool) {
	identity, ok := session.IdentityFromContext(ctx)
	if !ok {
		httpserver.WriteError(ctx, w, h.logger, apperr.Unauthenticated(messageUnauthenticated))
		return session.Identity{}, false
	}
	return identity, true
}

// identityUserID devuelve el id de la identidad autenticada, o nil en las rutas
// públicas.
func identityUserID(ctx context.Context) *uuid.UUID {
	identity, ok := session.IdentityFromContext(ctx)
	if !ok {
		return nil
	}
	id := identity.UserID
	return &id
}

// errAlreadyHandled señala que el helper ya escribió la respuesta (p. ej. el 400
// de un DTO inválido): el envoltorio de las operaciones no debe volver a
// responder.
var errAlreadyHandled = errors.New("respuesta ya escrita")

// updateItem unifica el preludio de las operaciones del panel que responden 200:
// exige sesión y servicio, delega en fn y responde el DTO del panel.
func (h *Handler) updateItem(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, actorID uuid.UUID) (any, error)) {
	ctx := r.Context()
	identity, ok := h.adminIdentity(ctx, w)
	if !ok {
		return
	}
	if !h.serviceConfigured(ctx, w) {
		return
	}
	item, err := fn(ctx, identity.UserID)
	if err != nil {
		if errors.Is(err, errAlreadyHandled) {
			return
		}
		httpserver.WriteError(ctx, w, h.logger, err)
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, item)
}

// medioIDFromRequest lee el id de la ruta ({id}) y lo convierte a UUID. Un valor
// que no es UUID → 400 invalid con details.id.
func medioIDFromRequest(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return uuid.Nil, apperr.Invalid(
			"El identificador del episodio no es válido",
			apperr.WithDetails(map[string]any{"id": "Debe ser un identificador UUID válido"}),
		)
	}
	return id, nil
}

// listParams lee la paginación del panel (`limit`/`offset` vía platform/paginate,
// §8.1.3) y los filtros del contrato (FR-019): `visibility`, `reviewMark` e
// `includeMerged`. Un `includeMerged` que no sea booleano es `400 invalid` con el
// campo en `details`; el catálogo cerrado de `visibility`/`reviewMark` lo valida
// el service (T918), que es su dueño.
func (h *Handler) listParams(r *http.Request) (FiltroPanel, paginate.Params, error) {
	params, err := paginate.FromValues(r.URL.Query())
	if err != nil {
		return FiltroPanel{}, paginate.Params{}, err
	}

	var filtro FiltroPanel
	if raw := strings.TrimSpace(r.URL.Query().Get("visibility")); raw != "" {
		visibilidad := Visibilidad(raw)
		filtro.Visibilidad = &visibilidad
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("reviewMark")); raw != "" {
		marca := MarcaRevision(raw)
		filtro.MarcaRevision = &marca
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("includeMerged")); raw != "" {
		incluir, err := strconv.ParseBool(raw)
		if err != nil {
			return FiltroPanel{}, paginate.Params{}, apperr.Invalid(
				"El filtro de absorbidas no es válido",
				apperr.WithDetails(map[string]any{"includeMerged": "debe ser true o false"}),
			)
		}
		filtro.IncluirAbsorbidas = incluir
	}
	return filtro, params, nil
}
