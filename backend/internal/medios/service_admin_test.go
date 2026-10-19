package medios

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
)

// service_admin_test.go cubre la gestión del panel (T918/T919): el listado con
// `lastSync` y la edición de la descripción corta y la visibilidad. Fija la
// regla exacta de auditoría de FR-022 (una fila por cambio de datos, otra por
// cambio de visibilidad) y la protección de lo que decidió el equipo
// (FR-015/SC-007: `descriptionEdited`).

// TestListMediosDelPanel fija el listado del panel (FR-019): **todo** —ocultas y
// marcas incluidas— menos las absorbidas (salvo `IncluirAbsorbidas`), con el
// mismo orden que ve el visitante, el filtro por marca **pendiente** de cada eje
// y `lastSync` solo cuando ya hubo una búsqueda.
func TestListMediosDelPanel(t *testing.T) {
	base := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)
	repo := newFakeRepository()
	visible := Episodio{Titulo: "visible", PublicadoEn: base}
	oculta := Episodio{Titulo: "oculta", PublicadoEn: base.Add(-time.Hour), Visibilidad: VisibilidadOculta}
	absorbida := Episodio{Titulo: "absorbida", PublicadoEn: base.Add(-2 * time.Hour), UnidoAID: ptrUUID(uuid.New())}
	parAbierto := Episodio{Titulo: "par abierto", PublicadoEn: base.Add(-3 * time.Hour), DuplicadoDeID: ptrUUID(uuid.New())}
	fueraDelFeed := Episodio{
		Titulo:              "fuera del feed",
		PublicadoEn:         base.Add(-4 * time.Hour),
		MarcaDisponibilidad: ptrMarca(MarcaDisponibilidadFueraDelFeed),
	}
	repo.seed(visible, oculta, absorbida, parAbierto, fueraDelFeed)
	repo.seedFuentes(FuenteEpisodio{
		EpisodioID: repo.episodios[0].ID,
		Proveedor:  ProveedorYouTube,
		IDExterno:  "SimienteAA1",
		TituloFeed: visible.Titulo,
		Enlace:     "https://www.youtube.com/watch?v=SimienteAA1",
	})
	service := serviceDePrueba(repo)

	t.Run("todo menos las absorbidas", func(t *testing.T) {
		out, err := service.List(context.Background(), FiltroPanel{}, paginate.Params{Limit: 20})
		if err != nil {
			t.Fatalf("List devolvió %v", err)
		}
		if out.Total != 4 || len(out.Items) != 4 {
			t.Fatalf("listado = %d fichas (total %d), se esperaban 4", len(out.Items), out.Total)
		}
		if out.Items[0].Titulo != "visible" || out.Items[3].Titulo != "fuera del feed" {
			t.Errorf("orden del panel = %v", titulosAdmin(out.Items))
		}
		if out.Items[0].Visibilidad != string(VisibilidadVisible) || out.Items[1].Visibilidad != string(VisibilidadOculta) {
			t.Errorf("la visibilidad no viaja en la fila: %+v", out.Items)
		}
		if len(out.Items[0].Fuentes) != 1 || out.Items[0].Fuentes[0].Proveedor != string(ProveedorYouTube) {
			t.Errorf("la procedencia no viaja en la fila: %+v", out.Items[0].Fuentes)
		}
		if out.LastSync != nil {
			t.Errorf("lastSync debe estar ausente antes de la primera búsqueda: %+v", out.LastSync)
		}
		if out.Limit != 20 || out.Offset != 0 {
			t.Errorf("sobre de paginación = {%d, %d}", out.Limit, out.Offset)
		}
	})

	t.Run("con absorbidas", func(t *testing.T) {
		out, err := service.List(context.Background(), FiltroPanel{IncluirAbsorbidas: true}, paginate.Params{Limit: 20})
		if err != nil {
			t.Fatalf("List devolvió %v", err)
		}
		if out.Total != 5 {
			t.Fatalf("total con absorbidas = %d, se esperaba 5", out.Total)
		}
		if out.Items[2].Titulo != "absorbida" || out.Items[2].UnidoAID == "" {
			t.Errorf("la absorbida debe viajar con mergedIntoId: %+v", out.Items[2])
		}
	})

	t.Run("filtro por visibilidad", func(t *testing.T) {
		ocultaFiltro := VisibilidadOculta
		out, err := service.List(context.Background(), FiltroPanel{Visibilidad: &ocultaFiltro}, paginate.Params{Limit: 20})
		if err != nil {
			t.Fatalf("List devolvió %v", err)
		}
		if out.Total != 1 || out.Items[0].Titulo != "oculta" {
			t.Errorf("filtro por visibilidad = %v", titulosAdmin(out.Items))
		}
	})

	t.Run("filtro por marca pendiente de cada eje", func(t *testing.T) {
		duplicado := MarcaPosibleDuplicado
		out, err := service.List(context.Background(), FiltroPanel{MarcaRevision: &duplicado}, paginate.Params{Limit: 20})
		if err != nil {
			t.Fatalf("List devolvió %v", err)
		}
		if out.Total != 1 || out.Items[0].Titulo != "par abierto" {
			t.Errorf("filtro possible_duplicate = %v", titulosAdmin(out.Items))
		}

		fuera := MarcaFueraDelFeed
		out, err = service.List(context.Background(), FiltroPanel{MarcaRevision: &fuera}, paginate.Params{Limit: 20})
		if err != nil {
			t.Fatalf("List devolvió %v", err)
		}
		if out.Total != 1 || out.Items[0].Titulo != "fuera del feed" {
			t.Errorf("filtro missing_from_feed = %v", titulosAdmin(out.Items))
		}
	})

	t.Run("marcas y disponibilidad viajan en la fila", func(t *testing.T) {
		out, err := service.List(context.Background(), FiltroPanel{}, paginate.Params{Limit: 20})
		if err != nil {
			t.Fatalf("List devolvió %v", err)
		}
		fila := out.Items[3]
		if fila.MarcaDisponibilidad != string(MarcaDisponibilidadFueraDelFeed) {
			t.Errorf("availabilityMark = %q", fila.MarcaDisponibilidad)
		}
	})

	t.Run("lastSync cuando ya hubo una búsqueda", func(t *testing.T) {
		repo.syncRuns = append(repo.syncRuns, SyncRun{
			IniciadoEn: base, TerminadoEn: base.Add(time.Second),
			Importados: 10, Unidos: 8, Marcados: 2, Ocultos: 1, Comprobados: 3, Inconclusos: 1,
			EstadoYouTube: FeedOK, EstadoPodcast: FeedFallido,
		})
		out, err := service.List(context.Background(), FiltroPanel{}, paginate.Params{Limit: 20})
		if err != nil {
			t.Fatalf("List devolvió %v", err)
		}
		if out.LastSync == nil {
			t.Fatal("lastSync debe estar presente tras la primera búsqueda")
		}
		if out.LastSync.Importados != 10 || out.LastSync.Unidos != 8 || out.LastSync.Marcados != 2 ||
			out.LastSync.Ocultos != 1 || out.LastSync.Comprobados != 3 || out.LastSync.Inconclusos != 1 {
			t.Errorf("lastSync = %+v", out.LastSync)
		}
		if out.LastSync.EstadoYouTube != string(FeedOK) || out.LastSync.EstadoPodcast != string(FeedFallido) {
			t.Errorf("el estado de los feeds no viaja: %+v", out.LastSync)
		}
	})

	t.Run("filtros fuera del catálogo", func(t *testing.T) {
		invalida := Visibilidad("draft")
		if _, err := service.List(context.Background(), FiltroPanel{Visibilidad: &invalida}, paginate.Params{Limit: 20}); err == nil {
			t.Error("una visibilidad fuera del catálogo debía ser invalid")
		}
		desconocida := MarcaRevision("otra")
		if _, err := service.List(context.Background(), FiltroPanel{MarcaRevision: &desconocida}, paginate.Params{Limit: 20}); err == nil {
			t.Error("una marca fuera del catálogo debía ser invalid")
		}
	})
}

