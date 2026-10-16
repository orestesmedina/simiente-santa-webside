//go:build integration

package donaciones

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración del dominio donaciones contra PostgreSQL real (F7,
// T709): la persistencia de `donation_methods` con sus invariantes protegidos
// por la BD —traducidos a errores del dominio—, el duplicado exacto normalizado
// por tipo y número, el orden de la sección y la escritura de la fila completa
// ya fusionada (R7-15).
//
// Requieren DATABASE_URL_TEST con las migraciones aplicadas (lo aporta el CI o
// el compose local; si no existe, se omiten solas). La tabla
// `donation_methods` es exclusiva de F7, así que cada prueba parte de la tabla
// vacía y la deja vacía al terminar.

// newDonationRepo construye el repositorio contra la base de pruebas y vacía las
// tablas de F7 para partir de un estado conocido (y las deja vacías al
// terminar). Las mutaciones se confirman (la transacción del dominio es real),
// así que se limpia por `DELETE`.
func newDonationRepo(t *testing.T) *repository {
	t.Helper()

	ctx := context.Background()
	pool := testutil.Pool(t, ctx)
	for _, table := range []string{"donation_methods", "donation_texts"} {
		if _, err := pool.Exec(ctx, "DELETE FROM "+table); err != nil {
			t.Fatalf("vaciar %s: %v", table, err)
		}
	}
	t.Cleanup(func() {
		for _, table := range []string{"donation_methods", "donation_texts"} {
			if _, err := pool.Exec(context.Background(), "DELETE FROM "+table); err != nil {
				t.Logf("no se pudo vaciar %s al terminar la prueba: %v", table, err)
			}
		}
	})
	return NewRepository(pool)
}

// baseIban arma un método IBAN válido y completo: es el punto de partida de las
// pruebas, que cambian un solo aspecto cada vez.
func baseIban() MetodoDonacion {
	return MetodoDonacion{
		Kind:             KindIban,
		Number:           "CR05015202001234567890",
		Institution:      "Banco Nacional",
		HolderEs:         textOf("Iglesia Simiente Santa"),
		HolderEn:         textOf("Simiente Santa Church"),
		Currency:         currencyOf(CurrencyCRC),
		PublicationState: StatePublished,
		SortOrder:        1,
	}
}

// baseSinpe arma un método SINPE válido (sin moneda: Q2/FR-014).
func baseSinpe() MetodoDonacion {
	return MetodoDonacion{
		Kind:             KindSinpe,
		Number:           "+50688888888",
		Institution:      "Banco Nacional",
		PublicationState: StatePublished,
		SortOrder:        1,
	}
}

// seedMetodo inserta un método por el repositorio y devuelve la fila guardada;
// si la inserción falla, la prueba se detiene.
func seedMetodo(t *testing.T, repo *repository, method MetodoDonacion) MetodoDonacion {
	t.Helper()
	saved, err := repo.InsertDonationMethod(context.Background(), method)
	if err != nil {
		t.Fatalf("insertar método %q: %v", method.Number, err)
	}
	return saved
}

// countMetodos cuenta las filas de `donation_methods` con SQL directo
// (independiente del código bajo prueba) para comprobar que nada se guardó a
// medias.
func countMetodos(t *testing.T, repo *repository) int {
	t.Helper()
	var total int
	if err := repo.pool.QueryRow(context.Background(), `SELECT count(*) FROM donation_methods`).Scan(&total); err != nil {
		t.Fatalf("contar filas de donation_methods: %v", err)
	}
	return total
}

// TestIntegrationInsertDonationMethodPersistsEveryColumn comprueba que el alta
// escribe y devuelve la fila completa (FR-010/FR-019).
func TestIntegrationInsertDonationMethodPersistsEveryColumn(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	in := baseIban()
	saved := seedMetodo(t, repo, in)

	if saved.ID == uuid.Nil {
		t.Fatal("el alta no devolvió id")
	}
	if saved.CreatedAt.IsZero() || saved.UpdatedAt.IsZero() {
		t.Errorf("el alta no devolvió created_at/updated_at: %v / %v", saved.CreatedAt, saved.UpdatedAt)
	}
	assertMetodoEquals(t, "alta IBAN", in, saved)

	got, found, err := repo.GetDonationMethodByID(ctx, saved.ID)
	if err != nil {
		t.Fatalf("GetDonationMethodByID: %v", err)
	}
	if !found {
		t.Fatal("GetDonationMethodByID no encontró el método recién creado")
	}
	assertMetodoEquals(t, "detalle del panel", saved, got)

	if _, found, err := repo.GetDonationMethodByID(ctx, uuid.New()); err != nil || found {
		t.Errorf("GetDonationMethodByID(inexistente) = (found=%v, err=%v), se esperaba (false, nil)", found, err)
	}
}

// TestIntegrationInsertDonationMethodRejectsUnnormalizedNumbers fija el CHECK de
// formato por tipo de la BD (FR-015/FR-016/SC-009): el número se guarda
// normalizado —un IBAN con espacios o en minúsculas y un SINPE sin `+506`
// fallan— y el normalizado procede.
func TestIntegrationInsertDonationMethodRejectsUnnormalizedNumbers(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*MetodoDonacion)
		field  string
	}{
		{
			name:   "IBAN con espacios",
			mutate: func(m *MetodoDonacion) { m.Number = "CR05 0152 0200 1234 5678 90" },
			field:  "number",
		},
		{
			name:   "IBAN en minúsculas",
			mutate: func(m *MetodoDonacion) { m.Number = "cr05015202001234567890" },
			field:  "number",
		},
		{
			name:   "SINPE sin +506",
			mutate: func(m *MetodoDonacion) { *m = baseSinpe(); m.Number = "88888888" },
			field:  "number",
		},
		{
			name:   "SINPE con guiones",
			mutate: func(m *MetodoDonacion) { *m = baseSinpe(); m.Number = "+506-8888-8888" },
			field:  "number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newDonationRepo(t)
			in := baseIban()
			tt.mutate(&in)

			_, err := repo.InsertDonationMethod(context.Background(), in)
			if err == nil {
				t.Fatal("la BD aceptó un número sin normalizar")
			}
			if kind := errorKind(err); kind != apperr.KindInvalid {
				t.Fatalf("kind = %v (%v), se esperaba invalid", kind, err)
			}
			details := errorDetails(err)
			if details == nil || details[tt.field] == nil {
				t.Fatalf("details = %v, se esperaba el campo %q", details, tt.field)
			}
			if got := countMetodos(t, repo); got != 0 {
				t.Errorf("se guardó algo a medias: %d filas", got)
			}
		})
	}

	// El número ya normalizado procede en ambos tipos.
	repo := newDonationRepo(t)
	iban := seedMetodo(t, repo, baseIban())
	if iban.Number != "CR05015202001234567890" {
		t.Errorf("IBAN guardado = %q", iban.Number)
	}
	sinpe := seedMetodo(t, repo, baseSinpe())
	if sinpe.Number != "+50688888888" {
		t.Errorf("SINPE guardado = %q", sinpe.Number)
	}
}

