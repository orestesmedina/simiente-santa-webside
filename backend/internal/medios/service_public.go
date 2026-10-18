package medios

import (
	"context"
	"fmt"
	"time"
)

// service_public.go implementa la cara pública del dominio (T916): la sección
// `GET /api/v1/medios` —lista simple con «Cargar más»—.
//
// Solo salen fichas **publicables** (FR-009/SC-002): el predicado vive en la
// consulta (repository_episodios.go), de modo que una ficha oculta, absorbida o
// con un «posible duplicado» abierto no llega siquiera a este archivo. Aquí se
// resuelve el idioma con fallback `en → es` **por campo** (FR-025) y se compone
// el DTO público —`MedioPublico`, el único lugar donde se arma—: el cliente no
// reimplementa ni el fallback ni la regla de omitidos.
//
// El DTO público **no tiene campos** para `visibility`, marcas de ningún eje,
// `mergedIntoId` ni `sources`: con esos datos fuera del contrato, una ficha no
// publicable no tiene por dónde salir ni aunque una capa superior se descuidara.

// MediosPublicos es la tanda de la sección pública (FR-002/FR-006). `items` es
// siempre un arreglo (vacío cuando no hay nada publicable, nunca `null`): la
// redirección de `/medios` la decide el frontend (P9-16).
type MediosPublicos struct {
	Lang   string         `json:"lang"`
	Items  []MedioPublico `json:"items"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// MedioPublico es la ficha del visitante (FR-001…FR-005): el nombre del feed
// (FR-003), la descripción corta ya localizada, la fecha de publicación, la
// carátula **servida por el sitio** (SC-005) y los enlaces externos que la
// importación unió. Las claves opcionales se **omiten** cuando no aplican
// (SC-010: sin líneas vacías ni marcadores) y nunca son `null`.
type MedioPublico struct {
	ID          string    `json:"id"`
	Titulo      string    `json:"title"`
	Descripcion string    `json:"description,omitempty"`
	PublicadoEn time.Time `json:"publishedAt"`
	CaratulaURL string    `json:"coverUrl,omitempty"`
	VideoURL    string    `json:"videoUrl,omitempty"`
	AudioURL    string    `json:"audioUrl,omitempty"`
}

// GetMedios devuelve una tanda de la sección pública resuelta a `lang`
// (FR-002…FR-006): las fichas publicables en el orden de la sección
// (`published_at DESC, created_at ASC, id`), la tanda por `limit`/`offset`
// (default 12, tope 50: R9-10) y el sobre `{lang, items, total, limit, offset}`
// que sostiene «Cargar más».
//
// Un idioma fuera de `es|en` es `400 invalid` con `details.lang` y unos límites
// imposibles son `400 invalid` con su campo (T915). El total lo da el conteo con
// el mismo predicado público, de modo que el sobre de paginación nunca miente.
func (s *service) GetMedios(ctx context.Context, lang string, limit, offset int) (MediosPublicos, error) {
	lang, err := normalizeLang(lang)
	if err != nil {
		return MediosPublicos{}, err
	}
	params, err := normalizarPaginacionPublica(limit, offset)
	if err != nil {
		return MediosPublicos{}, err
	}

	total, err := s.repository.CountPublishedMediaItems(ctx)
	if err != nil {
		return MediosPublicos{}, fmt.Errorf("contar los episodios publicables: %w", err)
	}
	episodios, err := s.repository.ListPublishedMediaItems(ctx, params)
	if err != nil {
		return MediosPublicos{}, fmt.Errorf("listar los episodios publicables: %w", err)
	}

	out := MediosPublicos{
		Lang:   lang,
		Items:  make([]MedioPublico, 0, len(episodios)),
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}
	for _, episodio := range episodios {
		out.Items = append(out.Items, medioPublico(episodio, lang))
	}
	return out, nil
}

// medioPublico compone la ficha del visitante (FR-003): es el ÚNICO lugar donde
// se arma el DTO público, así que la sección y el bloque no pueden divergir.
// `title`, `publishedAt`, `coverUrl`, `videoUrl` y `audioUrl` son un único valor
// idéntico en ambos idiomas (FR-026); solo la descripción depende del idioma.
func medioPublico(episodio Episodio, lang string) MedioPublico {
	publico := MedioPublico{
		ID:          episodio.ID.String(),
		Titulo:      episodio.Titulo,
		PublicadoEn: episodio.PublicadoEn,
	}
	if descripcion := descripcionLocalizada(episodio, lang); descripcion != "" {
		publico.Descripcion = descripcion
	}
	if episodio.Caratula != nil {
		publico.CaratulaURL = rutaCaratula(*episodio.Caratula)
	}
	publico.VideoURL = episodio.VideoURL
	publico.AudioURL = episodio.AudioURL
	return publico
}

// descripcionLocalizada resuelve la descripción corta con el fallback `en → es`
// en el servidor (FR-025/Decisión 8): en inglés solo si el equipo la tradujo; si
// no, la española —nunca un campo vacío—. Sin ninguna de las dos devuelve `""`
// y la ficha **omite la línea** (precisión 3 de la puerta 2).
func descripcionLocalizada(episodio Episodio, lang string) string {
	if lang == LangEN && episodio.DescripcionEn != nil {
		return *episodio.DescripcionEn
	}
	if episodio.DescripcionEs != nil {
		return *episodio.DescripcionEs
	}
	return ""
}

// rutaCaratula compone la ruta relativa con la que se sirve la carátula copiada
// al sitio (SC-005: el navegador del visitante no pide nada a YouTube ni a
// Spotify). El nombre lo generó el servidor al importar (`img_<uuid>.<ext>`).
func rutaCaratula(archivo string) string {
	return "/api/v1/media/" + archivo
}
