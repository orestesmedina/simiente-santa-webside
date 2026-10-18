package medios

import (
	"fmt"

	"simiente-santa/backend/internal/platform/storage"
)

// red_saliente.go reúne la **fábrica de la red saliente** del dominio (T928): las
// tres implementaciones reales —el lector de los dos feeds (T908), el probador de
// disponibilidad (T909) y el descargador de carátulas (T910)— construidas con la
// fábrica única de cliente HTTP (T907A/P9-20). Es la pieza que `cmd/api` inyecta
// en el service: sus constructores viven en los archivos de cada consumidor
// (donde están sus allowlists, topes y timeouts) y esta es la única puerta
// exportada, para que el cableado no pueda montar un consumidor sin sus barreras.
//
// Los tres consumidores salen **solo** por esta vía: el dominio es el único lugar
// del backend con red saliente y sus barreras (https + allowlist, sin IPs
// privadas, redirects validados y cuerpo acotado) se aplican por construcción.

// RedSaliente agrupa los tres consumidores reales de la red saliente del dominio
// (P9-20). Cada campo es la interfaz que consume el service, de modo que las
// pruebas pueden sustituir cualquiera de ellos por un doble.
type RedSaliente struct {
	// Feeds lee los dos feeds del cliente (FR-010).
	Feeds FeedSource
	// Probador comprueba si el enlace de un episodio sigue disponible (FR-017).
	Probador AvailabilityProber
	// Caratulas copia al sitio la carátula de un elemento importado (P9-7).
	Caratulas CoverFetcher
}

// NewRedSaliente construye los tres consumidores reales de la red saliente con la
// fábrica única de cliente HTTP (T907A). `store` es el almacén de archivos del
// sitio (el mismo de F3/F4/F6: las carátulas se sirven como
// `/api/v1/media/<file>`) y `maxBytes` es `UPLOAD_MAX_BYTES`, el tope que aplica
// tanto el almacén como la descarga de carátulas (P9-7/R9-8).
//
// Un error aquí es de **composición** (una allowlist o un tope imposibles): el
// arranque debe fallar en vez de dejar el módulo sin red o sin barreras.
func NewRedSaliente(store storage.Store, maxBytes int64) (RedSaliente, error) {
	feeds, err := nuevoLectorFeeds()
	if err != nil {
		return RedSaliente{}, fmt.Errorf("medios: construir el lector de feeds: %w", err)
	}
	probador, err := nuevoProbadorDisponibilidad()
	if err != nil {
		return RedSaliente{}, fmt.Errorf("medios: construir el probador de disponibilidad: %w", err)
	}
	caratulas, err := nuevoDescargadorCaratulas(store, maxBytes)
	if err != nil {
		return RedSaliente{}, fmt.Errorf("medios: construir el descargador de carátulas: %w", err)
	}
	return RedSaliente{Feeds: feeds, Probador: probador, Caratulas: caratulas}, nil
}
