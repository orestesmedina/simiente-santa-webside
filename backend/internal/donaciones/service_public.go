package donaciones

import (
	"context"
	"fmt"
)

// service_public.go implementa la cara pública de las ofrendas (T713/T714): la
// sección completa `GET /api/v1/ofrendas` y el bloque de la portada
// `GET /api/v1/portada/ofrendas`. Resuelve el idioma en el servidor con fallback
// `en → es` por campo (FR-008), lee SOLO contenido publicado
// (FR-012/SC-002) y —aquí, en un solo lugar— compone el DTO público de cada
// método (`currency` solo en IBAN y `holder` solo cuando existe): el cliente no
// reimplementa ni el fallback ni la regla de borradores.
//
// El DTO público no tiene campos para `publicationState`: con el estado fuera del
// contrato, un borrador no tiene por dónde salir ni aunque una capa superior se
// descuidara.

// GetOfrendas devuelve la sección pública resuelta a `lang` (FR-002…FR-006): el
// texto introductorio y el de uso si están publicados (cada uno con su estado:
// un texto en borrador no sale aunque haya métodos, US3 esc. 7) y los métodos
// publicados en el orden de la sección (`iban` → `sinpe`; dentro de cada grupo
// `sortOrder`, `id`, FR-004), acotados por `maxPublicItems`.
//
// Cada clave (`items`, `intro`, `usage`) se **omite** cuando no hay nada
// publicado de esa clave (SC-010: sin huecos, nunca `null`); sin nada publicado
// la respuesta es solo `{lang}` y la redirección a la portada la decide el
// frontend (P7-21). Un idioma fuera de `es|en` es `400 invalid` con
// `details.lang`.
func (s *service) GetOfrendas(ctx context.Context, lang string) (OfrendasPublicas, error) {
	lang, err := normalizeLang(lang)
	if err != nil {
		return OfrendasPublicas{}, err
	}

	methods, err := s.repository.ListPublishedDonationMethods(ctx, maxPublicItems)
	if err != nil {
		return OfrendasPublicas{}, fmt.Errorf("listar los métodos publicados: %w", err)
	}
	texts, err := s.repository.ListPublishedDonationTexts(ctx)
	if err != nil {
		return OfrendasPublicas{}, fmt.Errorf("listar los textos publicados: %w", err)
	}

	out := OfrendasPublicas{Lang: lang}
	if len(methods) > 0 {
		out.Items = make([]MetodoDonacionPublico, 0, len(methods))
		for _, method := range methods {
			out.Items = append(out.Items, localizeMetodo(method, lang))
		}
	}
	for _, text := range texts {
		switch text.TextKey {
		case TextKeyIntro:
			out.Intro = &TextoPublico{Text: pickString(text.BodyEs, text.BodyEn, lang)}
		case TextKeyUsage:
			out.Usage = &TextoPublico{Text: pickString(text.BodyEs, text.BodyEn, lang)}
		}
	}
	return out, nil
}

// localizeMetodo traduce un método publicado al DTO del visitante. Es el ÚNICO
// lugar donde se compone el ítem público (FR-003): `kind`, `number` e
// `institution` siempre; `currency` solo en las cuentas IBAN y `holder` solo
// cuando el titular existe (US1 esc. 5: sin líneas vacías ni marcadores). Lo
// comparten la sección y el bloque (FR-022: el mismo esquema de ítem).
func localizeMetodo(method MetodoDonacion, lang string) MetodoDonacionPublico {
	public := MetodoDonacionPublico{
		ID:          method.ID.String(),
		Kind:        string(method.Kind),
		Number:      method.Number,
		Institution: method.Institution,
	}
	if method.Kind == KindIban && method.Currency != nil {
		public.Currency = string(*method.Currency)
	}
	if method.HolderEs != nil {
		public.Holder = pickString(*method.HolderEs, method.HolderEn, lang)
	}
	return public
}

// pickString resuelve un campo obligatorio en español: si se pide inglés y hay
// traducción, la usa; si no, devuelve el español (nunca vacío, SC-006).
func pickString(es string, en *string, lang string) string {
	if lang == LangEN && en != nil {
		return *en
	}
	return es
}
