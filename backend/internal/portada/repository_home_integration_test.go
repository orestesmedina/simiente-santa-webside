//go:build integration

package portada

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"simiente-santa/backend/internal/platform/apperr"
	"simiente-santa/backend/internal/platform/testutil"
)

// Pruebas de integración del repository de la portada contra PostgreSQL real
// (T314/T315/T316). Requieren DATABASE_URL_TEST con las migraciones aplicadas
// (lo aporta el CI o el compose local; si no existe, se omiten solas). Estas
// pruebas truncan las tablas home_*: DATABASE_URL_TEST debe apuntar a una base
// de pruebas dedicada.

// newHomeRepo construye el repositorio contra la base de pruebas y vacía las
// tablas de contenido para partir de un estado conocido.
func newHomeRepo(t *testing.T) (*repository, *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	pool := testutil.Pool(t, ctx)
	const truncate = `TRUNCATE home_identity, home_about, home_contact,
		home_services, home_whatsapp_channels, home_social_links`
	if _, err := pool.Exec(ctx, truncate); err != nil {
		t.Fatalf("vaciar tablas de la portada: %v", err)
	}
	return NewRepository(pool), pool
}

// countHomeRows cuenta las filas de una tabla del dominio (para comprobar el
// singleton: exactamente una). La tabla se resuelve por una lista fija: no hay
// SQL construido con datos variables.
func countHomeRows(t *testing.T, pool *pgxpool.Pool, table string) int {
	t.Helper()

	queries := map[string]string{
		"home_identity":          `SELECT count(*) FROM home_identity`,
		"home_about":             `SELECT count(*) FROM home_about`,
		"home_contact":           `SELECT count(*) FROM home_contact`,
		"home_services":          `SELECT count(*) FROM home_services`,
		"home_whatsapp_channels": `SELECT count(*) FROM home_whatsapp_channels`,
		"home_social_links":      `SELECT count(*) FROM home_social_links`,
	}
	query, ok := queries[table]
	if !ok {
		t.Fatalf("tabla desconocida en countHomeRows: %s", table)
	}
	var total int
	if err := pool.QueryRow(context.Background(), query).Scan(&total); err != nil {
		t.Fatalf("contar filas de %s: %v", table, err)
	}
	return total
}

// publishedIdentity devuelve una identidad válida en el estado indicado.
func publishedIdentity(state PublicationState) Identity {
	return Identity{
		NameEs:           "Iglesia Simiente",
		NameEn:           ptr("Seed Church"),
		PublicationState: state,
	}
}

func ptr(s string) *string { return &s }

func TestIntegrationUpsertHomeIdentitySingleton(t *testing.T) {
	repo, pool := newHomeRepo(t)
	ctx := context.Background()

	// Sin fila todavía: el panel la tipa como null (bool false).
	if _, ok, err := repo.GetHomeIdentity(ctx); err != nil || ok {
		t.Fatalf("GetHomeIdentity antes de guardar = ok %v (%v), se esperaba false", ok, err)
	}

	first, err := repo.UpsertHomeIdentity(ctx, publishedIdentity(StateDraft))
	if err != nil {
		t.Fatalf("UpsertHomeIdentity: %v", err)
	}
	if count := countHomeRows(t, pool, "home_identity"); count != 1 {
		t.Fatalf("filas de home_identity = %d, se esperaba 1", count)
	}
	if first.NameEs != "Iglesia Simiente" || first.PublicationState != StateDraft {
		t.Fatalf("identidad guardada = %+v", first)
	}

	// Segundo upsert: sigue habiendo una fila y se actualiza (updated_at).
	time.Sleep(2 * time.Millisecond)
	second := publishedIdentity(StatePublished)
	second.NameEs = "Iglesia Simiente Renovada"
	updated, err := repo.UpsertHomeIdentity(ctx, second)
	if err != nil {
		t.Fatalf("segundo UpsertHomeIdentity: %v", err)
	}
	if count := countHomeRows(t, pool, "home_identity"); count != 1 {
		t.Fatalf("tras el segundo upsert filas = %d, se esperaba 1", count)
	}
	if updated.NameEs != "Iglesia Simiente Renovada" {
		t.Fatalf("nombre tras el upsert = %q", updated.NameEs)
	}
	if !updated.UpdatedAt.After(first.UpdatedAt) {
		t.Fatalf("updated_at no avanzó: %v → %v", first.UpdatedAt, updated.UpdatedAt)
	}
	if updated.CreatedAt != first.CreatedAt {
		t.Fatalf("created_at cambió: %v → %v", first.CreatedAt, updated.CreatedAt)
	}
}

