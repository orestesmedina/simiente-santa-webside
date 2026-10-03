// Package database construye y administra el pool de conexiones a PostgreSQL
// (pgxpool), el helper de salud con timeout que usa /healthz y el helper de
// transacciones.
//
// Es infraestructura: no conoce ningún dominio (arq. R1) ni escribe SQL de
// negocio (el código generado por sqlc vive en internal/db).
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PingTimeout acota cada comprobación de salud de la base de datos. /healthz
// responde en ≤2 s aunque la BD esté caída (D8).
const PingTimeout = 2 * time.Second

// NewPool construye el *pgxpool.Pool desde DATABASE_URL. El pool es perezoso:
// no abre conexiones al construirse, así que una base de datos ausente o lenta
// no bloquea ni tumba el arranque (plan R7). Un DATABASE_URL malformado sí
// falla de inmediato.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}
	return pool, nil
}

// pinger es lo mínimo que necesita Ping, para poder probar la derivación del
// contexto y el timeout sin una base de datos real.
type pinger interface {
	Ping(ctx context.Context) error
}

// Ping comprueba la salud de la base de datos con un timeout de PingTimeout.
// La llama el repositorio del dominio status en cada petición a /healthz.
func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	return ping(ctx, pool)
}

func ping(ctx context.Context, p pinger) error {
	ctx, cancel := context.WithTimeout(ctx, PingTimeout)
	defer cancel()

	if err := p.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	return nil
}

// WithTx ejecuta fn dentro de una transacción: confirma si fn devuelve nil y
// revierte en cualquier otro caso (incluido un panic, que se propaga).
func WithTx(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	// Rollback tras un Commit ya confirmado es un no-op (devuelve ErrTxClosed),
	// así que se ignora de forma deliberada.
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
