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
