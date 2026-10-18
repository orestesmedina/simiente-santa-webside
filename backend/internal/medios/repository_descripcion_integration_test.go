//go:build integration

package medios

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/paginate"
)

// Pruebas de integración de la DESCRIPCIÓN CORTA contra PostgreSQL real
// (T946/decisión del humano del 2026-10-18): vaciar la descripción desde el
// panel deja `description_es` en NULL —la ficha queda **sin descripción** y la
// línea se omite en la superficie pública (SC-010)—, la cadena vacía **no** cabe
// en la columna (el CHECK exige 1…300) y el vaciado pasa a ser del equipo
// (`description_edited`).
//
// La frontera es la misma que en el resto del módulo: aquí se verifica lo que el
// PATCH **persiste** (repository + PostgreSQL); la normalización `""` →
// «sin descripción» y la fila de auditoría `media.description.update` viven en el
// service y se fijan en sus pruebas unitarias (service_admin_test.go), porque las
// pruebas de integración del módulo no dependen de `users` (lo truncan en
// paralelo las pruebas del dominio usuarios: un actor real no es estable aquí).

// TestIntegrationUpdateMediaItemClearsTheDescription fija el efecto persistido
// del PATCH «vaciar la descripción» (FR-019/FR-020 + decisión del 2026-10-18):
// la fila queda con `description_es` NULL y `description_edited` en true, y la
// vía pública —el MISMO predicado que ve el visitante— devuelve la ficha con la
// descripción ausente (el service omite la línea, SC-010).
func TestIntegrationUpdateMediaItemClearsTheDescription(t *testing.T) {
	repo := newMediosRepo(t)
	ctx := context.Background()

	publicado := time.Date(2026, 7, 2, 6, 0, 0, 0, time.UTC)
	guardada, err := repo.InsertMediaItem(ctx, Episodio{
		Titulo:        "Sermón con descripción",
		PublicadoEn:   publicado,
		DescripcionEs: textOf("Una prédica del equipo"),
		Origen:        OrigenVideo,
		Visibilidad:   VisibilidadVisible,
	})
	if err != nil {
		t.Fatalf("InsertMediaItem: %v", err)
	}

	// Lo que el service escribe cuando el PATCH trae `descriptionEs: ""`: el
	// texto normalizado a «sin descripción» (NULL) y la marca del equipo
	// (`service.Update` fija `DescripcionEditada` cuando los datos cambian).
	vaciada := guardada
	vaciada.DescripcionEs = nil
	vaciada.DescripcionEditada = true
	if _, err := repo.UpdateMediaItem(ctx, vaciada); err != nil {
		t.Fatalf("UpdateMediaItem (vaciar): %v", err)
	}

	// La columna queda NULL y la descripción sigue siendo del equipo.
	var descripcion *string
	var editada bool
	if err := repo.pool.QueryRow(ctx,
		`SELECT description_es, description_edited FROM media_items WHERE id = $1`,
		guardada.ID).Scan(&descripcion, &editada); err != nil {
		t.Fatalf("leer la fila editada: %v", err)
	}
	if descripcion != nil {
		t.Errorf("description_es = %q, se esperaba NULL (sin descripción)", *descripcion)
	}
	if !editada {
		t.Error("description_edited debe quedar en true: el vaciado es del equipo")
	}

	// La vía pública devuelve la ficha con la descripción ausente: el service
	// compone el DTO sin la línea (SC-010).
	items, err := repo.ListPublishedMediaItems(ctx, paginate.Params{Limit: pageDefault})
	if err != nil {
		t.Fatalf("ListPublishedMediaItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("salieron %d fichas publicables, se esperaba 1", len(items))
	}
	if items[0].DescripcionEs != nil {
		t.Errorf("DescripcionEs = %q, se esperaba nil", *items[0].DescripcionEs)
	}
}

// TestIntegrationDescriptionEmptyStringIsRejected fija la última línea de
// defensa de la decisión: la columna NO admite la cadena vacía (el CHECK exige
// 1…300 caracteres), así que «sin descripción» se representa **solo** con NULL
// —igual que la traducción opcional (patrón I6)— y una escritura que se saltara
// la normalización del service sería rechazada por la base.
func TestIntegrationDescriptionEmptyStringIsRejected(t *testing.T) {
	ctx := context.Background()

	_, tx := testTx(t, ctx)

	// La fila se siembra dentro de la transacción de prueba (que se revierte):
	// el intento inválido aborta la transacción, así que va en un savepoint.
	id := uuid.NewString()
	err := insertRaw(ctx, tx, "media_items", map[string]any{
		"id": id, "title": "Sermón sin descripción", "published_at": time.Now(), "origin": "video",
	})
	if err != nil {
		t.Fatalf("sembrar la ficha: %v", err)
	}

	err = insertInSavepoint(ctx, tx, "media_items", map[string]any{
		"id": uuid.NewString(), "title": "Sermón con descripción vacía",
		"published_at": time.Now(), "origin": "video", "description_es": "",
	})
	if err == nil {
		t.Fatal("la base aceptó una descripción en español vacía")
	}
	if got := constraintName(err); got != "media_items_description_es_check" {
		t.Errorf("restricción = %q, se esperaba media_items_description_es_check", got)
	}

	// Con NULL (lo que escribe el PATCH al vaciar) la ficha es legal.
	var descripcion *string
	if err := tx.QueryRow(ctx, `SELECT description_es FROM media_items WHERE id = $1`, id).Scan(&descripcion); err != nil {
		t.Fatalf("leer la ficha sin descripción: %v", err)
	}
	if descripcion != nil {
		t.Errorf("description_es = %q, se esperaba NULL", *descripcion)
	}
}