func TestIntegrationConcurrentUpsertsLeaveOneRow(t *testing.T) {
	repo, pool := newHomeRepo(t)
	ctx := context.Background()

	const writers = 4
	var wg sync.WaitGroup
	errs := make([]error, writers)
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			identity := publishedIdentity(StateDraft)
			identity.NameEs = "Iglesia " + strings.Repeat("x", n+1)
			_, errs[n] = repo.UpsertHomeIdentity(ctx, identity)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("upsert concurrente %d: %v", i, err)
		}
	}
	if count := countHomeRows(t, pool, "home_identity"); count != 1 {
		t.Fatalf("filas tras upserts concurrentes = %d, se esperaba 1", count)
	}

	got, ok, err := repo.GetHomeIdentity(ctx)
	if err != nil || !ok {
		t.Fatalf("GetHomeIdentity = ok %v (%v)", ok, err)
	}
	// La fila es una escritura COMPLETA de alguno de los escritores (nunca un
	// valor mezclado) y su nombre corresponde a uno de ellos.
	if !strings.HasPrefix(got.NameEs, "Iglesia ") {
		t.Fatalf("nombre final = %q", got.NameEs)
	}
}

func TestIntegrationSingletonsPublishPerSection(t *testing.T) {
	repo, pool := newHomeRepo(t)
	ctx := context.Background()

	// Identidad en borrador: nunca sale por la vía publicada.
	if _, err := repo.UpsertHomeIdentity(ctx, publishedIdentity(StateDraft)); err != nil {
		t.Fatalf("UpsertHomeIdentity(draft): %v", err)
	}
	if _, ok, err := repo.GetHomeIdentityPublished(ctx); err != nil || ok {
		t.Fatalf("GetHomeIdentityPublished(draft) = ok %v (%v), se esperaba false", ok, err)
	}

	// Publicar la identidad no publica «quiénes somos» (por sección, FR-013).
	if _, err := repo.UpsertHomeIdentity(ctx, publishedIdentity(StatePublished)); err != nil {
		t.Fatalf("UpsertHomeIdentity(published): %v", err)
	}
	if _, err := repo.UpsertHomeAbout(ctx, About{TextEs: "Somos comunidad", PublicationState: StateDraft}); err != nil {
		t.Fatalf("UpsertHomeAbout(draft): %v", err)
	}
	if _, ok, err := repo.GetHomeIdentityPublished(ctx); err != nil || !ok {
		t.Fatalf("GetHomeIdentityPublished(published) = ok %v (%v), se esperaba true", ok, err)
	}
	if _, ok, err := repo.GetHomeAboutPublished(ctx); err != nil || ok {
		t.Fatalf("GetHomeAboutPublished(draft) = ok %v (%v), se esperaba false", ok, err)
	}

	// Contacto publicado con su dirección traducida.
	if _, err := repo.UpsertHomeContact(ctx, Contact{
		AddressEs:        "Calle 1",
		Email:            "info@simiente.org",
		Phone:            "+34612345678",
		PublicationState: StatePublished,
	}); err != nil {
		t.Fatalf("UpsertHomeContact: %v", err)
	}
	contact, ok, err := repo.GetHomeContactPublished(ctx)
	if err != nil || !ok {
		t.Fatalf("GetHomeContactPublished = ok %v (%v)", ok, err)
	}
	if contact.Email != "info@simiente.org" {
		t.Fatalf("contacto publicado = %+v", contact)
	}
	if count := countHomeRows(t, pool, "home_about"); count != 1 {
		t.Fatalf("filas de home_about = %d", count)
	}

	// Retirar la identidad vuelve a ocultarla (FR-014).
	if _, err := repo.UpsertHomeIdentity(ctx, publishedIdentity(StateDraft)); err != nil {
		t.Fatalf("UpsertHomeIdentity(retirar): %v", err)
	}
	if _, ok, err := repo.GetHomeIdentityPublished(ctx); err != nil || ok {
		t.Fatalf("GetHomeIdentityPublished tras retirar = ok %v (%v), se esperaba false", ok, err)
	}
}

