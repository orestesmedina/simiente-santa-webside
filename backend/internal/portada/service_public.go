package portada

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/storage"
)

// service_public.go implementa la cara pública de la portada (T318): resuelve el
// idioma en el servidor con fallback `en → es` por campo (R3-2/FR-009), lee SOLO
// las variantes publicadas (FR-013/SC-002), omite por completo las secciones sin
// elementos publicados (SC-012) y arma los enlaces listos para abrir (SC-009).
// Además expone la política de descarga de imágenes (analyze C4/M6).
//
// El cliente NO reimplementa el fallback (analyze I2): los DTOs de salida ya
// traen los valores resueltos.

// --- DTOs de salida pública (contrato `PortadaPublica`) ---

// PublicPortada es la información general publicada, ya resuelta a `lang`. Las
// claves de sección solo están presentes si tienen elementos publicados
// (Q10/SC-012).
type PublicPortada struct {
	Lang     string                  `json:"lang"`
	Identity *PublicIdentity         `json:"identity,omitempty"`
	About    *PublicAbout            `json:"about,omitempty"`
	Schedule []PublicScheduleItem    `json:"schedule,omitempty"`
	Whatsapp []PublicWhatsappChannel `json:"whatsapp,omitempty"`
	Socials  []PublicSocialLink      `json:"socials,omitempty"`
	Contact  *PublicContact          `json:"contact,omitempty"`
}

// PublicIdentity es la identidad publicada, ya localizada (FR-002).
type PublicIdentity struct {
	Name          string  `json:"name"`
	Tagline       *string `json:"tagline,omitempty"`
	Mission       *string `json:"mission,omitempty"`
	Vision        *string `json:"vision,omitempty"`
	LogoURL       string  `json:"logoUrl,omitempty"`
	LogoAlt       *string `json:"logoAlt,omitempty"`
	CoverImageURL string  `json:"coverImageUrl,omitempty"`
	CoverImageAlt *string `json:"coverImageAlt,omitempty"`
}

// PublicAbout es «quiénes somos» publicado, ya localizado (FR-003).
type PublicAbout struct {
	Text string `json:"text"`
}

// PublicScheduleItem es un servicio del horario publicado, ya localizado.
type PublicScheduleItem struct {
	ID          string  `json:"id"`
	DayOfWeek   int     `json:"dayOfWeek"`
	StartTime   string  `json:"startTime"`
	EndTime     *string `json:"endTime,omitempty"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Place       string  `json:"place"`
}

// PublicWhatsappChannel es un canal de WhatsApp publicado, ya localizado. URL es
// el enlace listo para abrir (SC-009).
type PublicWhatsappChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
	URL  string `json:"url"`
}

// PublicSocialLink es un enlace de red social publicado.
type PublicSocialLink struct {
	ID      string `json:"id"`
	Network string `json:"network"`
	URL     string `json:"url"`
}

// PublicContact son los datos de contacto publicados, ya localizados (FR-007).
type PublicContact struct {
	Address string `json:"address"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
}

// --- Lectura pública ---

// GetPortada devuelve la portada publicada resuelta a `lang` (FR-001/FR-008/
// FR-009). Un idioma fuera de `es|en` es `400 invalid` (details.lang). No hay
// ninguna vía pública a borradores: todas las lecturas usan las consultas
// `…Published` (FR-013).
func (s *service) GetPortada(ctx context.Context, lang string) (PublicPortada, error) {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		lang = LangES
	}
	if !validLang(lang) {
		return PublicPortada{}, apperr.Invalid(
			"El idioma solicitado no es válido",
			apperr.WithDetails(map[string]any{"lang": "solo se admite es o en"}),
		)
	}

	portada := PublicPortada{Lang: lang}

	if identity, ok, err := s.repository.GetHomeIdentityPublished(ctx); err != nil {
		return PublicPortada{}, fmt.Errorf("leer la identidad publicada: %w", err)
	} else if ok {
		localized := localizeIdentity(identity, lang)
		portada.Identity = &localized
	}

	if about, ok, err := s.repository.GetHomeAboutPublished(ctx); err != nil {
		return PublicPortada{}, fmt.Errorf("leer quiénes somos publicado: %w", err)
	} else if ok {
		portada.About = &PublicAbout{Text: pickString(about.TextEs, about.TextEn, lang)}
	}

	if contact, ok, err := s.repository.GetHomeContactPublished(ctx); err != nil {
		return PublicPortada{}, fmt.Errorf("leer el contacto publicado: %w", err)
	} else if ok {
		portada.Contact = &PublicContact{
			Address: pickString(contact.AddressEs, contact.AddressEn, lang),
			Email:   contact.Email,
			Phone:   contact.Phone,
		}
	}

	services, err := s.repository.ListHomeServicesPublished(ctx)
	if err != nil {
		return PublicPortada{}, fmt.Errorf("leer el horario publicado: %w", err)
	}
	if len(services) > 0 {
		portada.Schedule = make([]PublicScheduleItem, 0, len(services))
		for _, service := range services {
			portada.Schedule = append(portada.Schedule, localizeService(service, lang))
		}
	}

	channels, err := s.repository.ListHomeWhatsappChannelsPublished(ctx)
	if err != nil {
		return PublicPortada{}, fmt.Errorf("leer los canales de WhatsApp publicados: %w", err)
	}
	if len(channels) > 0 {
		portada.Whatsapp = make([]PublicWhatsappChannel, 0, len(channels))
		for _, channel := range channels {
			portada.Whatsapp = append(portada.Whatsapp, localizeWhatsappChannel(channel, lang))
		}
	}

	links, err := s.repository.ListHomeSocialLinksPublished(ctx)
	if err != nil {
		return PublicPortada{}, fmt.Errorf("leer las redes publicadas: %w", err)
	}
	if len(links) > 0 {
		portada.Socials = make([]PublicSocialLink, 0, len(links))
		for _, link := range links {
			portada.Socials = append(portada.Socials, PublicSocialLink{
				ID:      link.ID.String(),
				Network: link.Network,
				URL:     link.URL,
			})
		}
	}

	return portada, nil
}