// TestIntegrationInsertDonationMethodRejectsCurrencyMismatch fija el CHECK de
// moneda por tipo (Q2/FR-014): obligatoria en IBAN y prohibida en SINPE.
func TestIntegrationInsertDonationMethodRejectsCurrencyMismatch(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	withoutCurrency := baseIban()
	withoutCurrency.Currency = nil
	if _, err := repo.InsertDonationMethod(ctx, withoutCurrency); err == nil {
		t.Error("la BD aceptó un IBAN sin moneda")
	} else if kind := errorKind(err); kind != apperr.KindInvalid {
		t.Errorf("IBAN sin moneda: kind = %v, se esperaba invalid", kind)
	} else if details := errorDetails(err); details == nil || details["currency"] == nil {
		t.Errorf("IBAN sin moneda: details = %v, se esperaba currency", details)
	}

	sinpeWithCurrency := baseSinpe()
	sinpeWithCurrency.Currency = currencyOf(CurrencyCRC)
	if _, err := repo.InsertDonationMethod(ctx, sinpeWithCurrency); err == nil {
		t.Error("la BD aceptó un SINPE con moneda")
	} else if details := errorDetails(err); details == nil || details["currency"] == nil {
		t.Errorf("SINPE con moneda: details = %v, se esperaba currency", details)
	}

	// IBAN + CRC y SINPE sin moneda proceden.
	if _, err := repo.InsertDonationMethod(ctx, baseIban()); err != nil {
		t.Errorf("la BD rechazó un IBAN con CRC: %v", err)
	}
	if _, err := repo.InsertDonationMethod(ctx, baseSinpe()); err != nil {
		t.Errorf("la BD rechazó un SINPE sin moneda: %v", err)
	}
}