// TestUpdateMedioDescripcion fija la edición de la descripción corta
// (FR-019/FR-020): validación completa con `details` por campo y **sin guardar
// nada a medias**, el guardado fija `descriptionEdited` (FR-015/SC-007: la
// sincronización no la pisa), los caracteres del español se conservan y `""`
// **vacía** la descripción —dejando la ficha sin ella, con su fila de auditoría
// (decisión del humano del 2026-10-18)— con el mismo tratamiento que el inglés.
func TestUpdateMedioDescripcion(t *testing.T) {
	actor := uuid.New()
	base := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)

	t.Run("guarda la descripción y la marca como del equipo", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Potencial sin Resultados", PublicadoEn: base, DescripcionEs: ptr("Del feed")})
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		out, err := service.Update(context.Background(), actor, id, MedioPatch{DescriptionEs: ptr("  ¿Qué dice la Lección de la niñez? 🙌  ")})
		if err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		if out.DescripcionEs == nil || *out.DescripcionEs != "¿Qué dice la Lección de la niñez? 🙌" {
			t.Errorf("descripción guardada = %v", out.DescripcionEs)
		}
		if !out.DescripcionEditada {
			t.Error("descriptionEdited debe quedar en true: la descripción es del equipo")
		}
		if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaDescriptionUpdate {
			t.Errorf("auditoría = %+v, se esperaba una fila media.description.update", repo.committed)
		}
		if repo.committed[0].TargetLabel != "Medios · Potencial sin Resultados" {
			t.Errorf("targetLabel = %q", repo.committed[0].TargetLabel)
		}
		if repo.committed[0].TargetKind != audit.TargetContent || repo.committed[0].Result != audit.ResultSuccess {
			t.Errorf("objetivo/resultado = %+v", repo.committed[0])
		}
	})

	t.Run("borra la traducción opcional con vacío", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{
			Titulo: "Traducida", PublicadoEn: base,
			DescripcionEs: ptr("Una prédica"), DescripcionEn: ptr("A sermon"), DescripcionEditada: true,
		})
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		out, err := service.Update(context.Background(), actor, id, MedioPatch{DescriptionEn: ptr("   ")})
		if err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		if out.DescripcionEn != nil {
			t.Errorf("descriptionEn = %v, se esperaba NULL (sin traducción)", *out.DescripcionEn)
		}
	})

	t.Run("vacía la descripción con vacío", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{
			Titulo: "Sin descripción", PublicadoEn: base,
			DescripcionEs: ptr("Del feed"), DescripcionEditada: true,
		})
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		out, err := service.Update(context.Background(), actor, id, MedioPatch{DescriptionEs: ptr("")})
		if err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		if out.DescripcionEs != nil {
			t.Errorf("descriptionEs = %v, se esperaba NULL (sin descripción)", *out.DescripcionEs)
		}
		if !out.DescripcionEditada {
			t.Error("descriptionEdited debe quedar en true: el vaciado es del equipo")
		}
		if len(repo.updates) != 1 {
			t.Fatalf("updates = %v, se esperaba una escritura", repo.updates)
		}
		if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaDescriptionUpdate {
			t.Errorf("auditoría = %+v, se esperaba una fila media.description.update", repo.committed)
		}
		if repo.committed[0].TargetLabel != "Medios · Sin descripción" {
			t.Errorf("targetLabel = %q", repo.committed[0].TargetLabel)
		}
	})

	t.Run("vacía la descripción con solo espacios", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Sin descripción", PublicadoEn: base, DescripcionEs: ptr("Del feed")})
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		out, err := service.Update(context.Background(), actor, id, MedioPatch{DescriptionEs: ptr("   ")})
		if err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		if out.DescripcionEs != nil {
			t.Errorf("descriptionEs = %v, se esperaba NULL (sin descripción)", *out.DescripcionEs)
		}
		if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaDescriptionUpdate {
			t.Errorf("auditoría = %+v, se esperaba una fila media.description.update", repo.committed)
		}
	})

	t.Run("vacía las dos descripciones", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{
			Titulo: "Traducida", PublicadoEn: base,
			DescripcionEs: ptr("Una prédica"), DescripcionEn: ptr("A sermon"), DescripcionEditada: true,
		})
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		out, err := service.Update(context.Background(), actor, id,
			MedioPatch{DescriptionEs: ptr(""), DescriptionEn: ptr("")})
		if err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		if out.DescripcionEs != nil || out.DescripcionEn != nil {
			t.Errorf("descripciones = (%v, %v), se esperaban NULL", out.DescripcionEs, out.DescripcionEn)
		}
		// Un solo cambio de datos: una sola fila (FR-022).
		if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaDescriptionUpdate {
			t.Errorf("auditoría = %+v, se esperaba una fila media.description.update", repo.committed)
		}
	})

	t.Run("vaciar el español con traducción presente sigue exigiendo el inglés fuera", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{
			Titulo: "Traducida", PublicadoEn: base,
			DescripcionEs: ptr("Una prédica"), DescripcionEn: ptr("A sermon"), DescripcionEditada: true,
		})
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		// FR-023 sigue en pie: el inglés exige el español por campo.
		_, err := service.Update(context.Background(), actor, id, MedioPatch{DescriptionEs: ptr("")})
		requireInvalid(t, err, "descriptionEs")
		if len(repo.updates) != 0 || len(repo.committed) != 0 {
			t.Error("un rechazo no debe guardar nada ni dejar auditoría de éxito")
		}
		if len(repo.recorded) != 1 || repo.recorded[0].Result != audit.ResultFailure {
			t.Errorf("el rechazo debe quedar registrado best-effort: %+v", repo.recorded)
		}
	})

	t.Run("vaciar una descripción que ya estaba vacía no es una operación", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Sin descripción", PublicadoEn: base})
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		out, err := service.Update(context.Background(), actor, id, MedioPatch{DescriptionEs: ptr("")})
		if err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		if out.DescripcionEs != nil {
			t.Errorf("descriptionEs = %v, se esperaba NULL", *out.DescripcionEs)
		}
		if len(repo.updates) != 0 || len(repo.committed) != 0 {
			t.Errorf("un PATCH sin cambios no debe escribir ni registrar: %v / %+v", repo.updates, repo.committed)
		}
		if repo.episodios[0].DescripcionEditada {
			t.Error("sin cambio de datos, la descripción no debe quedar marcada como del equipo")
		}
	})

	t.Run("inglés sin español", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Sin descripción", PublicadoEn: base})
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		_, err := service.Update(context.Background(), actor, id, MedioPatch{DescriptionEn: ptr("A sermon")})
		requireInvalid(t, err, "descriptionEs")
		if len(repo.updates) != 0 || len(repo.committed) != 0 {
			t.Error("un rechazo no debe guardar nada ni dejar auditoría de éxito")
		}
		if len(repo.recorded) != 1 || repo.recorded[0].Result != audit.ResultFailure {
			t.Errorf("el rechazo debe quedar registrado best-effort: %+v", repo.recorded)
		}
	})

	t.Run("campos inválidos", func(t *testing.T) {
		casos := []struct {
			name  string
			patch MedioPatch
			campo string
		}{
			{name: "descripción en español de 301", patch: MedioPatch{DescriptionEs: ptr(strings.Repeat("a", maxDescripcion+1))}, campo: "descriptionEs"},
			{name: "descripción en inglés de 301", patch: MedioPatch{DescriptionEn: ptr(strings.Repeat("a", maxDescripcion+1))}, campo: "descriptionEn"},
			{name: "visibilidad fuera del catálogo", patch: MedioPatch{Visibility: ptr("draft")}, campo: "visibility"},
		}
		for _, caso := range casos {
			t.Run(caso.name, func(t *testing.T) {
				repo := newFakeRepository()
				repo.seed(Episodio{Titulo: "Ficha", PublicadoEn: base, DescripcionEs: ptr("Del feed")})
				service := serviceDePrueba(repo)
				id := repo.episodios[0].ID

				_, err := service.Update(context.Background(), actor, id, caso.patch)
				requireInvalid(t, err, caso.campo)
				if len(repo.updates) != 0 {
					t.Error("no debe guardarse nada a medias")
				}
				if repo.episodios[0].DescripcionEditada {
					t.Error("la descripción no debe quedar marcada como del equipo")
				}
			})
		}
	})

	t.Run("sin cambios no escribe ni audita", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Ficha", PublicadoEn: base, DescripcionEs: ptr("Del feed")})
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		out, err := service.Update(context.Background(), actor, id, MedioPatch{DescriptionEs: ptr("Del feed")})
		if err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		if out.DescripcionEditada {
			t.Error("un PATCH idéntico no convierte la descripción en del equipo")
		}
		if len(repo.updates) != 0 || len(repo.committed) != 0 {
			t.Errorf("un PATCH sin cambios no escribe ni audita: %+v / %+v", repo.updates, repo.committed)
		}
	})

	t.Run("id inexistente", func(t *testing.T) {
		repo := newFakeRepository()
		service := serviceDePrueba(repo)

		_, err := service.Update(context.Background(), actor, uuid.New(), MedioPatch{DescriptionEs: ptr("Nueva")})
		requireKind(t, err, apperr.KindNotFound)
	})

	t.Run("si la auditoría falla no se aplica nada", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Ficha", PublicadoEn: base, DescripcionEs: ptr("Del feed")})
		repo.auditFalla = true
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		if _, err := service.Update(context.Background(), actor, id, MedioPatch{DescriptionEs: ptr("Nueva")}); err == nil {
			t.Fatal("se esperaba el fallo del registro")
		}
		if repo.episodios[0].DescripcionEs == nil || *repo.episodios[0].DescripcionEs != "Del feed" {
			t.Error("la ficha no debe cambiar si su auditoría no se pudo registrar (FR-022)")
		}
	})
}