// OpenMedia aplica la política de descarga de imágenes (analyze C4/M6): un
// nombre que no cumple el patrón generado por el servidor es `400 invalid`; un
// nombre válido pero inexistente o NO referenciado por contenido publicado es
// `404 not_found`; solo entonces se abre el archivo del Store.
func (s *service) OpenMedia(ctx context.Context, fileName string) (io.ReadCloser, error) {
	if !storage.ValidName(fileName) {
		return nil, apperr.Invalid(
			"El nombre del archivo no es válido",
			apperr.WithDetails(map[string]any{"fileName": "formato no reconocido"}),
		)
	}
	published, err := s.repository.IsHomeFilePublished(ctx, fileName)
	if err != nil {
		return nil, fmt.Errorf("comprobar si la imagen está publicada: %w", err)
	}
	if !published {
		return nil, apperr.NotFound("La imagen no está disponible")
	}
	if s.store == nil {
		return nil, apperr.NotFound("La imagen no está disponible")
	}
	reader, err := s.store.Open(ctx, fileName)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, apperr.NotFound("La imagen no está disponible")
		}
		return nil, fmt.Errorf("abrir la imagen: %w", err)
	}
	return reader, nil
}

// --- Localización (fallback en → es por campo, R3-2) ---

// pickString resuelve un campo obligatorio en español: si se pide inglés y hay
// traducción, la usa; si no, devuelve el español (nunca vacío, SC-006).
func pickString(es string, en *string, lang string) string {
	if lang == LangEN && en != nil {
		return *en
	}
	return es
}

// pickOptional resuelve un campo opcional: prefiere la traducción pedida y cae
// al español; si tampoco hay español, devuelve nil (el campo se omite).
func pickOptional(es, en *string, lang string) *string {
	if lang == LangEN && en != nil {
		return en
	}
	return es
}

// localizeIdentity resuelve la identidad publicada al idioma pedido y compone
// las URLs de sus imágenes (`/api/v1/media/<file>`).
func localizeIdentity(identity Identity, lang string) PublicIdentity {
	public := PublicIdentity{
		Name:          pickString(identity.NameEs, identity.NameEn, lang),
		Tagline:       pickOptional(identity.TaglineEs, identity.TaglineEn, lang),
		Mission:       pickOptional(identity.MissionEs, identity.MissionEn, lang),
		Vision:        pickOptional(identity.VisionEs, identity.VisionEn, lang),
		LogoAlt:       pickOptional(identity.LogoAltEs, identity.LogoAltEn, lang),
		CoverImageAlt: pickOptional(identity.CoverImageAltEs, identity.CoverImageAltEn, lang),
	}
	if identity.LogoFile != nil {
		public.LogoURL = MediaPathPrefix + *identity.LogoFile
	}
	if identity.CoverImageFile != nil {
		public.CoverImageURL = MediaPathPrefix + *identity.CoverImageFile
	}
	return public
}

// localizeService resuelve un servicio del horario al idioma pedido.
func localizeService(service Service, lang string) PublicScheduleItem {
	return PublicScheduleItem{
		ID:          service.ID.String(),
		DayOfWeek:   service.DayOfWeek,
		StartTime:   service.StartTime,
		EndTime:     service.EndTime,
		Name:        pickString(service.NameEs, service.NameEn, lang),
		Description: pickOptional(service.DescriptionEs, service.DescriptionEn, lang),
		Place:       pickString(service.PlaceEs, service.PlaceEn, lang),
	}
}

// localizeWhatsappChannel resuelve un canal y arma su enlace público (SC-009):
// `https://wa.me/<dígitos>` para `direct` y la propia URL para `group`.
func localizeWhatsappChannel(channel WhatsappChannel, lang string) PublicWhatsappChannel {
	return PublicWhatsappChannel{
		ID:   channel.ID.String(),
		Name: pickString(channel.NameEs, channel.NameEn, lang),
		Kind: channel.Kind,
		URL:  whatsappPublicURL(channel),
	}
}

// whatsappPublicURL construye el enlace abrible del canal.
func whatsappPublicURL(channel WhatsappChannel) string {
	if channel.Kind == KindDirect {
		return "https://wa.me/" + phoneDigits(channel.Destination)
	}
	return channel.Destination
}
