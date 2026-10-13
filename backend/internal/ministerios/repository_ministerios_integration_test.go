//go:build integration

package ministerios

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	gendb "simiente-santa/backend/internal/db"
	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/audit"
	"simiente-santa/backend/internal/platform/paginate"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración del dominio ministerios contra PostgreSQL real (F6,
// T608): la persistencia de `ministries` con sus invariantes protegidos por la
// BD —traducidos a errores del dominio—, el duplicado exacto normalizado, el
// aviso de enlace compartido recalculado en cada lectura, la escritura de la
// fila completa ya fusionada (A6-3/P6-20) y el ciclo de vida del archivo
// (A6-11).
//
// Requieren DATABASE_URL_TEST con las migraciones aplicadas (lo aporta el CI o
// el compose local; si no existe, se omiten solas). La tabla `ministries` es
// exclusiva de F6, así que cada prueba parte de la tabla vacía y la deja vacía
// al terminar: las mutaciones se confirman (la transacción del dominio es real),
// así que se limpia por `DELETE` (que pide un bloqueo compatible con las
// transacciones abiertas por las pruebas de integración de otros dominios, que
// solo leen `ministries`).

// testToday/testNowTime son el corte FIJO de la consulta compartida
// `IsFilePublished` (nunca el reloj real) que usan las pruebas de archivos.
const (
	testToday   = "2026-10-13"
	testNowTime = "12:00"
)

// newMinistryRepo construye el repositorio contra la base de pruebas y vacía
// `ministries` para partir de un estado conocido (y lo deja vacío al terminar).
func newMinistryRepo(t *testing.T) *repository {
	t.Helper()

	ctx := context.Background()
	pool := testutil.Pool(t, ctx)
	if _, err := pool.Exec(ctx, `DELETE FROM ministries`); err != nil {
		t.Fatalf("vaciar ministries: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM ministries`); err != nil {
			t.Logf("no se pudo vaciar ministries al terminar la prueba: %v", err)
		}
	})
	return NewRepository(pool)
}

// ptr devuelve un *string con el valor indicado.
func ptr(s string) *string { return &s }

// uniqueImageFile genera un nombre de archivo irrepetible: `IsFilePublished`
// mira TODAS las tablas que referencian archivos y otras pruebas de integración
// (agenda, portada) pueden dejar filas confirmadas con sus propios nombres, así
// que una prueba sobre archivos nunca debe reutilizar uno ajeno.
func uniqueImageFile() string {
	return "img_" + uuid.NewString() + ".jpg"
}

// baseMinistry arma un ministerio válido y completo: es el punto de partida de
// las pruebas, que cambian un solo aspecto cada vez.
func baseMinistry(name string) Ministry {
	return Ministry{
		NameEs:           name,
		NameEn:           ptr("Worship"),
		DescriptionEs:    "Grupo de alabanza de la iglesia",
		DescriptionEn:    ptr("Church worship team"),
		LeaderNameEs:     ptr("Ana Pérez"),
		LeaderNameEn:     ptr("Ana Perez"),
		LeaderPhone:      ptr("+584120000000"),
		LeaderVisible:    true,
		WhatsappURL:      ptr("https://chat.whatsapp.com/AbCdEf123"),
		ScheduleEs:       ptr("Todos los viernes, 7:00 p. m."),
		ScheduleEn:       ptr("Fridays, 7:00 pm"),
		PlaceEs:          ptr("Templo"),
		PlaceEn:          ptr("Temple"),
		ImageFile:        ptr("img_11111111-1111-1111-1111-111111111111.jpg"),
		ImageAltEs:       ptr("Grupo de alabanza"),
		ImageAltEn:       ptr("Worship team"),
		PublicationState: StatePublished,
		SortOrder:        1,
	}
}

// seedMinistry inserta un ministerio por el repositorio y devuelve la fila
// guardada; si la inserción falla, la prueba se detiene.
func seedMinistry(t *testing.T, repo *repository, ministry Ministry) Ministry {
	t.Helper()
	saved, err := repo.InsertMinistry(context.Background(), ministry)
	if err != nil {
		t.Fatalf("insertar ministerio %q: %v", ministry.NameEs, err)
	}
	return saved
}