// TestUpdateMedioVisibilidad fija la regla exacta de códigos de FR-022: el cambio
// de visibilidad registra `media.hide`/`media.show` y un guardado que cambia
// datos **y** visibilidad deja **dos** filas (patrón de F7).
func TestUpdateMedioVisibilidad(t *testing.T) {
	actor := uuid.New()
	base := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)

	t.Run("ocultar deja media.hide", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Ficha", PublicadoEn: base})
		service := serviceDePrueba(repo)

		out, err := service.Update(context.Background(), actor, repo.episodios[0].ID, MedioPatch{Visibility: ptr("hidden")})
		if err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		if out.Visibilidad != string(VisibilidadOculta) {
			t.Errorf("visibility = %q", out.Visibilidad)
		}
		if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaHide {
			t.Errorf("auditoría = %+v, se esperaba una fila media.hide", repo.committed)
		}
	})

	t.Run("mostrar deja media.show", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Ficha", PublicadoEn: base, Visibilidad: VisibilidadOculta})
		service := serviceDePrueba(repo)

		if _, err := service.Update(context.Background(), actor, repo.episodios[0].ID, MedioPatch{Visibility: ptr("visible")}); err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaShow {
			t.Errorf("auditoría = %+v, se esperaba una fila media.show", repo.committed)
		}
	})

	t.Run("datos y visibilidad dejan dos filas", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Ficha", PublicadoEn: base, DescripcionEs: ptr("Del feed")})
		service := serviceDePrueba(repo)

		_, err := service.Update(context.Background(), actor, repo.episodios[0].ID, MedioPatch{
			DescriptionEs: ptr("Del equipo"),
			Visibility:    ptr("hidden"),
		})
		if err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		if len(repo.committed) != 2 {
			t.Fatalf("auditoría = %+v, se esperaban dos filas", repo.committed)
		}
		codigos := []string{repo.committed[0].Code, repo.committed[1].Code}
		if codigos[0] != audit.ActionMediaDescriptionUpdate || codigos[1] != audit.ActionMediaHide {
			t.Errorf("códigos = %v", codigos)
		}
	})

	t.Run("repetir la misma visibilidad no audita", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Ficha", PublicadoEn: base})
		service := serviceDePrueba(repo)

		if _, err := service.Update(context.Background(), actor, repo.episodios[0].ID, MedioPatch{Visibility: ptr("visible")}); err != nil {
			t.Fatalf("Update devolvió %v", err)
		}
		if len(repo.committed) != 0 || len(repo.updates) != 0 {
			t.Errorf("no hubo cambio: %+v / %+v", repo.committed, repo.updates)
		}
	})
}

