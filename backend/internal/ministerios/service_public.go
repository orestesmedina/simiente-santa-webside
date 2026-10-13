package ministerios

import (
	"context"
	"fmt"
)

// service_public.go implementa la cara pública de los ministerios (T613/T614):
// la sección completa `GET /api/v1/ministerios` y el bloque de la portada
// `GET /api/v1/portada/ministerios`. Resuelve el idioma en el servidor con
// fallback `en → es` por campo (FR-008), lee SOLO contenido publicado
// (FR-012/SC-002) y —aquí, en un solo lugar— compone la ficha del encargado
// **solo si su interruptor está publicado** (FR-026/SC-014, P6-2/RG6-1): el
// cliente no reimplementa ni el fallback ni la privacidad.
//
// El DTO público no tiene campos para `publicationState` ni para el encargado
// fuera de `leader`: con la ficha oculta, el dato no tiene por dónde salir ni
// aunque una capa superior se descuidara.

// MinisteriosPublicos es la sección pública completa (contrato
// `MinisteriosPublicos`, FR-002): `items` está presente **solo** si hay
// ministerios publicados (clave omitida con 0, nunca `null`, SC-010: la
// navegación oculta la entrada y la ruta SPA `/ministerios` redirige a la
// portada —decisión del frontend, P6-21—).
type MinisteriosPublicos struct {
	Lang  string              `json:"lang"`
	Items []MinisterioPublico `json:"items,omitempty"`
}

// GetMinisterios devuelve la sección pública resuelta a `lang` (FR-002/FR-005):
// los ministerios publicados, en el orden manual del módulo (`sortOrder`, `id`,
// FR-004) y acotados por `maxPublicItems` (tope de seguridad del servidor,
// RG6-12). Un idioma fuera de `es|en` es `400 invalid` con `details.lang`; no hay
// ninguna vía pública a borradores ni a fichas de encargado no publicadas.
func (s *service) GetMinisterios(ctx context.Context, lang string) (MinisteriosPublicos, error) {
	lang, err := normalizeLang(lang)
	if err != nil {
		return MinisteriosPublicos{}, err
	}

	ministries, err := s.repository.ListPublishedMinistries(ctx, maxPublicItems)
	if err != nil {
		return MinisteriosPublicos{}, fmt.Errorf("listar los ministerios publicados: %w", err)
	}

	out := MinisteriosPublicos{Lang: lang}
	if len(ministries) > 0 {
		out.Items = make([]MinisterioPublico, 0, len(ministries))
		for _, ministry := range ministries {
			out.Items = append(out.Items, localizeMinistry(ministry, lang))
		}
	}
	return out, nil
}

// localizeMinistry traduce un ministerio publicado al DTO del visitante.
//
// Es el **único** lugar donde se construye la ficha `leader` (P6-2/RG6-1): solo
// cuando el interruptor está publicado **y** hay nombre del encargado. Con el
// interruptor apagado el DTO no lleva ni el campo ni la cadena del nombre, del
// celular o del correo, aunque el repositorio haya entregado la fila completa.
func localizeMinistry(ministry Ministry, lang string) MinisterioPublico {
	public := MinisterioPublico{
		ID:          ministry.ID.String(),
		Name:        pickString(ministry.NameEs, ministry.NameEn, lang),
		Description: pickString(ministry.DescriptionEs, ministry.DescriptionEn, lang),
		Schedule:    pickOptional(ministry.ScheduleEs, ministry.ScheduleEn, lang),
		Place:       pickOptional(ministry.PlaceEs, ministry.PlaceEn, lang),
		WhatsappURL: ministry.WhatsappURL,
		Image:       publicImage(ministry.ImageFile, ministry.ImageAltEs, ministry.ImageAltEn, lang),
	}

	if ministry.LeaderVisible && ministry.LeaderNameEs != nil {
		leader := LiderPublico{Name: pickString(*ministry.LeaderNameEs, ministry.LeaderNameEn, lang)}
		// Un solo medio de contacto, nunca los dos (Q8; la validación y la BD lo
		// garantizan). Un encargado solo con nombre deja la ficha sin contacto.
		switch {
		case ministry.LeaderPhone != nil:
			leader.Phone = ministry.LeaderPhone
		case ministry.LeaderEmail != nil:
			leader.Email = ministry.LeaderEmail
		}
		public.Leader = &leader
	}
	return public
}