func TestIntegrationIsHomeFilePublishedFollowsState(t *testing.T) {
	repo, _ := newHomeRepo(t)
	ctx := context.Background()

	const file = "img_0123456789abcdef0123456789abcdef0123.png"
	identity := publishedIdentity(StateDraft)
	identity.LogoFile = ptr(file)
	identity.LogoAltEs = ptr("Logotipo")

	if _, err := repo.UpsertHomeIdentity(ctx, identity); err != nil {
		t.Fatalf("UpsertHomeIdentity(draft): %v", err)
	}
	if published, err := repo.IsHomeFilePublished(ctx, file); err != nil || published {
		t.Fatalf("IsHomeFilePublished(draft) = %v (%v), se esperaba false", published, err)
	}

	identity.PublicationState = StatePublished
	if _, err := repo.UpsertHomeIdentity(ctx, identity); err != nil {
		t.Fatalf("UpsertHomeIdentity(published): %v", err)
	}
	if published, err := repo.IsHomeFilePublished(ctx, file); err != nil || !published {
		t.Fatalf("IsHomeFilePublished(published) = %v (%v), se esperaba true", published, err)
	}

	// Un archivo huérfano (no referenciado) nunca se sirve.
	if published, err := repo.IsHomeFilePublished(ctx, "img_000000000000000000000000000000000000.jpg"); err != nil || published {
		t.Fatalf("IsHomeFilePublished(huérfano) = %v (%v), se esperaba false", published, err)
	}

	identity.PublicationState = StateDraft
	if _, err := repo.UpsertHomeIdentity(ctx, identity); err != nil {
		t.Fatalf("UpsertHomeIdentity(retirar): %v", err)
	}
	if published, err := repo.IsHomeFilePublished(ctx, file); err != nil || published {
		t.Fatalf("IsHomeFilePublished(retirado) = %v (%v), se esperaba false", published, err)
	}
}

func TestIntegrationSingletonChecks(t *testing.T) {
	repo, _ := newHomeRepo(t)
	ctx := context.Background()

	t.Run("nombre vacío o demasiado largo", func(t *testing.T) {
		empty := publishedIdentity(StateDraft)
		empty.NameEs = ""
		if _, err := repo.UpsertHomeIdentity(ctx, empty); !isInvalid(err) {
			t.Fatalf("UpsertHomeIdentity(nombre vacío) = %v, se esperaba invalid", err)
		}
		long := publishedIdentity(StateDraft)
		long.NameEs = strings.Repeat("a", 161)
		if _, err := repo.UpsertHomeIdentity(ctx, long); !isInvalid(err) {
			t.Fatalf("UpsertHomeIdentity(nombre largo) = %v, se esperaba invalid", err)
		}
	})

	t.Run("alt obligatorio con imagen", func(t *testing.T) {
		identity := publishedIdentity(StateDraft)
		identity.LogoFile = ptr("img_0123456789abcdef0123456789abcdef0123.png")
		// Sin LogoAltEs: el CHECK de FR-019 lo rechaza.
		if _, err := repo.UpsertHomeIdentity(ctx, identity); !isInvalid(err) {
			t.Fatalf("UpsertHomeIdentity(imagen sin alt) = %v, se esperaba invalid", err)
		}
	})

	t.Run("quiénes somos ≤ 1000", func(t *testing.T) {
		if _, err := repo.UpsertHomeAbout(ctx, About{
			TextEs:           strings.Repeat("a", 1001),
			PublicationState: StateDraft,
		}); !isInvalid(err) {
			t.Fatalf("UpsertHomeAbout(1001) = %v, se esperaba invalid", err)
		}
	})
}

// isInvalid indica si err es un apperr de kind Invalid (violación de CHECK
// traducida por classify).
func isInvalid(err error) bool {
	var domainErr *apperr.Error
	return errors.As(err, &domainErr) && domainErr.Kind == apperr.KindInvalid
}

// isConflict indica si err es un apperr de kind Conflict (violación de UNIQUE
// traducida por classify).
func isConflict(err error) bool {
	var domainErr *apperr.Error
	return errors.As(err, &domainErr) && domainErr.Kind == apperr.KindConflict
}

// --- T315: colecciones ---