// countRows cuenta las filas de `ministries` con SQL directo (independiente del
// código bajo prueba) para comprobar que nada se guardó a medias.
func countRows(t *testing.T, repo *repository) int {
	t.Helper()
	var total int
	if err := repo.pool.QueryRow(context.Background(), `SELECT count(*) FROM ministries`).Scan(&total); err != nil {
		t.Fatalf("contar filas de ministries: %v", err)
	}
	return total
}

// TestIntegrationInsertMinistryPersistsEveryColumn comprueba que el alta escribe
// y devuelve la fila completa: los dos idiomas, la ficha del encargado (que el
// panel SÍ ve aunque no esté publicada), el enlace, horario, lugar, imagen con
// alt, estado y orden (FR-010/FR-019).
func TestIntegrationInsertMinistryPersistsEveryColumn(t *testing.T) {
	repo := newMinistryRepo(t)
	ctx := context.Background()

	in := baseMinistry("Alabanza")
	in.LeaderVisible = false // dato guardado ≠ dato publicado: el panel lo ve igual
	saved := seedMinistry(t, repo, in)

	if saved.ID == uuid.Nil {
		t.Fatal("el alta no devolvió id")
	}
	if saved.CreatedAt.IsZero() || saved.UpdatedAt.IsZero() {
		t.Errorf("el alta no devolvió created_at/updated_at: %v / %v", saved.CreatedAt, saved.UpdatedAt)
	}
	assertMinistryEquals(t, "alta", in, saved)

	// El detalle del panel devuelve exactamente lo guardado, con la ficha del
	// encargado íntegra y sin enlace compartido (no hay otro igual).
	got, found, err := repo.GetMinistryByID(ctx, saved.ID)
	if err != nil {
		t.Fatalf("GetMinistryByID: %v", err)
	}
	if !found {
		t.Fatal("GetMinistryByID no encontró el ministerio recién creado")
	}
	assertMinistryEquals(t, "detalle del panel", saved, got.Ministry)
	if got.WhatsappURLShared {
		t.Error("whatsappUrlShared = true sin otra fila con el mismo enlace")
	}

	// Un id inexistente no es un error: el handler responde 404.
	if _, found, err := repo.GetMinistryByID(ctx, uuid.New()); err != nil || found {
		t.Errorf("GetMinistryByID(inexistente) = (found=%v, err=%v), se esperaba (false, nil)", found, err)
	}
}

// TestIntegrationInsertMinistryRejectsInvalidRows fija los invariantes de la
// tabla (FR-014/FR-015/FR-020): la BD es la última línea de defensa y cada
// violación se traduce a un 400 con el campo que falla, sin guardar nada.
func TestIntegrationInsertMinistryRejectsInvalidRows(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Ministry)
		field  string
	}{
		{
			name:   "celular y correo juntos",
			mutate: func(m *Ministry) { m.LeaderEmail = ptr("ana@example.com") },
			field:  "leaderEmail",
		},
		{
			name:   "contacto sin nombre del encargado",
			mutate: func(m *Ministry) { m.LeaderNameEs = nil },
			field:  "leaderNameEs",
		},
		{
			name:   "nombre en inglés sin español",
			mutate: func(m *Ministry) { m.NameEs = "  " },
			field:  "nameEs",
		},
		{
			name:   "horario en inglés sin español",
			mutate: func(m *Ministry) { m.ScheduleEs = nil },
			field:  "scheduleEs",
		},
		{
			name:   "lugar en inglés sin español",
			mutate: func(m *Ministry) { m.PlaceEs = nil },
			field:  "placeEs",
		},
		{
			name:   "encargado en inglés sin español",
			mutate: func(m *Ministry) { m.LeaderNameEs = nil; m.LeaderPhone = nil },
			field:  "leaderNameEs",
		},
		{
			name:   "texto alternativo en inglés sin español",
			mutate: func(m *Ministry) { m.ImageAltEs = nil },
			field:  "imageAltEs",
		},
		{
			name:   "imagen sin texto alternativo",
			mutate: func(m *Ministry) { m.ImageAltEs = nil; m.ImageAltEn = nil },
			field:  "imageAltEs",
		},
		{
			name:   "enlace http",
			mutate: func(m *Ministry) { m.WhatsappURL = ptr("http://chat.whatsapp.com/AbCdEf123") },
			field:  "whatsappUrl",
		},
		{
			name:   "enlace javascript",
			mutate: func(m *Ministry) { m.WhatsappURL = ptr("javascript:alert(1)") },
			field:  "whatsappUrl",
		},
		{
			name:   "enlace con espacio",
			mutate: func(m *Ministry) { m.WhatsappURL = ptr("https://chat.whatsapp.com/Ab Cd") },
			field:  "whatsappUrl",
		},
		{
			name:   "celular con espacios sobrantes",
			mutate: func(m *Ministry) { m.LeaderPhone = ptr(" +584120000000 ") },
			field:  "leaderPhone",
		},
		{
			name:   "nombre de más de 160 caracteres",
			mutate: func(m *Ministry) { m.NameEs = strings.Repeat("x", 161) },
			field:  "nameEs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMinistryRepo(t)
			in := baseMinistry("Alabanza")
			tt.mutate(&in)

			_, err := repo.InsertMinistry(context.Background(), in)
			if err == nil {
				t.Fatal("la BD aceptó una fila inválida")
			}
			if kind := errorKind(err); kind != apperr.KindInvalid {
				t.Fatalf("kind = %v (%v), se esperaba invalid", kind, err)
			}
			details := errorDetails(err)
			if details == nil || details[tt.field] == nil {
				t.Fatalf("details = %v, se esperaba el campo %q", details, tt.field)
			}
			if got := countRows(t, repo); got != 0 {
				t.Errorf("se guardó algo a medias: %d filas", got)
			}
		})
	}
}