// ---------------------------------------------------------------------------
// Unión por puntero (T919)
// ---------------------------------------------------------------------------

// TestMergeDejaUnaFichaConLosDosEnlaces fija P9-10/FR-013/FR-019: `{id}` se
// conserva (con los dos enlaces, `origin='both'` y las dos procedencias) y la
// otra queda absorbida —`merged_into_id` + oculta, fuera de todas las consultas—
// **sin ningún DELETE** (la fila sigue existiendo).
func TestMergeDejaUnaFichaConLosDosEnlaces(t *testing.T) {
	actor := uuid.New()
	base := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)

	t.Run("la absorbida desaparece de la vía pública y del panel por defecto", func(t *testing.T) {
		repo := newFakeRepository()
		videoID, podcastID := sembrarVideoYAudio(repo, base)
		service := serviceDePrueba(repo)

		out, err := service.Merge(context.Background(), actor, podcastID, videoID)
		if err != nil {
			t.Fatalf("Merge devolvió %v", err)
		}
		if out.Origen != string(OrigenAmbos) {
			t.Errorf("origin = %q, se esperaba both", out.Origen)
		}
		if len(out.Fuentes) != 2 {
			t.Fatalf("la ficha resultante tiene %d fuentes, se esperaban 2", len(out.Fuentes))
		}
		if out.VideoURL == "" || out.AudioURL == "" {
			t.Errorf("faltan enlaces en la resultante: %+v", out)
		}
		if out.DuplicadoDeID != "" || out.MarcaDuplicadoResueltaEn != nil {
			t.Errorf("el eje de duplicados debe quedar despejado: %+v", out)
		}
		if out.Visibilidad != string(VisibilidadVisible) {
			t.Errorf("la base era visible y debe seguir visible: %q", out.Visibilidad)
		}

		// Cero DELETE: la absorbida sigue en la tabla, con puntero y oculta.
		absorbida, ok := repo.episodios[1], false
		for _, ep := range repo.episodios {
			if ep.ID == videoID {
				absorbida, ok = ep, true
			}
		}
		if !ok {
			t.Fatal("la ficha absorbida desapareció del almacén (FR-019: sin DELETE)")
		}
		if absorbida.UnidoAID == nil || *absorbida.UnidoAID != podcastID {
			t.Errorf("mergedIntoId = %v, se esperaba %s", absorbida.UnidoAID, podcastID)
		}
		if absorbida.Visibilidad != VisibilidadOculta {
			t.Errorf("la absorbida debe quedar oculta, no %q", absorbida.Visibilidad)
		}

		// El repunte fue en UNA sola transacción (una llamada al repositorio).
		if len(repo.repoints) != 1 || repo.repoints[0] != [2]uuid.UUID{podcastID, videoID} {
			t.Errorf("repuntes = %v, se esperaba [{%s %s}]", repo.repoints, podcastID, videoID)
		}
		if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaMerge {
			t.Errorf("auditoría = %+v, se esperaba una fila media.merge", repo.committed)
		}
		if !strings.Contains(repo.committed[0].TargetLabel, videoID.String()) {
			t.Errorf("la absorbida debe citarse en targetLabel: %q", repo.committed[0].TargetLabel)
		}

		// La absorbida queda fuera de todas las consultas (como el predicado real).
		publicos, err := service.GetMedios(context.Background(), LangES, 0, 0)
		if err != nil {
			t.Fatalf("GetMedios devolvió %v", err)
		}
		if len(publicos.Items) != 1 || publicos.Items[0].ID != podcastID.String() {
			t.Errorf("la vía pública = %v", titulosPublicos(publicos.Items))
		}
		panel, err := service.List(context.Background(), FiltroPanel{}, paginate.Params{Limit: 20})
		if err != nil {
			t.Fatalf("List devolvió %v", err)
		}
		if panel.Total != 1 {
			t.Errorf("el panel por defecto no debe listar absorbidas: total %d", panel.Total)
		}
	})

	t.Run("el nombre, la carátula y la fecha son los del video (precisión 5)", func(t *testing.T) {
		repo := newFakeRepository()
		videoID, podcastID := sembrarVideoYAudio(repo, base)
		service := serviceDePrueba(repo)

		out, err := service.Merge(context.Background(), actor, podcastID, videoID)
		if err != nil {
			t.Fatalf("Merge devolvió %v", err)
		}
		if out.Titulo != "Potencial sin Resultados La Lección de la Higuera" {
			t.Errorf("el nombre que manda es el del feed del video: %q", out.Titulo)
		}
		if out.CaratulaURL != "/api/v1/media/img_33333333-3333-3333-3333-333333333333.jpg" {
			t.Errorf("la carátula que manda es la del video: %q", out.CaratulaURL)
		}

		// El orden inverso (la base es el video) conserva lo suyo.
		repo2 := newFakeRepository()
		videoID2, podcastID2 := sembrarVideoYAudio(repo2, base)
		service2 := serviceDePrueba(repo2)
		out2, err := service2.Merge(context.Background(), actor, videoID2, podcastID2)
		if err != nil {
			t.Fatalf("Merge devolvió %v", err)
		}
		if out2.Titulo != "Potencial sin Resultados La Lección de la Higuera" {
			t.Errorf("unir en el otro orden debe publicar el mismo nombre: %q", out2.Titulo)
		}
		if out2.Origen != string(OrigenAmbos) || len(out2.Fuentes) != 2 {
			t.Errorf("la unión en el otro orden = %+v", out2)
		}
	})

	t.Run("manda la descripción editada y sobrevive a la unión", func(t *testing.T) {
		for _, orden := range []string{"base primero", "absorbida primero"} {
			t.Run(orden, func(t *testing.T) {
				repo := newFakeRepository()
				videoID, podcastID := sembrarVideoYAudio(repo, base)
				// Solo la ficha del podcast tiene la descripción ajustada.
				repo.episodios[1].DescripcionEs = ptr("Del equipo")
				repo.episodios[1].DescripcionEditada = true
				service := serviceDePrueba(repo)

				baseID, otraID := podcastID, videoID
				if orden == "absorbida primero" {
					baseID, otraID = videoID, podcastID
				}
				out, err := service.Merge(context.Background(), actor, baseID, otraID)
				if err != nil {
					t.Fatalf("Merge devolvió %v", err)
				}
				if out.DescripcionEs == nil || *out.DescripcionEs != "Del equipo" || !out.DescripcionEditada {
					t.Errorf("la descripción del equipo debe sobrevivir: %+v", out)
				}
			})
		}
	})

	t.Run("el eje de disponibilidad de la base no se toca", func(t *testing.T) {
		repo := newFakeRepository()
		videoID, podcastID := sembrarVideoYAudio(repo, base)
		repo.episodios[1].MarcaDisponibilidad = ptrMarca(MarcaDisponibilidadFueraDelFeed)
		service := serviceDePrueba(repo)

		out, err := service.Merge(context.Background(), actor, podcastID, videoID)
		if err != nil {
			t.Fatalf("Merge devolvió %v", err)
		}
		if out.MarcaDisponibilidad != string(MarcaDisponibilidadFueraDelFeed) {
			t.Errorf("la marca de disponibilidad abierta debe seguir: %+v", out)
		}
	})

	t.Run("la carátula de la absorbida que queda sin uso se borra en mejor esfuerzo", func(t *testing.T) {
		repo := newFakeRepository()
		videoID, podcastID := sembrarVideoYAudio(repo, base)
		almacen := &fakeAlmacen{}
		service := NewService(ServiceDeps{Repository: repo, Store: almacen})

		if _, err := service.Merge(context.Background(), actor, podcastID, videoID); err != nil {
			t.Fatalf("Merge devolvió %v", err)
		}
		if len(almacen.borrados) != 1 || almacen.borrados[0] != "img_44444444-4444-4444-4444-444444444444.jpg" {
			t.Errorf("borrados = %v, se esperaba la carátula propia del podcast", almacen.borrados)
		}
	})
}

