package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Prefijos de las claves Redis (data-model.md, P1/P9).
const (
	sessionKeyPrefix      = "sess:"
	userSessionsKeyPrefix = "user_sessions:"
)

// NewRedisClient construye el cliente de Redis a partir de REDIS_URL. No abre
// conexión: go-redis es perezoso, igual que el pool de PostgreSQL (D8/plan R7).
func NewRedisClient(redisURL string) (*redis.Client, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	return redis.NewClient(options), nil
}

// redisStore es la implementación de Store sobre Redis.
type redisStore struct {
	client redis.UniversalClient
	cfg    StoreConfig
}

// NewRedisStore construye el Store de sesiones sobre el cliente Redis dado.
// Devuelve un error si el cliente o los tiempos de sesión no son válidos.
func NewRedisStore(client redis.UniversalClient, cfg StoreConfig) (Store, error) {
	if client == nil {
		return nil, errors.New("session: cliente de Redis nulo")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &redisStore{client: client, cfg: cfg}, nil
}

// sessionKey es la clave de la sesión de un token: sess:<sha256(token)>. El
// token en claro nunca se usa como clave.
func sessionKey(token string) string {
	return sessionKeyPrefix + HashToken(token)
}

// sessionKeyFromHash reconstruye la clave desde un hash ya calculado.
func sessionKeyFromHash(hash string) string {
	return sessionKeyPrefix + hash
}

// userSessionsKey es el índice de sesiones abiertas de una cuenta, que permite
// revocarlas todas sin SCAN (FR-012).
func userSessionsKey(userID uuid.UUID) string {
	return userSessionsKeyPrefix + userID.String()
}

// Create abre una sesión nueva y devuelve el token en claro (el único momento
// en que este existe: en Redis queda su hash).
func (s *redisStore) Create(ctx context.Context, userID uuid.UUID) (string, error) {
	if userID == uuid.Nil {
		return "", errors.New("session: userID vacío")
	}

	token, err := NewToken()
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	record := Session{
		UserID:            userID,
		CreatedAt:         now,
		LastSeenAt:        now,
		AbsoluteExpiresAt: now.Add(s.cfg.AbsoluteTTL),
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return "", fmt.Errorf("serializar sesión: %w", err)
	}

	key := sessionKey(token)
	ttl := nextTTL(now, record, s.cfg.IdleTTL)
	if err := s.client.Set(ctx, key, payload, ttl).Err(); err != nil {
		return "", fmt.Errorf("guardar sesión: %w", err)
	}

	// Índice por cuenta para revocar todas sus sesiones. El TTL del índice no
	// supera la vida absoluta de la sesión recién creada.
	indexKey := userSessionsKey(userID)
	pipe := s.client.TxPipeline()
	pipe.SAdd(ctx, indexKey, HashToken(token))
	pipe.Expire(ctx, indexKey, s.cfg.AbsoluteTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		// Best effort: si el índice falla, no se deja una sesión huérfana.
		_ = s.client.Del(ctx, key).Err()
		return "", fmt.Errorf("indexar sesión: %w", err)
	}

	return token, nil
}

// Resolve valida y refresca la sesión del token. Comprueba la vida absoluta
// inmóvil, refresca el TTL de inactividad acotado a ella y actualiza lastSeenAt
// como máximo una vez por LastSeenThrottle.
func (s *redisStore) Resolve(ctx context.Context, token string) (Session, error) {
	if token == "" {
		return Session{}, ErrSessionNotFound
	}

	key := sessionKey(token)
	raw, err := s.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("leer sesión: %w", err)
	}

	var record Session
	if err := json.Unmarshal(raw, &record); err != nil {
		// Dato corrupto: se descarta y se trata como sesión inexistente.
		_ = s.revokeHash(ctx, record.UserID, HashToken(token))
		return Session{}, fmt.Errorf("decodificar sesión: %w", err)
	}

	now := time.Now().UTC()
	ttl := nextTTL(now, record, s.cfg.IdleTTL)
	if record.ExpiredAt(now) || ttl <= 0 {
		if err := s.revokeHash(ctx, record.UserID, HashToken(token)); err != nil {
			return Session{}, fmt.Errorf("revocar sesión expirada: %w", err)
		}
		return Session{}, ErrSessionNotFound
	}

	if now.Sub(record.LastSeenAt) >= s.cfg.lastSeenThrottle() {
		record.LastSeenAt = now
		payload, err := json.Marshal(record)
		if err != nil {
			return Session{}, fmt.Errorf("serializar sesión refrescada: %w", err)
		}
		if err := s.client.Set(ctx, key, payload, ttl).Err(); err != nil {
			return Session{}, fmt.Errorf("refrescar sesión: %w", err)
		}
	} else if err := s.client.Expire(ctx, key, ttl).Err(); err != nil {
		return Session{}, fmt.Errorf("refrescar TTL de sesión: %w", err)
	}

	return record, nil
}

// Revoke cierra la sesión del token.
func (s *redisStore) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}

	key := sessionKey(token)
	raw, err := s.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("leer sesión a revocar: %w", err)
	}

	var record Session
	if err := json.Unmarshal(raw, &record); err != nil {
		// Sin userID fiable: al menos se borra la clave.
		if err := s.client.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("revocar sesión: %w", err)
		}
		return nil
	}

	if err := s.revokeHash(ctx, record.UserID, HashToken(token)); err != nil {
		return fmt.Errorf("revocar sesión: %w", err)
	}
	return nil
}

// RevokeUser cierra todas las sesiones de la cuenta (FR-012/R17).
func (s *redisStore) RevokeUser(ctx context.Context, userID uuid.UUID) error {
	indexKey := userSessionsKey(userID)
	hashes, err := s.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return fmt.Errorf("listar sesiones de la cuenta: %w", err)
	}

	keys := make([]string, 0, len(hashes)+1)
	for _, hash := range hashes {
		keys = append(keys, sessionKeyFromHash(hash))
	}
	keys = append(keys, indexKey)
	if err := s.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("revocar sesiones de la cuenta: %w", err)
	}
	return nil
}

// RevokeUserExcept cierra todas las sesiones de la cuenta salvo la del token
// indicado (cambio de contraseña propio, R17).
func (s *redisStore) RevokeUserExcept(ctx context.Context, userID uuid.UUID, keepToken string) error {
	indexKey := userSessionsKey(userID)
	hashes, err := s.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return fmt.Errorf("listar sesiones de la cuenta: %w", err)
	}

	keepHash := HashToken(keepToken)
	toRevoke := make([]string, 0, len(hashes))
	for _, hash := range hashes {
		if hash != keepHash {
			toRevoke = append(toRevoke, hash)
		}
	}
	if len(toRevoke) == 0 {
		return nil
	}

	pipe := s.client.TxPipeline()
	for _, hash := range toRevoke {
		pipe.Del(ctx, sessionKeyFromHash(hash))
		pipe.SRem(ctx, indexKey, hash)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("revocar las demás sesiones de la cuenta: %w", err)
	}
	return nil
}

// revokeHash borra la clave de sesión de un hash y lo quita del índice de la
// cuenta.
func (s *redisStore) revokeHash(ctx context.Context, userID uuid.UUID, hash string) error {
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, sessionKeyFromHash(hash))
	pipe.SRem(ctx, userSessionsKey(userID), hash)
	_, err := pipe.Exec(ctx)
	return err
}
