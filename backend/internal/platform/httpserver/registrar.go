// Package httpserver concentra tres responsabilidades transversales de la capa
// HTTP: la interfaz neutral Registrar (los dominios no conocen el router), el
// formato uniforme de respuestas (WriteJSON / WriteError) y el servidor con
// timeouts y apagado ordenado.
//
// Es plumbing de infraestructura: no conoce ningún dominio (arq. R1). F1 lo
// implementa sobre net/http (patrones "GET /ruta" de Go 1.22+, D-A4); adoptar
// otro router después solo cambia este paquete.
package httpserver

import (
	"net/http"
	"strings"
)

// Middleware envuelve un handler HTTP. Es el alias del tipo de la stdlib, así
// que cualquier middleware compatible con net/http entra sin adaptación.
type Middleware = func(http.Handler) http.Handler

// Registrar es la única puerta que tienen los dominios para publicar rutas. No
// expone el router concreto. Es la excepción declarada a "la interfaz la define
// quien la consume" (arq. §1.3): su consumidor son todos los routes.go y su
// neutralidad justifica definirla una sola vez.
type Registrar interface {
	// Handle publica un handler en method + path (patrón estilo Go:
	// "GET /contactos/{id}"). El método va aparte para obligar a cada ruta a
	// declararlo.
	Handle(method, path string, h http.HandlerFunc)
	// Group devuelve un Registrar con el prefijo y los middlewares acumulados.
	// El primer middleware de la lista queda más externo.
	Group(prefix string, mws ...Middleware) Registrar
}

// muxRegistrar adapta un *http.ServeMux estándar a la interfaz Registrar.
type muxRegistrar struct {
	mux  *http.ServeMux
	base string       // prefijo del grupo ("" en la raíz)
	mws  []Middleware // middlewares acumulados del grupo
}

// NewMuxRegistrar adapta un http.ServeMux estándar a la interfaz Registrar.
func NewMuxRegistrar(mux *http.ServeMux) Registrar {
	return &muxRegistrar{mux: mux}
}

// Handle publica el handler en el patrón method + path del grupo actual,
// envuelto por los middlewares acumulados (el primero de la lista, el más
// externo).
func (r *muxRegistrar) Handle(method, path string, h http.HandlerFunc) {
	full := joinPath(r.base, path)
	var handler http.Handler = h
	for i := len(r.mws) - 1; i >= 0; i-- { // el primer middleware de la lista queda más fuera
		handler = r.mws[i](handler)
	}
	r.mux.Handle(method+" "+full, handler) // patrón de Go 1.22+
}

// Group hereda el prefijo y los middlewares del grupo actual y añade los
// nuevos. El orden se conserva: los heredados primero (más externos).
func (r *muxRegistrar) Group(prefix string, mws ...Middleware) Registrar {
	acumulados := make([]Middleware, 0, len(r.mws)+len(mws))
	acumulados = append(acumulados, r.mws...)
	acumulados = append(acumulados, mws...)
	return &muxRegistrar{mux: r.mux, base: joinPath(r.base, prefix), mws: acumulados}
}

// joinPath une el prefijo de un grupo con la ruta de una operación sin duplicar
// ni perder la barra separadora.
func joinPath(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	default:
		return strings.TrimSuffix(a, "/") + "/" + strings.TrimPrefix(b, "/")
	}
}