// TestMergeErrores fija los errores del contrato: unir consigo misma es
// `invalid`, una ficha inexistente o ya absorbida es `not_found` y dos fichas con
// fuente del **mismo** proveedor (no caben en una ficha: una entrada por sermón)
// es `conflict`.
func TestMergeErrores(t *testing.T) {
	actor := uuid.New()
	base := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)

	t.Run("unir consigo misma", func(t *testing.T) {
		repo := newFakeRepository()
		videoID, _ := sembrarVideoYAudio(repo, base)
		service := serviceDePrueba(repo)

		_, err := service.Merge(context.Background(), actor, videoID, videoID)
		requireInvalid(t, err, "otherId")
		if len(repo.repoints) != 0 {
			t.Error("un rechazo no puede tocar las fuentes")
		}
	})

	t.Run("ficha inexistente o ya absorbida", func(t *testing.T) {
		repo := newFakeRepository()
		videoID, podcastID := sembrarVideoYAudio(repo, base)
		service := serviceDePrueba(repo)

		_, err := service.Merge(context.Background(), actor, podcastID, uuid.New())
		requireKind(t, err, apperr.KindNotFound)

		if _, err := service.Merge(context.Background(), actor, podcastID, videoID); err != nil {
			t.Fatalf("la primera unión devolvió %v", err)
		}
		_, err = service.Merge(context.Background(), actor, podcastID, videoID)
		requireKind(t, err, apperr.KindNotFound)
	})

	t.Run("dos fuentes del mismo proveedor no caben en una ficha", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(
			Episodio{Titulo: "Video A", PublicadoEn: base, Origen: OrigenVideo},
			Episodio{Titulo: "Video B", PublicadoEn: base, Origen: OrigenVideo},
		)
		repo.seedFuentes(
			FuenteEpisodio{EpisodioID: repo.episodios[0].ID, Proveedor: ProveedorYouTube,
				IDExterno: "SimienteAA1", TituloFeed: "Video A", Enlace: "https://www.youtube.com/watch?v=SimienteAA1"},
			FuenteEpisodio{EpisodioID: repo.episodios[1].ID, Proveedor: ProveedorYouTube,
				IDExterno: "SimienteAA2", TituloFeed: "Video B", Enlace: "https://www.youtube.com/watch?v=SimienteAA2"},
		)
		service := serviceDePrueba(repo)

		_, err := service.Merge(context.Background(), actor, repo.episodios[0].ID, repo.episodios[1].ID)
		requireKind(t, err, apperr.KindConflict)
	})
}