func TestIntegrationServicesCRUDAndOrder(t *testing.T) {
	repo, _ := newHomeRepo(t)
	ctx := context.Background()

	second, err := repo.InsertHomeService(ctx, Service{
		DayOfWeek: 0, StartTime: "12:00", EndTime: ptr("13:00"),
		NameEs: "Segundo", PlaceEs: "Templo", PublicationState: StatePublished, SortOrder: 2,
	})
	if err != nil {
		t.Fatalf("InsertHomeService(2): %v", err)
	}
	if _, err := repo.InsertHomeService(ctx, Service{
		DayOfWeek: 0, StartTime: "10:00",
		NameEs: "Primero", PlaceEs: "Templo", PublicationState: StatePublished, SortOrder: 1,
	}); err != nil {
		t.Fatalf("InsertHomeService(1): %v", err)
	}
	if _, err := repo.InsertHomeService(ctx, Service{
		DayOfWeek: 6, StartTime: "09:00",
		NameEs: "Borrador", PlaceEs: "Anexo", PublicationState: StateDraft, SortOrder: 3,
	}); err != nil {
		t.Fatalf("InsertHomeService(draft): %v", err)
	}

	// Orden estable por `sort_order, id` en ambas vías.
	all, err := repo.ListHomeServices(ctx)
	if err != nil {
		t.Fatalf("ListHomeServices: %v", err)
	}
	if len(all) != 3 || all[0].NameEs != "Primero" || all[1].NameEs != "Segundo" {
		t.Fatalf("ListHomeServices = %+v", all)
	}
	published, err := repo.ListHomeServicesPublished(ctx)
	if err != nil {
		t.Fatalf("ListHomeServicesPublished: %v", err)
	}
	if len(published) != 2 || published[0].NameEs != "Primero" {
		t.Fatalf("ListHomeServicesPublished = %+v (no debe incluir borradores)", published)
	}

	// Edición: cambia el estado publicando el borrador.
	third := all[2]
	third.PublicationState = StatePublished
	if _, err := repo.UpdateHomeService(ctx, third); err != nil {
		t.Fatalf("UpdateHomeService: %v", err)
	}
	if published, _ := repo.ListHomeServicesPublished(ctx); len(published) != 3 {
		t.Fatalf("tras publicar hay %d publicados, se esperaban 3", len(published))
	}

	// Lectura por id y borrado físico.
	got, ok, err := repo.GetHomeServiceByID(ctx, second.ID)
	if err != nil || !ok || got.NameEs != "Segundo" {
		t.Fatalf("GetHomeServiceByID = %+v ok=%v err=%v", got, ok, err)
	}
	deleted, err := repo.DeleteHomeService(ctx, second.ID)
	if err != nil || !deleted {
		t.Fatalf("DeleteHomeService = %v (%v)", deleted, err)
	}
	if _, ok, _ := repo.GetHomeServiceByID(ctx, second.ID); ok {
		t.Fatal("el servicio seguía existiendo tras borrarlo")
	}
	if deleted, _ := repo.DeleteHomeService(ctx, second.ID); deleted {
		t.Fatal("el segundo borrado debía devolver false")
	}
}

func TestIntegrationServiceChecks(t *testing.T) {
	repo, _ := newHomeRepo(t)
	ctx := context.Background()

	t.Run("day_of_week fuera de rango", func(t *testing.T) {
		_, err := repo.InsertHomeService(ctx, Service{
			DayOfWeek: 7, StartTime: "10:00", NameEs: "X", PlaceEs: "Y", PublicationState: StateDraft,
		})
		if !isInvalid(err) {
			t.Fatalf("InsertHomeService(día 7) = %v, se esperaba invalid", err)
		}
	})

	t.Run("start_time mal formado", func(t *testing.T) {
		_, err := repo.InsertHomeService(ctx, Service{
			DayOfWeek: 1, StartTime: "25:00", NameEs: "X", PlaceEs: "Y", PublicationState: StateDraft,
		})
		if !isInvalid(err) {
			t.Fatalf("InsertHomeService(hora 25:00) = %v, se esperaba invalid", err)
		}
	})

	t.Run("end_time no posterior", func(t *testing.T) {
		_, err := repo.InsertHomeService(ctx, Service{
			DayOfWeek: 1, StartTime: "10:00", EndTime: ptr("09:00"),
			NameEs: "X", PlaceEs: "Y", PublicationState: StateDraft,
		})
		if !isInvalid(err) {
			t.Fatalf("InsertHomeService(fin anterior) = %v, se esperaba invalid", err)
		}
	})
}