// TestIntegrationInsertDonationMethodRejectsHolderEnWithoutEs fija el CHECK de
// español por campo (FR-007/Decisión 8): `holder_en` exige `holder_es`.
func TestIntegrationInsertDonationMethodRejectsHolderEnWithoutEs(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	in := baseIban()
	in.HolderEs = nil
	in.HolderEn = textOf("Simiente Santa Church")
	if _, err := repo.InsertDonationMethod(ctx, in); err == nil {
		t.Fatal("la BD aceptó holderEn sin holderEs")
	} else if details := errorDetails(err); details == nil || details["holderEs"] == nil {
		t.Fatalf("details = %v, se esperaba holderEs", details)
	}

	// Con el español presente procede.
	if _, err := repo.InsertDonationMethod(ctx, baseIban()); err != nil {
		t.Errorf("la BD rechazó holderEn con holderEs: %v", err)
	}
}

// TestIntegrationInsertDonationMethodRejectsUntrimmedText fija el CHECK de
// `btrim` (normalización de F2/P13): los textos cortos se guardan ya recortados.
func TestIntegrationInsertDonationMethodRejectsUntrimmedText(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*MetodoDonacion)
		field  string
	}{
		{
			name:   "entidad con espacios sobrantes",
			mutate: func(m *MetodoDonacion) { m.Institution = " Banco Nacional " },
			field:  "institution",
		},
		{
			name:   "titular español con espacios sobrantes",
			mutate: func(m *MetodoDonacion) { m.HolderEs = textOf(" Iglesia Simiente Santa ") },
			field:  "holderEs",
		},
		{
			name: "titular inglés con espacios sobrantes",
			mutate: func(m *MetodoDonacion) {
				m.HolderEn = textOf(" Simiente Santa Church ")
			},
			field: "holderEn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newDonationRepo(t)
			in := baseIban()
			tt.mutate(&in)

			_, err := repo.InsertDonationMethod(context.Background(), in)
			if err == nil {
				t.Fatal("la BD aceptó un texto sin recortar")
			}
			if details := errorDetails(err); details == nil || details[tt.field] == nil {
				t.Fatalf("details = %v, se esperaba el campo %q", details, tt.field)
			}
		})
	}
}

// TestIntegrationExistsDonationMethodDuplicateNormalizes fija el duplicado exacto
// por tipo y número normalizado (R7-9/P7-10): se detecta tras normalizar la
// variante de escritura, no dispara por el mismo número en otro tipo y respeta
// `exclude_id` (editar sin cambiar el número no choca contra sí mismo, pero sí
// contra otro método).
func TestIntegrationExistsDonationMethodDuplicateNormalizes(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	saved := seedMetodo(t, repo, baseIban())
	other := baseIban()
	other.Number = "CR05015202001234567891"
	savedOther := seedMetodo(t, repo, other)

	exists := func(t *testing.T, kind MetodoKind, number string, exclude *uuid.UUID) bool {
		t.Helper()
		duplicate, err := repo.ExistsDonationMethodDuplicate(ctx, kind, number, exclude)
		if err != nil {
			t.Fatalf("ExistsDonationMethodDuplicate: %v", err)
		}
		return duplicate
	}

	// La variante de escritura se normaliza al MISMO valor guardado.
	normalized, err := NormalizarIban("cr05 0152-0200-1234-5678-90")
	if err != nil {
		t.Fatalf("NormalizarIban: %v", err)
	}
	if !exists(t, KindIban, normalized, nil) {
		t.Error("no se detectó el duplicado normalizado («cr05 0152-…» vs «CR050152…»)")
	}

	// `exclude_id` = la propia fila → no es duplicado de sí misma.
	if exists(t, KindIban, saved.Number, &saved.ID) {
		t.Error("el método se detectó a sí mismo siendo excluido")
	}
	// `exclude_id` = otra fila, con el número de esa otra fila → sí choca (el
	// `PATCH` que cambia el número al de otro método debe devolver 409).
	if !exists(t, KindIban, savedOther.Number, &saved.ID) {
		t.Error("no se detectó el duplicado contra OTRO método al excluir la propia fila")
	}
	// Un número distinto no es duplicado.
	if exists(t, KindIban, "ES9121000418450200051332", nil) {
		t.Error("un número distinto no es duplicado")
	}
	// El mismo número «en otro tipo» no dispara (la consulta es tipo + número).
	if exists(t, KindSinpe, saved.Number, nil) {
		t.Error("el mismo número en otro tipo no es duplicado")
	}
}