// TestResolveMark fija la resolución de marcas (FR-017/US3 esc. 6, P9-19): la
// operación indica **qué eje** resuelve, la visibilidad la decide el equipo, es
// idempotente sobre una marca ya resuelta y «no son el mismo sermón» deja **las
// dos** fichas del par visibles y con el par resuelto.
func TestResolveMark(t *testing.T) {
	actor := uuid.New()
	base := time.Date(2026, 7, 2, 15, 0, 0, 0, time.UTC)

	t.Run("no son el mismo sermón deja las dos visibles", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(
			Episodio{Titulo: "Video", PublicadoEn: base, Origen: OrigenVideo},
			Episodio{Titulo: "Audio", PublicadoEn: base, Origen: OrigenAudio, Visibilidad: VisibilidadOculta},
		)
		video, audio := repo.episodios[0].ID, repo.episodios[1].ID
		repo.episodios[0].DuplicadoDeID = ptrUUID(audio)
		repo.episodios[1].DuplicadoDeID = ptrUUID(video)
		service := serviceDePrueba(repo)

		out, err := service.ResolveMark(context.Background(), actor, video, MarcaPosibleDuplicado, VisibilidadVisible)
		if err != nil {
			t.Fatalf("ResolveMark devolvió %v", err)
		}
		if out.MarcaDuplicadoResueltaEn == nil {
			t.Error("el par debe quedar resuelto")
		}
		if out.DuplicadoDeID != audio.String() {
			t.Errorf("el par se conserva como índice de no-reapertura: %+v", out.DuplicadoDeID)
		}
		otra := repo.episodios[1]
		if otra.Visibilidad != VisibilidadVisible || otra.MarcaDuplicadoResueltaEn == nil {
			t.Errorf("la otra ficha del par debe quedar visible y resuelta: %+v", otra)
		}
		// Las dos salen al público (US3 esc. 2).
		publicos, err := service.GetMedios(context.Background(), LangES, 0, 0)
		if err != nil {
			t.Fatalf("GetMedios devolvió %v", err)
		}
		if len(publicos.Items) != 2 {
			t.Errorf("el público debe ver las dos: %v", titulosPublicos(publicos.Items))
		}
		if len(repo.committed) != 1 || repo.committed[0].Code != audit.ActionMediaMarkResolve {
			t.Errorf("auditoría = %+v, se esperaba una fila media.mark.resolve", repo.committed)
		}
		if !strings.Contains(repo.committed[0].TargetLabel, string(MarcaPosibleDuplicado)) ||
			!strings.Contains(repo.committed[0].TargetLabel, string(VisibilidadVisible)) {
			t.Errorf("la marca y la visibilidad deben ir en targetLabel: %q", repo.committed[0].TargetLabel)
		}
	})

	t.Run("si la auditoría falla no se resuelve ninguna de las dos", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(
			Episodio{Titulo: "Video", PublicadoEn: base, Origen: OrigenVideo},
			Episodio{Titulo: "Audio", PublicadoEn: base, Origen: OrigenAudio, Visibilidad: VisibilidadOculta},
		)
		video, audio := repo.episodios[0].ID, repo.episodios[1].ID
		repo.episodios[0].DuplicadoDeID = ptrUUID(audio)
		repo.episodios[1].DuplicadoDeID = ptrUUID(video)
		repo.auditFalla = true
		service := serviceDePrueba(repo)

		if _, err := service.ResolveMark(context.Background(), actor, video, MarcaPosibleDuplicado, VisibilidadVisible); err == nil {
			t.Fatal("se esperaba el fallo del registro")
		}
		// La pareja entera sigue como estaba: ni la ficha que pidió la operación
		// ni su par quedan resueltas (FR-022/M2).
		for _, ficha := range repo.episodios {
			if ficha.MarcaDuplicadoResueltaEn != nil {
				t.Errorf("la pareja no puede quedar resuelta sin su registro: %+v", ficha)
			}
		}
		if repo.episodios[1].Visibilidad != VisibilidadOculta {
			t.Errorf("la visibilidad de la pareja no se aplica sin auditoría: %q", repo.episodios[1].Visibilidad)
		}
	})

	t.Run("mostrar de nuevo desde no disponible", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{
			Titulo: "Caída", PublicadoEn: base, Visibilidad: VisibilidadOculta,
			MarcaDisponibilidad: ptrMarca(MarcaDisponibilidadNoDisponible),
		})
		service := serviceDePrueba(repo)

		out, err := service.ResolveMark(context.Background(), actor, repo.episodios[0].ID, MarcaNoDisponible, VisibilidadVisible)
		if err != nil {
			t.Fatalf("ResolveMark devolvió %v", err)
		}
		if out.Visibilidad != string(VisibilidadVisible) {
			t.Errorf("«Mostrar de nuevo» debe dejarla visible: %q", out.Visibilidad)
		}
		if out.MarcaDisponibilidadResueltaEn == nil || out.MarcaDisponibilidad != string(MarcaDisponibilidadNoDisponible) {
			t.Errorf("la marca queda cerrada (no borrada): %+v", out)
		}
	})

	t.Run("mantener visible cierra la marca de la ventana del feed", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{
			Titulo: "Fuera del feed", PublicadoEn: base,
			MarcaDisponibilidad: ptrMarca(MarcaDisponibilidadFueraDelFeed),
		})
		service := serviceDePrueba(repo)

		out, err := service.ResolveMark(context.Background(), actor, repo.episodios[0].ID, MarcaFueraDelFeed, VisibilidadVisible)
		if err != nil {
			t.Fatalf("ResolveMark devolvió %v", err)
		}
		if out.Visibilidad != string(VisibilidadVisible) || out.MarcaDisponibilidadResueltaEn == nil {
			t.Errorf("«Mantener visible» cierra la marca sin ocultar: %+v", out)
		}
	})

	t.Run("idempotente sobre una marca ya resuelta", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{
			Titulo: "Fuera del feed", PublicadoEn: base,
			MarcaDisponibilidad:           ptrMarca(MarcaDisponibilidadFueraDelFeed),
			MarcaDisponibilidadResueltaEn: ptrTime(base),
		})
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		out, err := service.ResolveMark(context.Background(), actor, id, MarcaFueraDelFeed, VisibilidadVisible)
		if err != nil {
			t.Fatalf("ResolveMark devolvió %v", err)
		}
		if !out.MarcaDisponibilidadResueltaEn.Equal(base) {
			t.Errorf("repetir la resolución no debe mover la fecha: %v", out.MarcaDisponibilidadResueltaEn)
		}
		if len(repo.updates) != 0 || len(repo.committed) != 0 {
			t.Errorf("una marca ya resuelta y sin cambio de visibilidad no se reescribe: %+v / %+v",
				repo.updates, repo.committed)
		}
	})

	t.Run("conflicto si la marca indicada no está activa", func(t *testing.T) {
		repo := newFakeRepository()
		// Los dos ejes activos a la vez: resolver el que NO está activo es conflicto.
		repo.seed(Episodio{
			Titulo: "Par y disponibilidad", PublicadoEn: base,
			DuplicadoDeID:       ptrUUID(uuid.New()),
			MarcaDisponibilidad: ptrMarca(MarcaDisponibilidadFueraDelFeed),
		})
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		_, err := service.ResolveMark(context.Background(), actor, id, MarcaNoDisponible, VisibilidadVisible)
		requireKind(t, err, apperr.KindConflict)

		// Sin ninguna marca de duplicado, resolver «posible duplicado» es conflicto.
		repo2 := newFakeRepository()
		repo2.seed(Episodio{Titulo: "Solo disponibilidad", PublicadoEn: base,
			MarcaDisponibilidad: ptrMarca(MarcaDisponibilidadFueraDelFeed)})
		service2 := serviceDePrueba(repo2)
		_, err = service2.ResolveMark(context.Background(), actor, repo2.episodios[0].ID, MarcaPosibleDuplicado, VisibilidadVisible)
		requireKind(t, err, apperr.KindConflict)
	})

	t.Run("validación y ficha inexistente", func(t *testing.T) {
		repo := newFakeRepository()
		repo.seed(Episodio{Titulo: "Ficha", PublicadoEn: base})
		service := serviceDePrueba(repo)
		id := repo.episodios[0].ID

		_, err := service.ResolveMark(context.Background(), actor, id, MarcaRevision("otra"), VisibilidadVisible)
		requireInvalid(t, err, "mark")
		_, err = service.ResolveMark(context.Background(), actor, id, MarcaNoDisponible, Visibilidad("draft"))
		requireInvalid(t, err, "visibility")
		_, err = service.ResolveMark(context.Background(), actor, uuid.New(), MarcaNoDisponible, VisibilidadVisible)
		requireKind(t, err, apperr.KindNotFound)
	})
}

