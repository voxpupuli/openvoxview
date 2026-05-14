package db

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var (
	ErrTokenNotFound = errors.New("refresh token not found or expired")
	ErrTokenRevoked  = errors.New("refresh token has been revoked")
)

type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

func GenerateRefreshToken() (raw string, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}
	raw = hex.EncodeToString(b)
	hash = hashToken(raw)
	return raw, hash, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (d *Database) StoreRefreshToken(userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := d.db.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}
	return nil
}

func (d *Database) ValidateRefreshToken(rawToken string) (*RefreshToken, error) {
	h := hashToken(rawToken)

	var rt RefreshToken
	var revokedAt sql.NullTime
	err := d.db.QueryRow(
		`SELECT id, user_id, token_hash, expires_at, created_at, revoked_at
		 FROM refresh_tokens WHERE token_hash = ?`, h,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt, &revokedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if revokedAt.Valid {
		d.RevokeAllUserTokens(rt.UserID)
		return nil, ErrTokenRevoked
	}

	if time.Now().After(rt.ExpiresAt) {
		return nil, ErrTokenNotFound
	}

	return &rt, nil
}

func (d *Database) RevokeRefreshToken(rawToken string) error {
	h := hashToken(rawToken)
	_, err := d.db.Exec(
		`UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE token_hash = ? AND revoked_at IS NULL`,
		h,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	return nil
}

func (d *Database) RevokeAllUserTokens(userID int64) error {
	_, err := d.db.Exec(
		`UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE user_id = ? AND revoked_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}
	return nil
}

func (d *Database) CleanupExpiredTokens() error {
	_, err := d.db.Exec(
		`DELETE FROM refresh_tokens WHERE expires_at < CURRENT_TIMESTAMP OR revoked_at IS NOT NULL`,
	)
	if err != nil {
		return fmt.Errorf("failed to cleanup tokens: %w", err)
	}
	return nil
}