// TestIntegrationInsertMinistryAcceptsAnyHTTPSHost fija Q3: el enlace solo exige
// `https` seguro, sin lista blanca de dominios, y se guarda tal cual (sin
// transformarlo).
func TestIntegrationInsertMinistryAcceptsAnyHTTPSHost(t *testing.T) {
	repo := newMinistryRepo(t)

	for _, url := range []string{
		"https://chat.whatsapp.com/AbCdEf123",
		"https://ejemplo.com/grupo",
	} {
		in := baseMinistry("Ministerio " + url)
		in.WhatsappURL = ptr(url)

		saved := seedMinistry(t, repo, in)
		if saved.WhatsappURL == nil || *saved.WhatsappURL != url {
			t.Errorf("el enlace guardado = %v, se esperaba %q tal cual", saved.WhatsappURL, url)
		}
	}
}

// TestIntegrationExistsMinistryDuplicateNormalizes fija el duplicado exacto
// (R6-11): se detecta con el texto normalizado (`btrim` + minúsculas) y NO se
// dispara por compartir el enlace (eso es un aviso, no un duplicado).
func TestIntegrationExistsMinistryDuplicateNormalizes(t *testing.T) {
	repo := newMinistryRepo(t)
	ctx := context.Background()

	in := baseMinistry("Alabanza")
	saved := seedMinistry(t, repo, in)

	exists := func(t *testing.T, name string, whatsapp *string, exclude *uuid.UUID) bool {
		t.Helper()
		duplicate, err := repo.ExistsMinistryDuplicate(ctx, name, in.DescriptionEs, in.LeaderNameEs, whatsapp, exclude)
		if err != nil {
			t.Fatalf("ExistsMinistryDuplicate: %v", err)
		}
		return duplicate
	}

	if !exists(t, "  alabanza  ", in.WhatsappURL, nil) {
		t.Error("no se detectó el duplicado normalizado («  alabanza  » vs «Alabanza»)")
	}
	if !exists(t, "ALABANZA", in.WhatsappURL, nil) {
		t.Error("no se detectó el duplicado en mayúsculas")
	}
	if exists(t, "Alabanza", in.WhatsappURL, &saved.ID) {
		t.Error("el ministerio se detectó a sí mismo siendo excluido")
	}
	if exists(t, "Alabanza B", in.WhatsappURL, nil) {
		t.Error("dos nombres distintos no son duplicados")
	}
	if exists(t, "Alabanza", ptr("https://ejemplo.com/otro"), nil) {
		t.Error("un enlace distinto no es duplicado")
	}
	// Compartir el enlace con datos distintos NO es duplicado: es el aviso no
	// bloqueante `whatsappUrlShared` (R6-11). El enlace forma parte de la
	// comparación, así que una fila con enlace no duplica a una sin él.
	if exists(t, "Alabanza", nil, nil) {
		t.Error("una fila sin enlace no es el duplicado de una que sí lo tiene")
	}
}