// TestIntegrationListDonationMethodsOrdersAndFilters fija el listado del panel
// (P7-20/US2 esc. 10): el MISMO orden que ve el visitante (iban primero y luego
// sinpe, cada grupo por `sort_order`, `id`), el filtro opcional por estado, el
// total del sobre y la paginación.
func TestIntegrationListDonationMethodsOrdersAndFilters(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	ibanTwo := baseIban()
	ibanTwo.Number = "CR05015202001234567892"
	ibanTwo.SortOrder = 2
	savedIbanTwo := seedMetodo(t, repo, ibanTwo)

	sinpeOne := baseSinpe()
	sinpeOne.SortOrder = 1
	savedSinpeOne := seedMetodo(t, repo, sinpeOne)

	ibanOne := baseIban()
	ibanOne.SortOrder = 1
	savedIbanOne := seedMetodo(t, repo, ibanOne)

	draft := baseSinpe()
	draft.Number = "+50677777777"
	draft.PublicationState = StateDraft
	draft.SortOrder = 0
	savedDraft := seedMetodo(t, repo, draft)

	// Panel completo: iban (sort_order 1, 2) y luego sinpe.
	items, err := repo.ListDonationMethods(ctx, nil, paginate.Params{Limit: 50})
	if err != nil {
		t.Fatalf("ListDonationMethods: %v", err)
	}
	if len(items) != 4 {
		t.Fatalf("el panel devolvió %d métodos, se esperaban 4", len(items))
	}
	wantOrder := []uuid.UUID{savedIbanOne.ID, savedIbanTwo.ID, savedDraft.ID, savedSinpeOne.ID}
	for i, want := range wantOrder {
		if items[i].ID != want {
			t.Fatalf("orden del panel[%d] = %v, se esperaba %v (CASE kind, sort_order, id)", i, items[i].ID, want)
		}
	}

	// Filtro por estado y total del sobre.
	published := StatePublished
	publishedItems, err := repo.ListDonationMethods(ctx, &published, paginate.Params{Limit: 50})
	if err != nil {
		t.Fatalf("ListDonationMethods(published): %v", err)
	}
	if len(publishedItems) != 3 {
		t.Errorf("publicados = %d, se esperaban 3", len(publishedItems))
	}
	total, err := repo.CountDonationMethods(ctx, &published)
	if err != nil {
		t.Fatalf("CountDonationMethods: %v", err)
	}
	if total != 3 {
		t.Errorf("total publicado = %d, se esperaba 3", total)
	}
	all, err := repo.CountDonationMethods(ctx, nil)
	if err != nil {
		t.Fatalf("CountDonationMethods(sin filtro): %v", err)
	}
	if all != 4 {
		t.Errorf("total sin filtro = %d, se esperaba 4", all)
	}

	// Paginación: el offset respeta el mismo orden.
	page, err := repo.ListDonationMethods(ctx, nil, paginate.Params{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("ListDonationMethods(paginado): %v", err)
	}
	if len(page) != 1 || page[0].ID != savedIbanTwo.ID {
		t.Errorf("segunda página = %+v, se esperaba el segundo del orden", page)
	}
}

// TestIntegrationUpdateDonationMethodWritesTheMergedRow fija R7-15: el `PATCH`
// fusiona los campos presentes sobre la fila actual y la transacción escribe la
// fila COMPLETA ya fusionada (los ausentes conservan su valor y los opcionales
// borrados viajan como NULL).
func TestIntegrationUpdateDonationMethodWritesTheMergedRow(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	saved := seedMetodo(t, repo, baseIban())

	merged := saved
	merged.Institution = "Banco de Costa Rica"
	merged.SortOrder = 7
	merged.HolderEn = nil // null explícito borra el titular inglés
	merged.Currency = currencyOf(CurrencyUSD)

	updated, err := repo.UpdateDonationMethod(ctx, merged)
	if err != nil {
		t.Fatalf("UpdateDonationMethod: %v", err)
	}
	assertMetodoEquals(t, "update fusionado", merged, updated)

	got, found, err := repo.GetDonationMethodByID(ctx, saved.ID)
	if err != nil || !found {
		t.Fatalf("GetDonationMethodByID = (found=%v, err=%v)", found, err)
	}
	assertMetodoEquals(t, "fila persistida", merged, got)
	if got.HolderEs == nil || *got.HolderEs != *saved.HolderEs {
		t.Errorf("un campo ausente del PATCH cambió: HolderEs = %v", got.HolderEs)
	}
	if got.Number != saved.Number || got.Kind != saved.Kind {
		t.Errorf("un campo ausente del PATCH cambió: %+v", got)
	}
	if got.CreatedAt.IsZero() || !got.CreatedAt.Equal(saved.CreatedAt) {
		t.Errorf("el update cambió la identidad de la fila: %v / %v", got.CreatedAt, saved.CreatedAt)
	}

	// Un id inexistente es "no existe" (404 del handler).
	missing := merged
	missing.ID = uuid.New()
	if _, err := repo.UpdateDonationMethod(ctx, missing); errorKind(err) != apperr.KindNotFound {
		t.Errorf("UpdateDonationMethod(inexistente) = %v, se esperaba not_found", err)
	}
}

// TestIntegrationDeleteDonationMethodRemovesTheRow fija el borrado físico (sin
// `deleted_at`).
func TestIntegrationDeleteDonationMethodRemovesTheRow(t *testing.T) {
	repo := newDonationRepo(t)
	ctx := context.Background()

	saved := seedMetodo(t, repo, baseIban())

	deleted, err := repo.DeleteDonationMethod(ctx, saved.ID)
	if err != nil {
		t.Fatalf("DeleteDonationMethod: %v", err)
	}
	if !deleted {
		t.Error("DeleteDonationMethod no reportó el borrado de una fila existente")
	}
	if got := countMetodos(t, repo); got != 0 {
		t.Errorf("quedaron %d filas tras el borrado", got)
	}

	deleted, err = repo.DeleteDonationMethod(ctx, uuid.New())
	if err != nil {
		t.Fatalf("DeleteDonationMethod(inexistente): %v", err)
	}
	if deleted {
		t.Error("DeleteDonationMethod dijo haber borrado un id inexistente")
	}
}

// assertMetodoEquals compara la entidad esperada con la obtenida campo por campo
// (los sellos de tiempo se comprueban aparte: el alta los genera la base).
func assertMetodoEquals(t *testing.T, context string, want, got MetodoDonacion) {
	t.Helper()

	if got.Kind != want.Kind || got.Number != want.Number || got.Institution != want.Institution {
		t.Errorf("%s: tipo/número/entidad = (%q, %q, %q), se esperaba (%q, %q, %q)",
			context, got.Kind, got.Number, got.Institution, want.Kind, want.Number, want.Institution)
	}
	if !sameString(got.HolderEs, want.HolderEs) || !sameString(got.HolderEn, want.HolderEn) {
		t.Errorf("%s: titular = (%v, %v), se esperaba (%v, %v)",
			context, got.HolderEs, got.HolderEn, want.HolderEs, want.HolderEn)
	}
	if !sameCurrency(got.Currency, want.Currency) {
		t.Errorf("%s: moneda = %v, se esperaba %v", context, got.Currency, want.Currency)
	}
	if got.PublicationState != want.PublicationState {
		t.Errorf("%s: estado = %q, se esperaba %q", context, got.PublicationState, want.PublicationState)
	}
	if got.SortOrder != want.SortOrder {
		t.Errorf("%s: sortOrder = %d, se esperaba %d", context, got.SortOrder, want.SortOrder)
	}
}

// sameString compara dos cadenas anulables.
func sameString(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// sameCurrency compara dos monedas anulables.
func sameCurrency(a, b *Currency) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
