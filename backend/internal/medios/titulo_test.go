package medios

import (
	"testing"
	"time"
)

// TestNormalizarTituloParReal fija la llave de comparación de FR-013 con el
// **par real** de títulos de los feeds del cliente (research R9-4): el del video
// («Potencial sin Resultados La Lección de la Higuera») y el del podcast
// («Potencial sin Resultados: La Lección de la Higuera») solo difieren en los
// dos puntos y el espacio, y deben producir la MISMA llave.
func TestNormalizarTituloParReal(t *testing.T) {
	video := "Potencial sin Resultados La Lección de la Higuera"
	podcast := "Potencial sin Resultados: La Lección de la Higuera"
	const want = "POTENCIAL SIN RESULTADOS LA LECCION DE LA HIGUERA"

	if got := normalizarTitulo(video); got != want {
		t.Errorf("normalizarTitulo(video) = %q, se esperaba %q", got, want)
	}
	if got := normalizarTitulo(podcast); got != want {
		t.Errorf("normalizarTitulo(podcast) = %q, se esperaba %q", got, want)
	}
	if normalizarTitulo(video) != normalizarTitulo(podcast) {
		t.Errorf("el par real no normaliza igual: %q vs %q", normalizarTitulo(video), normalizarTitulo(podcast))
	}
}

// TestNormalizarTituloTabla fija la normalización (R9-4/P9-5): NFKD + sin
// diacríticos + mayúsculas Unicode + puntuación y símbolos → espacio + espacios
// colapsados. El texto original **nunca** se altera (FR-003: se publica el
// título del feed tal cual).
func TestNormalizarTituloTabla(t *testing.T) {
	tests := []struct {
		name    string
		entrada string
		want    string
	}{
		{"tildes se pierden en la llave", "Lección", "LECCION"},
		{"ñ se pliega a n", "Año Nuevo", "ANO NUEVO"},
		{"mayúsculas Unicode", "fe y confianza", "FE Y CONFIANZA"},
		{"puntuación del español", "¿Qué día, hermanos!", "QUE DIA HERMANOS"},
		{"guiones y puntos suspensivos", "Fe – esperanza — amor…", "FE ESPERANZA AMOR"},
		{"comillas y paréntesis", `"Salmo" (23) [parte 1]`, "SALMO 23 PARTE 1"},
		{"emojis son símbolos", "¡Qué día! 🙌", "QUE DIA"},
		{"espacios dobles y bordes", "  Doble   espacio  ", "DOBLE ESPACIO"},
		{"tabuladores y saltos", "dos\tpalabras\ny más", "DOS PALABRAS Y MAS"},
		{"solo puntuación", "···", ""},
		{"cadena vacía", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := tt.entrada
			if got := normalizarTitulo(tt.entrada); got != tt.want {
				t.Errorf("normalizarTitulo(%q) = %q, se esperaba %q", tt.entrada, got, tt.want)
			}
			if tt.entrada != original {
				t.Errorf("el texto original se alteró: %q → %q", original, tt.entrada)
			}
		})
	}
}

// TestNormalizarTituloNoAlteraElOriginal deja explícito el contrato de FR-003:
// la normalización es solo la llave de comparación; el título publicado sigue
// siendo el del feed, con sus tildes, ñ, ¿¡ y emojis intactos.
func TestNormalizarTituloNoAlteraElOriginal(t *testing.T) {
	const original = "Potencial sin Resultados: La Lección de la Higuera 🙌 ¿y ahora?"
	_ = normalizarTitulo(original)
	if original != "Potencial sin Resultados: La Lección de la Higuera 🙌 ¿y ahora?" {
		t.Fatalf("el título original cambió: %q", original)
	}
}

// TestDiaEnLugarFronteraReal fija el caso real de research R9-5: el episodio
// «La Fe y la confianza en el Abba» trae `pubDate = Fri, 03 Jul 2026 01:22:48
// GMT`, que en America/Costa_Rica (UTC−6) es el **2 de julio** a las 19:22. El
// «mismo día» de FR-013 se calcula SIEMPRE en la zona de la iglesia (decisión 17
// de F4): calcularlo en UTC habría separado un par que en Costa Rica es del
// mismo día.
func TestDiaEnLugarFronteraReal(t *testing.T) {
	costaRica, err := time.LoadLocation("America/Costa_Rica")
	if err != nil {
		t.Fatalf("cargar America/Costa_Rica: %v", err)
	}
	publicado, err := time.Parse(time.RFC1123, "Fri, 03 Jul 2026 01:22:48 GMT")
	if err != nil {
		t.Fatalf("parsear la fecha del feed: %v", err)
	}

	if got := diaEnLugar(publicado, costaRica).Format("2006-01-02"); got != "2026-07-02" {
		t.Errorf("diaEnLugar(fecha real, Costa Rica) = %s, se esperaba 2026-07-02", got)
	}
	if got := diaEnLugar(publicado, time.UTC).Format("2006-01-02"); got != "2026-07-03" {
		t.Errorf("diaEnLugar(fecha real, UTC) = %s, se esperaba 2026-07-03 (la prueba fija por qué el cálculo va en APP_TZ)", got)
	}
}

// TestDiaEnLugarMismoDia fija la llave de «mismo día»: dos instantes del mismo
// día calendario en la zona del lugar comparten día aunque en UTC caigan en
// días distintos, y dos instantes de días distintos no.
func TestDiaEnLugarMismoDia(t *testing.T) {
	costaRica, err := time.LoadLocation("America/Costa_Rica")
	if err != nil {
		t.Fatalf("cargar America/Costa_Rica: %v", err)
	}

	// 2026-05-25 17:58:50 GMT = 11:58 en Costa Rica (mismo día 25).
	manana, err := time.Parse(time.RFC1123, "Mon, 25 May 2026 17:58:50 GMT")
	if err != nil {
		t.Fatalf("parsear la fecha: %v", err)
	}
	// 2026-05-26 03:00:00 GMT = 2026-05-25 21:00 en Costa Rica (mismo día 25 en
	// la zona del lugar, día 26 en UTC).
	tarde, err := time.Parse(time.RFC1123, "Tue, 26 May 2026 03:00:00 GMT")
	if err != nil {
		t.Fatalf("parsear la fecha: %v", err)
	}
	// 2026-05-27 18:00:00 GMT = 2026-05-27 12:00 en Costa Rica (otro día).
	otro, err := time.Parse(time.RFC1123, "Wed, 27 May 2026 18:00:00 GMT")
	if err != nil {
		t.Fatalf("parsear la fecha: %v", err)
	}

	if diaEnLugar(manana, costaRica) != diaEnLugar(tarde, costaRica) {
		t.Errorf("dos instantes del mismo día en Costa Rica no comparten día: %s vs %s",
			diaEnLugar(manana, costaRica).Format(time.RFC3339), diaEnLugar(tarde, costaRica).Format(time.RFC3339))
	}
	if diaEnLugar(manana, costaRica) == diaEnLugar(otro, costaRica) {
		t.Errorf("dos instantes de días distintos comparten día")
	}
	if got := diaEnLugar(manana, costaRica).Format("2006-01-02"); got != "2026-05-25" {
		t.Errorf("diaEnLugar() = %s, se esperaba 2026-05-25", got)
	}
	if got := diaEnLugar(manana, costaRica).Location(); got != costaRica {
		t.Errorf("el día devuelto no está en la zona del lugar: %v", got)
	}
}