// sembrarVideoYAudio crea el par real del cliente (el video de YouTube y el
// audio del podcast del mismo sermón) con sus dos procedencias y devuelve sus
// identificadores: el video primero, el podcast después.
func sembrarVideoYAudio(repo *fakeRepository, base time.Time) (uuid.UUID, uuid.UUID) {
	repo.seed(
		Episodio{
			Titulo:      "Potencial sin Resultados La Lección de la Higuera",
			PublicadoEn: base,
			Origen:      OrigenVideo,
			Caratula:    ptr("img_33333333-3333-3333-3333-333333333333.jpg"),
		},
		Episodio{
			Titulo:      "Potencial sin Resultados: La Lección de la Higuera",
			PublicadoEn: base,
			Origen:      OrigenAudio,
			Caratula:    ptr("img_44444444-4444-4444-4444-444444444444.jpg"),
		},
	)
	videoID, podcastID := repo.episodios[0].ID, repo.episodios[1].ID
	repo.seedFuentes(
		FuenteEpisodio{
			EpisodioID: videoID, Proveedor: ProveedorYouTube, IDExterno: "SimienteAA1",
			TituloFeed: "Potencial sin Resultados La Lección de la Higuera",
			Enlace:     "https://www.youtube.com/watch?v=SimienteAA1",
		},
		FuenteEpisodio{
			EpisodioID: podcastID, Proveedor: ProveedorPodcast, IDExterno: "guid-podcast",
			TituloFeed: "Potencial sin Resultados: La Lección de la Higuera",
			Enlace:     "https://podcasters.spotify.com/pod/show/simiente/episodes/abc",
		},
	)
	return videoID, podcastID
}

// ---------------------------------------------------------------------------
// Helpers de prueba del panel
// ---------------------------------------------------------------------------

// titulosAdmin imprime los títulos de una lista de fichas del panel.
func titulosAdmin(items []MedioAdmin) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Titulo)
	}
	return out
}
