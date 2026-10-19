package medios

import "context"

// fake_feed_source_test.go declara el doble de `FeedSource` que usan las
// pruebas del dominio (patrón `fake_repository_test.go` de F4/F6/F7): devuelve
// entradas o error por origen, sin tocar la red.
//
// Es la pieza con la que T920/T921 prueban la importación, la unión, las marcas
// y el resumen de FR-016/FR-017 sin depender de los feeds reales.

// fakeFeedSource es el doble programable de los dos orígenes.
type fakeFeedSource struct {
	youtube    ResultadoFeed
	podcast    ResultadoFeed
	errYoutube error
	errPodcast error

	llamadasYouTube int
	llamadasPodcast int
}

func (f *fakeFeedSource) FetchYouTube(context.Context) (ResultadoFeed, error) {
	f.llamadasYouTube++
	return f.youtube, f.errYoutube
}

func (f *fakeFeedSource) FetchPodcast(context.Context) (ResultadoFeed, error) {
	f.llamadasPodcast++
	return f.podcast, f.errPodcast
}

// El doble implementa la interfaz del dominio: si cambia la firma, la prueba no
// compila.
var _ FeedSource = (*fakeFeedSource)(nil)