func TestIntegrationWhatsappDuplicateAndPublished(t *testing.T) {
	repo, _ := newHomeRepo(t)
	ctx := context.Background()

	channel := WhatsappChannel{
		Kind: KindDirect, Destination: "+34612345678", NameEs: "Culto",
		PublicationState: StatePublished, SortOrder: 0,
	}
	if _, err := repo.InsertHomeWhatsappChannel(ctx, channel); err != nil {
		t.Fatalf("InsertHomeWhatsappChannel: %v", err)
	}
	// Duplicado exacto (mismo kind, destino y nombre) → conflict.
	if _, err := repo.InsertHomeWhatsappChannel(ctx, channel); !isConflict(err) {
		t.Fatalf("InsertHomeWhatsappChannel(duplicado) = %v, se esperaba conflict", err)
	}
	// Mismo destino con otro nombre: permitido (la spec solo rechaza el exacto).
	other := channel
	other.NameEs = "Oración"
	if _, err := repo.InsertHomeWhatsappChannel(ctx, other); err != nil {
		t.Fatalf("InsertHomeWhatsappChannel(otro nombre): %v", err)
	}
	// Canal en borrador no aparece en la vía pública.
	draft := WhatsappChannel{
		Kind: KindGroup, Destination: "https://chat.whatsapp.com/abc", NameEs: "Grupo",
		PublicationState: StateDraft,
	}
	if _, err := repo.InsertHomeWhatsappChannel(ctx, draft); err != nil {
		t.Fatalf("InsertHomeWhatsappChannel(draft): %v", err)
	}

	published, err := repo.ListHomeWhatsappChannelsPublished(ctx)
	if err != nil {
		t.Fatalf("ListHomeWhatsappChannelsPublished: %v", err)
	}
	if len(published) != 2 {
		t.Fatalf("canales publicados = %d, se esperaban 2", len(published))
	}
	all, err := repo.ListHomeWhatsappChannels(ctx)
	if err != nil || len(all) != 3 {
		t.Fatalf("ListHomeWhatsappChannels = %d (%v)", len(all), err)
	}
}

func TestIntegrationSocialNetworkCatalogAndUnique(t *testing.T) {
	repo, _ := newHomeRepo(t)
	ctx := context.Background()

	if _, err := repo.InsertHomeSocialLink(ctx, SocialLink{
		Network: "facebook", URL: "https://facebook.com/simiente", PublicationState: StatePublished,
	}); err != nil {
		t.Fatalf("InsertHomeSocialLink(facebook): %v", err)
	}
	if _, err := repo.InsertHomeSocialLink(ctx, SocialLink{
		Network: "instagram", URL: "https://instagram.com/simiente", PublicationState: StateDraft,
	}); err != nil {
		t.Fatalf("InsertHomeSocialLink(instagram): %v", err)
	}
	// Segundo enlace para la misma red → conflict (UNIQUE (network)).
	if _, err := repo.InsertHomeSocialLink(ctx, SocialLink{
		Network: "facebook", URL: "https://facebook.com/otro", PublicationState: StateDraft,
	}); !isConflict(err) {
		t.Fatalf("InsertHomeSocialLink(duplicado) = %v, se esperaba conflict", err)
	}
	// Red fuera del catálogo → invalid (CHECK).
	if _, err := repo.InsertHomeSocialLink(ctx, SocialLink{
		Network: "x", URL: "https://x.com/simiente", PublicationState: StateDraft,
	}); !isInvalid(err) {
		t.Fatalf("InsertHomeSocialLink(red fuera) = %v, se esperaba invalid", err)
	}

	published, err := repo.ListHomeSocialLinksPublished(ctx)
	if err != nil {
		t.Fatalf("ListHomeSocialLinksPublished: %v", err)
	}
	if len(published) != 1 || published[0].Network != "facebook" {
		t.Fatalf("enlaces publicados = %+v", published)
	}
	all, err := repo.ListHomeSocialLinks(ctx)
	if err != nil || len(all) != 2 {
		t.Fatalf("ListHomeSocialLinks = %d (%v)", len(all), err)
	}
	// Orden por red.
	if all[0].Network != "facebook" || all[1].Network != "instagram" {
		t.Fatalf("orden de redes = %+v", all)
	}
}