// TestIntegrationListMinistriesOrdersAndSharesLink fija el listado del panel
// (P6-19): el MISMO orden que ve el visitante (`sort_order`, `id`), el filtro
// opcional por estado, el total del sobre y el aviso de enlace compartido
// recalculado en cada lectura (R6-11).
func TestIntegrationListMinistriesOrdersAndSharesLink(t *testing.T) {
	repo := newMinistryRepo(t)
	ctx := context.Background()

	const shared = "https://chat.whatsapp.com/Compartido"
	first := baseMinistry("Alabanza")
	first.SortOrder = 2
	first.WhatsappURL = ptr(shared)
	savedFirst := seedMinistry(t, repo, first)

	second := baseMinistry("Intercesión")
	second.SortOrder = 1
	second.PublicationState = StateDraft
	second.WhatsappURL = ptr(shared)
	savedSecond := seedMinistry(t, repo, second)

	third := baseMinistry("Niños")
	third.SortOrder = 3
	third.WhatsappURL = nil
	third.LeaderPhone = nil
	third.LeaderNameEs = nil
	third.LeaderNameEn = nil
	third.LeaderVisible = false
	savedThird := seedMinistry(t, repo, third)

	// Panel completo: borradores incluidos, en el orden del módulo.
	items, err := repo.ListMinistries(ctx, nil, paginate.Params{Limit: 50})
	if err != nil {
		t.Fatalf("ListMinistries: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("el panel devolvió %d ministerios, se esperaban 3", len(items))
	}
	wantOrder := []uuid.UUID{savedSecond.ID, savedFirst.ID, savedThird.ID}
	for i, want := range wantOrder {
		if items[i].ID != want {
			t.Fatalf("orden del panel = %v, se esperaba %v (sort_order, id)", items[i].ID, want)
		}
	}
	if !items[0].WhatsappURLShared || !items[1].WhatsappURLShared {
		t.Error("las dos filas que comparten enlace deben señalarlo")
	}
	if items[2].WhatsappURLShared {
		t.Error("la fila sin enlace no puede señalar un enlace compartido")
	}

	// Filtro por estado y total del sobre.
	published := StatePublished
	publishedItems, err := repo.ListMinistries(ctx, &published, paginate.Params{Limit: 50})
	if err != nil {
		t.Fatalf("ListMinistries(published): %v", err)
	}
	if len(publishedItems) != 2 {
		t.Errorf("publicados = %d, se esperaban 2", len(publishedItems))
	}
	total, err := repo.CountMinistries(ctx, &published)
	if err != nil {
		t.Fatalf("CountMinistries: %v", err)
	}
	if total != 2 {
		t.Errorf("total publicado = %d, se esperaba 2", total)
	}
	all, err := repo.CountMinistries(ctx, nil)
	if err != nil {
		t.Fatalf("CountMinistries(sin filtro): %v", err)
	}
	if all != 3 {
		t.Errorf("total sin filtro = %d, se esperaba 3", all)
	}

	// Paginación: el offset respeta el mismo orden.
	page, err := repo.ListMinistries(ctx, nil, paginate.Params{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("ListMinistries(paginado): %v", err)
	}
	if len(page) != 1 || page[0].ID != savedFirst.ID {
		t.Errorf("segunda página = %+v, se esperaba el segundo del orden", page)
	}

	// El aviso se recalcula: al quitar el enlace de una fila, la otra deja de
	// compartirlo.
	merged := savedSecond
	merged.WhatsappURL = nil
	if _, err := repo.UpdateMinistry(ctx, merged); err != nil {
		t.Fatalf("UpdateMinistry: %v", err)
	}
	detail, found, err := repo.GetMinistryByID(ctx, savedFirst.ID)
	if err != nil || !found {
		t.Fatalf("GetMinistryByID = (found=%v, err=%v)", found, err)
	}
	if detail.WhatsappURLShared {
		t.Error("el aviso no se recalculó tras cambiar el enlace de la otra fila")
	}
}

// TestIntegrationUpdateMinistryWritesTheMergedRow fija A6-3/P6-20: el `PATCH`
// fusiona los campos presentes sobre la fila actual y la transacción escribe la
// fila COMPLETA ya fusionada (los ausentes conservan su valor y los opcionales
// borrados viajan como NULL).
func TestIntegrationUpdateMinistryWritesTheMergedRow(t *testing.T) {
	repo := newMinistryRepo(t)
	ctx := context.Background()

	saved := seedMinistry(t, repo, baseMinistry("Alabanza"))

	// Fusión de un PATCH que solo trae `nameEs`, `sortOrder` y quita el enlace,
	// la imagen y la ficha del encargado (los demás campos conservan su valor).
	merged := saved
	merged.NameEs = "Alabanza y adoración"
	merged.SortOrder = 7
	merged.WhatsappURL = nil
	merged.ImageFile = nil
	merged.ImageAltEs = nil
	merged.ImageAltEn = nil
	merged.LeaderNameEs = nil
	merged.LeaderNameEn = nil
	merged.LeaderPhone = nil
	merged.LeaderVisible = false

	updated, err := repo.UpdateMinistry(ctx, merged)
	if err != nil {
		t.Fatalf("UpdateMinistry: %v", err)
	}
	assertMinistryEquals(t, "update fusionado", merged, updated)

	// Lo que quedó en la base es exactamente la fila fusionada, con los campos
	// ausentes del PATCH intactos.
	got, found, err := repo.GetMinistryByID(ctx, saved.ID)
	if err != nil || !found {
		t.Fatalf("GetMinistryByID = (found=%v, err=%v)", found, err)
	}
	assertMinistryEquals(t, "fila persistida", merged, got.Ministry)
	if got.NameEn == nil || *got.NameEn != "Worship" {
		t.Errorf("un campo ausente del PATCH cambió: NameEn = %v", got.NameEn)
	}
	if got.DescriptionEs != saved.DescriptionEs || got.ScheduleEn == nil ||
		*got.ScheduleEn != *saved.ScheduleEn || got.PlaceEs == nil || *got.PlaceEs != *saved.PlaceEs {
		t.Errorf("un campo ausente del PATCH cambió: %+v", got.Ministry)
	}
	if got.ID != saved.ID || !got.CreatedAt.Equal(saved.CreatedAt) {
		t.Errorf("el update cambió la identidad de la fila: %v / %v", got.ID, got.CreatedAt)
	}

	// Un id inexistente es "no existe" (404 del handler).
	missing := merged
	missing.ID = uuid.New()
	if _, err := repo.UpdateMinistry(ctx, missing); errorKind(err) != apperr.KindNotFound {
		t.Errorf("UpdateMinistry(inexistente) = %v, se esperaba not_found", err)
	}
}

// TestIntegrationDeleteMinistryRemovesTheRowWithoutTouchingTheFile fija A6-11 en
// la capa de datos: el borrado es físico y **no** borra el archivo de la imagen
// (el repository no tiene acceso al disco y no lo toca), simplemente la fila
// deja de referenciarlo, así que el archivo deja de servirse por
// `IsFilePublished`. La limpieza best-effort del archivo reemplazado o quitado
// vive en el service (T615, con un `Store` de prueba), no aquí.
func TestIntegrationDeleteMinistryRemovesTheRowWithoutTouchingTheFile(t *testing.T) {
	repo := newMinistryRepo(t)
	ctx := context.Background()

	in := baseMinistry("Alabanza")
	in.ImageFile = ptr(uniqueImageFile())
	saved := seedMinistry(t, repo, in)
	file := *saved.ImageFile

	if !filePublished(t, repo, file) {
		t.Fatal("la imagen de un ministerio publicado debe servirse")
	}

	deleted, err := repo.DeleteMinistry(ctx, saved.ID)
	if err != nil {
		t.Fatalf("DeleteMinistry: %v", err)
	}
	if !deleted {
		t.Error("DeleteMinistry no reportó el borrado de una fila existente")
	}
	if got := countRows(t, repo); got != 0 {
		t.Errorf("quedaron %d filas tras el borrado", got)
	}
	if filePublished(t, repo, file) {
		t.Error("la imagen de un ministerio borrado no debe servirse")
	}

	// Un id inexistente no es un error: el handler responde 404 con el `false`.
	deleted, err = repo.DeleteMinistry(ctx, uuid.New())
	if err != nil {
		t.Fatalf("DeleteMinistry(inexistente): %v", err)
	}
	if deleted {
		t.Error("DeleteMinistry dijo haber borrado un id inexistente")
	}
}

// TestIntegrationReplacingTheImageStopsServingTheOldFile es la otra mitad de
// A6-11: al reemplazar o quitar la imagen, la fila guarda el archivo nuevo (o
// ninguno) y el anterior deja de servirse sin que nadie lo borre desde el
// repository.
func TestIntegrationReplacingTheImageStopsServingTheOldFile(t *testing.T) {
	repo := newMinistryRepo(t)
	ctx := context.Background()

	in := baseMinistry("Alabanza")
	in.ImageFile = ptr(uniqueImageFile())
	saved := seedMinistry(t, repo, in)
	oldFile := *saved.ImageFile
	newFile := uniqueImageFile()

	withNewImage := saved
	withNewImage.ImageFile = ptr(newFile)
	if _, err := repo.UpdateMinistry(ctx, withNewImage); err != nil {
		t.Fatalf("UpdateMinistry(imagen nueva): %v", err)
	}
	if filePublished(t, repo, oldFile) {
		t.Error("el archivo reemplazado no debe servirse")
	}
	if !filePublished(t, repo, newFile) {
		t.Error("el archivo nuevo debe servirse")
	}

	// Y quitarla del todo (`null` explícito) deja la fila sin imagen.
	withoutImage := withNewImage
	withoutImage.ImageFile = nil
	withoutImage.ImageAltEs = nil
	withoutImage.ImageAltEn = nil
	refreshed, err := repo.UpdateMinistry(ctx, withoutImage)
	if err != nil {
		t.Fatalf("UpdateMinistry(sin imagen): %v", err)
	}
	if refreshed.ImageFile != nil || refreshed.ImageAltEs != nil {
		t.Errorf("la imagen quitada sigue en la fila: %+v", refreshed)
	}
	if filePublished(t, repo, newFile) {
		t.Error("la imagen quitada no debe servirse")
	}
}

// TestIntegrationMutationsRecordTheirAction verifica que cada mutación deja su
// fila de auditoría en la MISMA transacción, con el código y la etiqueta del
// módulo (FR-017/R6-12).
func TestIntegrationMutationsRecordTheirAction(t *testing.T) {
	repo := newMinistryRepo(t)
	ctx := context.Background()

	label := "Ministerios · auditoría de prueba " + uuid.NewString()
	// La fila de auditoría usa `user.create` con actor NULL: es el ÚNICO código
	// que la BD admite sin actor (F2) y así la prueba no toca la tabla `users`,
	// que las pruebas de integración del dominio usuarios truncan en paralelo
	// (mismo gesto que las pruebas de F4). Lo que aquí se verifica es la
	// transacción —mutación y registro juntos—, no el catálogo de códigos: que
	// los 8 códigos `ministry.*` estén en el registro cerrado de la BD y
	// coincidan con `platform/audit` lo fija T611 comparando el CHECK.
	action := audit.Action{
		Code:        audit.ActionUserCreate,
		TargetKind:  audit.TargetContent,
		TargetLabel: label,
		Result:      audit.ResultSuccess,
	}
	t.Cleanup(func() {
		if _, err := repo.pool.Exec(context.Background(),
			`DELETE FROM admin_actions WHERE target_label = $1`, label); err != nil {
			t.Logf("no se pudo limpiar la auditoría de la prueba: %v", err)
		}
	})

	saved, err := repo.InsertMinistry(ctx, baseMinistry("Alabanza"), action)
	if err != nil {
		t.Fatalf("InsertMinistry con auditoría: %v", err)
	}
	if got := countAdminActions(t, repo, label); got != 1 {
		t.Errorf("filas de auditoría = %d, se esperaba 1", got)
	}

	if deleted, err := repo.DeleteMinistry(ctx, saved.ID, action); err != nil || !deleted {
		t.Fatalf("DeleteMinistry con auditoría = (%v, %v)", deleted, err)
	}
	if got := countAdminActions(t, repo, label); got != 2 {
		t.Errorf("filas de auditoría tras el borrado = %d, se esperaban 2", got)
	}
}

// countAdminActions cuenta las filas de auditoría con una etiqueta exacta.
func countAdminActions(t *testing.T, repo *repository, label string) int {
	t.Helper()
	var total int
	if err := repo.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM admin_actions WHERE target_label = $1`, label).Scan(&total); err != nil {
		t.Fatalf("contar acciones administrativas: %v", err)
	}
	return total
}

// filePublished pregunta a la consulta compartida de `media` si el archivo se
// sirve hoy: es la misma consulta que resuelve `GET /api/v1/media/{fileName}`.
func filePublished(t *testing.T, repo *repository, file string) bool {
	t.Helper()
	today, err := time.Parse("2006-01-02", testToday)
	if err != nil {
		t.Fatalf("fecha de prueba %q: %v", testToday, err)
	}
	clock, err := time.Parse("15:04", testNowTime)
	if err != nil {
		t.Fatalf("hora de prueba %q: %v", testNowTime, err)
	}

	published, err := gendb.New(repo.pool).IsFilePublished(context.Background(), gendb.IsFilePublishedParams{
		FileName: file,
		Today:    pgtype.Date{Time: today, Valid: true},
		NowTime: pgtype.Time{
			Microseconds: (int64(clock.Hour())*3600 + int64(clock.Minute())*60) * 1_000_000,
			Valid:        true,
		},
	})
	if err != nil {
		t.Fatalf("IsFilePublished(%q): %v", file, err)
	}
	return published
}

// assertMinistryEquals compara la entidad esperada con la obtenida campo por
// campo (los sellos de tiempo se comprueban aparte: el alta los genera la base).
func assertMinistryEquals(t *testing.T, context string, want, got Ministry) {
	t.Helper()

	if got.NameEs != want.NameEs || !sameString(got.NameEn, want.NameEn) {
		t.Errorf("%s: nombre = (%q, %v), se esperaba (%q, %v)",
			context, got.NameEs, got.NameEn, want.NameEs, want.NameEn)
	}
	if got.DescriptionEs != want.DescriptionEs || !sameString(got.DescriptionEn, want.DescriptionEn) {
		t.Errorf("%s: descripción = (%q, %v), se esperaba (%q, %v)",
			context, got.DescriptionEs, got.DescriptionEn, want.DescriptionEs, want.DescriptionEn)
	}
	if !sameString(got.LeaderNameEs, want.LeaderNameEs) || !sameString(got.LeaderNameEn, want.LeaderNameEn) {
		t.Errorf("%s: encargado = (%v, %v), se esperaba (%v, %v)",
			context, got.LeaderNameEs, got.LeaderNameEn, want.LeaderNameEs, want.LeaderNameEn)
	}
	if !sameString(got.LeaderPhone, want.LeaderPhone) || !sameString(got.LeaderEmail, want.LeaderEmail) {
		t.Errorf("%s: contacto = (%v, %v), se esperaba (%v, %v)",
			context, got.LeaderPhone, got.LeaderEmail, want.LeaderPhone, want.LeaderEmail)
	}
	if got.LeaderVisible != want.LeaderVisible {
		t.Errorf("%s: leaderVisible = %v, se esperaba %v", context, got.LeaderVisible, want.LeaderVisible)
	}
	if !sameString(got.WhatsappURL, want.WhatsappURL) {
		t.Errorf("%s: whatsappUrl = %v, se esperaba %v", context, got.WhatsappURL, want.WhatsappURL)
	}
	if !sameString(got.ScheduleEs, want.ScheduleEs) || !sameString(got.ScheduleEn, want.ScheduleEn) {
		t.Errorf("%s: horario = (%v, %v), se esperaba (%v, %v)",
			context, got.ScheduleEs, got.ScheduleEn, want.ScheduleEs, want.ScheduleEn)
	}
	if !sameString(got.PlaceEs, want.PlaceEs) || !sameString(got.PlaceEn, want.PlaceEn) {
		t.Errorf("%s: lugar = (%v, %v), se esperaba (%v, %v)",
			context, got.PlaceEs, got.PlaceEn, want.PlaceEs, want.PlaceEn)
	}
	if !sameString(got.ImageFile, want.ImageFile) || !sameString(got.ImageAltEs, want.ImageAltEs) ||
		!sameString(got.ImageAltEn, want.ImageAltEn) {
		t.Errorf("%s: imagen = (%v, %v, %v), se esperaba (%v, %v, %v)",
			context, got.ImageFile, got.ImageAltEs, got.ImageAltEn,
			want.ImageFile, want.ImageAltEs, want.ImageAltEn)
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
