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
// sincronización no la pisa) y los caracteres del español se conservan.
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
			{name: "descripción en español vacía", patch: MedioPatch{DescriptionEs: ptr("")}, campo: "descriptionEs"},
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
