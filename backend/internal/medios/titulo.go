// Package medios es el dominio de la sección «Medios» de F9 (prédicas y
// podcasts): la ficha por sermón, sus fuentes de importación y el registro de
// cada búsqueda.
//
// Este archivo reúne las funciones **puras** de la regla de unión (FR-013,
// research R9-4): la normalización del título —la llave de comparación, nunca
// el texto publicado— y el día calendario en la zona del lugar (`APP_TZ`).
package medios

import (
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// normalizarTitulo construye la llave de comparación de títulos de FR-013
// (research R9-4): NFKD, sin marcas diacríticas, en mayúsculas Unicode, con la
// puntuación y los símbolos convertidos en espacio y los espacios colapsados y
// recortados.
//
// Es **solo** la llave: el título que se publica es el del feed, tal cual
// (FR-003). Con el par real de los feeds del cliente, «Potencial sin Resultados
// La Lección de la Higuera» y «Potencial sin Resultados: La Lección de la
// Higuera» producen la misma llave.
func normalizarTitulo(titulo string) string {
	// 1) NFKD: separa los caracteres base de sus marcas diacríticas (así
	// «Lección» y «Leccion» coinciden si un feed pierde las tildes).
	descompuesto := norm.NFKD.String(titulo)

	var b strings.Builder
	b.Grow(len(descompuesto))
	for _, r := range descompuesto {
		switch {
		case unicode.Is(unicode.Mn, r):
			// 2) Marca diacrítica (nonspacing mark): se descarta.
			continue
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			// 3) Puntuación y símbolos (incluidos ¿ ¡ – — … y los emojis) se
			// convierten en espacio: son separadores, no contenido.
			b.WriteRune(' ')
		default:
			// 4) Mayúsculas Unicode.
			b.WriteRune(unicode.ToUpper(r))
		}
	}

	// 5) Colapsa los espacios en blanco (incluidos tabuladores y saltos) y
	// recorta los bordes.
	return strings.Join(strings.Fields(b.String()), " ")
}

// diaEnLugar devuelve el día calendario de t en la zona del lugar (FR-006/
// decisión 17 de F4, `APP_TZ` = America/Costa_Rica): la llave de «mismo día» de
// la unión (FR-013, research R9-4) y el corte del orden público.
//
// Se calcula **siempre** en la zona del lugar y **sin margen**: el episodio real
// «Fri, 03 Jul 2026 01:22:48 GMT» es del 2 de julio en Costa Rica (UTC−6), y un
// «mismo día» calculado en UTC habría separado un par que aquí es del mismo día
// (research R9-5).
func diaEnLugar(t time.Time, zona *time.Location) time.Time {
	if zona == nil {
		zona = time.UTC
	}
	local := t.In(zona)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, zona)
}
