package media

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	gendb "simiente-santa/backend/internal/db"
)

// repository.go es el acceso a datos del dominio media: la consulta compartida
// `IsFilePublished` (internal/db/queries/media.sql, T404). No conoce HTTP ni las
// reglas de negocio: solo SQL parametrizado y el mapeo de los tipos de pgx, que
// se traduce aquí y solo aquí (§8.1.6).

// repository implementa el puerto Repository del service.
type repository struct {
	q *gendb.Queries
}

// NewRepository construye el repositorio del dominio media sobre el pool
// compartido que compone cmd/api.
func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{q: gendb.New(pool)}
}

// IsFilePublished responde si el nombre de archivo lo referencia contenido
// publicado Y visible hoy: la identidad de la portada publicada (F3, sin fecha),
// un evento/actividad publicados con fecha próxima (F4, regla R4-2) o un
// ministerio publicado (F6, sin fecha). `today` y `nowTime` los calcula el
// service en APP_TZ y viajan parametrizados.
func (r *repository) IsFilePublished(ctx context.Context, fileName, today, nowTime string) (bool, error) {
	published, err := r.q.IsFilePublished(ctx, gendb.IsFilePublishedParams{
		FileName: fileName,
		Today:    pgDate(today),
		NowTime:  pgTime(nowTime),
	})
	if err != nil {
		return false, fmt.Errorf("is file published: %w", err)
	}
	return published, nil
}

// clockPattern es el patrón "HH:MM" en 24 h que exige el contrato y el CHECK de
// la base: el mapeo no debe colar una hora que la BD rechazaría.
var clockPattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// pgDate traduce una fecha de calendario "YYYY-MM-DD" al parámetro DATE.
func pgDate(date string) pgtype.Date {
	parsed, err := time.Parse(dateFormat, strings.TrimSpace(date))
	if err != nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: parsed, Valid: true}
}

// pgTime traduce una hora "HH:MM" al parámetro TIME (microsegundos desde la
// medianoche).
func pgTime(clock string) pgtype.Time {
	clock = strings.TrimSpace(clock)
	if !clockPattern.MatchString(clock) {
		return pgtype.Time{}
	}
	parsed, err := time.Parse(timeFormat, clock)
	if err != nil {
		return pgtype.Time{}
	}
	microseconds := (int64(parsed.Hour())*3600 + int64(parsed.Minute())*60) * int64(time.Second/time.Microsecond)
	return pgtype.Time{Microseconds: microseconds, Valid: true}
}

// compile-time check: *repository must satisfy the port the service declares.
var _ Repository = (*repository)(nil)