// PortadaMinisterios es el bloque de ministerios de la portada (contrato
// `PortadaMinisterios`, US6/FR-027): hasta 5 ítems publicados —tope **firme del
// servidor**, sin parámetro de límite— y `items` omitida sin publicados
// (FR-028: la portada no dibuja el bloque **ni muestra mensaje**, a diferencia
// del bloque «Lo que viene» de F4).
type PortadaMinisterios struct {
	Lang  string                  `json:"lang"`
	Items []ItemMinisterioPortada `json:"items,omitempty"`
}

// ItemMinisterioPortada es la tarjeta compacta del bloque (FR-027 y `ux.md`
// §4.9): imagen + nombre, con el horario y el lugar cuando existen. **No tiene
// campos de encargado ni botón de unión** (P6-2/D-12: el detalle vive en la
// sección): la privacidad de FR-026 es estructural en esta ruta —con el
// interruptor apagado no hay nada que filtrar—.
type ItemMinisterioPortada struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Schedule *string        `json:"schedule,omitempty"`
	Place    *string        `json:"place,omitempty"`
	Image    *ImagenPublica `json:"image,omitempty"`
}

// GetPortadaMinisterios devuelve el extracto del bloque de la portada resuelto a
// `lang` (FR-027/FR-028): los mismos ministerios publicados que la sección —mismo
// predicado y mismo fallback `en → es` por campo— acotados al tope `maxPortadaItems`
// (5) en el orden manual del módulo (`sortOrder`, `id`). Sin publicados, `items`
// queda omitida y la portada no renderiza el bloque; el tope se aplica en el
// servidor, así que la respuesta no admite ningún parámetro de límite.
func (s *service) GetPortadaMinisterios(ctx context.Context, lang string) (PortadaMinisterios, error) {
	lang, err := normalizeLang(lang)
	if err != nil {
		return PortadaMinisterios{}, err
	}

	ministries, err := s.repository.ListPublishedMinistries(ctx, maxPortadaItems)
	if err != nil {
		return PortadaMinisterios{}, fmt.Errorf("listar los ministerios publicados: %w", err)
	}

	out := PortadaMinisterios{Lang: lang}
	if len(ministries) > 0 {
		out.Items = make([]ItemMinisterioPortada, 0, len(ministries))
		for _, ministry := range ministries {
			out.Items = append(out.Items, localizePortadaItem(ministry, lang))
		}
	}
	return out, nil
}

// localizePortadaItem traduce un ministerio publicado a la tarjeta del bloque:
// solo lo que la tarjeta muestra (la descripción, el encargado y el enlace se
// quedan en la sección).
func localizePortadaItem(ministry Ministry, lang string) ItemMinisterioPortada {
	return ItemMinisterioPortada{
		ID:       ministry.ID.String(),
		Name:     pickString(ministry.NameEs, ministry.NameEn, lang),
		Schedule: pickOptional(ministry.ScheduleEs, ministry.ScheduleEn, lang),
		Place:    pickOptional(ministry.PlaceEs, ministry.PlaceEn, lang),
		Image:    publicImage(ministry.ImageFile, ministry.ImageAltEs, ministry.ImageAltEn, lang),
	}
}

// --- Localización (fallback en → es por campo, FR-008) ---

// pickString resuelve un campo obligatorio en español: si se pide inglés y hay
// traducción, la usa; si no, devuelve el español (nunca vacío, SC-006).
func pickString(es string, en *string, lang string) string {
	if lang == LangEN && en != nil {
		return *en
	}
	return es
}

// pickOptional resuelve un campo opcional: prefiere la traducción pedida y cae
// al español; si tampoco hay español, devuelve nil (la clave se omite).
func pickOptional(es, en *string, lang string) *string {
	if lang == LangEN && en != nil {
		return en
	}
	return es
}

// publicImage compone la imagen pública (FR-020) con la URL relativa que sirve
// `GET /api/v1/media/{fileName}` y su texto alternativo ya localizado. Devuelve
// nil cuando el ministerio no tiene imagen; si la fila trajera un archivo sin
// texto alternativo (imposible por validación y por `CHECK`), tampoco se publica
// la imagen (fail-closed: sin huecos ni `alt` vacíos, SC-017).
func publicImage(file, altEs, altEn *string, lang string) *ImagenPublica {
	if file == nil || altEs == nil {
		return nil
	}
	return &ImagenPublica{URL: MediaPathPrefix + *file, Alt: pickString(*altEs, altEn, lang)}
}
