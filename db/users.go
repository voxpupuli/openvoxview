package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUsernameExists    = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email,omitempty"`
	DisplayName  string    `json:"display_name,omitempty"`
	GivenName    string    `json:"given_name,omitempty"`
	Surname      string    `json:"surname,omitempty"`
	PasswordHash string    `json:"-"`
	AuthSource   string    `json:"auth_source"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

func (d *Database) CreateUser(username, email, displayName, password string, isAdmin bool) (*User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	result, err := d.db.Exec(
		`INSERT INTO users (username, email, display_name, password_hash, auth_source, is_admin)
		 VALUES (?, ?, ?, ?, 'local', ?)`,
		username, email, displayName, hash, isAdmin,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, ErrUsernameExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, _ := result.LastInsertId()
	return d.GetUserByID(id)
}

func (d *Database) GetUserByID(id int64) (*User, error) {
	user := &User{}
	err := d.db.QueryRow(
		`SELECT id, username, COALESCE(email,''), COALESCE(display_name,''),
		        COALESCE(given_name,''), COALESCE(surname,''),
		        COALESCE(password_hash,''), auth_source, is_admin, created_at, updated_at
		 FROM users WHERE id = ?`, id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName,
		&user.GivenName, &user.Surname, &user.PasswordHash,
		&user.AuthSource, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (d *Database) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := d.db.QueryRow(
		`SELECT id, username, COALESCE(email,''), COALESCE(display_name,''),
		        COALESCE(given_name,''), COALESCE(surname,''),
		        COALESCE(password_hash,''), auth_source, is_admin, created_at, updated_at
		 FROM users WHERE username = ?`, username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.DisplayName,
		&user.GivenName, &user.Surname, &user.PasswordHash,
		&user.AuthSource, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (d *Database) AuthenticateUser(username, password string) (*User, error) {
	user, err := d.GetUserByUsername(username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if user.AuthSource != "local" {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (d *Database) ListUsers() ([]User, error) {
	rows, err := d.db.Query(
		`SELECT id, username, COALESCE(email,''), COALESCE(display_name,''),
		        COALESCE(given_name,''), COALESCE(surname,''),
		        auth_source, is_admin, created_at, updated_at
		 FROM users ORDER BY username`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.DisplayName,
			&u.GivenName, &u.Surname, &u.AuthSource, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}

	if users == nil {
		users = []User{}
	}
	return users, rows.Err()
}

func (d *Database) UpdateUser(id int64, email, displayName *string, password *string, isAdmin *bool) (*User, error) {
	if password != nil {
		hash, err := HashPassword(*password)
		if err != nil {
			return nil, err
		}
		if isAdmin != nil {
			_, err = d.db.Exec(
				`UPDATE users SET email = COALESCE(?, email), display_name = COALESCE(?, display_name),
				 password_hash = ?, is_admin = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
				email, displayName, hash, *isAdmin, id,
			)
		} else {
			_, err = d.db.Exec(
				`UPDATE users SET email = COALESCE(?, email), display_name = COALESCE(?, display_name),
				 password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
				email, displayName, hash, id,
			)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	} else {
		var err error
		if isAdmin != nil {
			_, err = d.db.Exec(
				`UPDATE users SET email = COALESCE(?, email), display_name = COALESCE(?, display_name),
				 is_admin = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
				email, displayName, *isAdmin, id,
			)
		} else {
			_, err = d.db.Exec(
				`UPDATE users SET email = COALESCE(?, email), display_name = COALESCE(?, display_name),
				 updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
				email, displayName, id,
			)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	return d.GetUserByID(id)
}

func (d *Database) DeleteUser(id int64) error {
	result, err := d.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (d *Database) UpsertSamlUser(email, givenName, surname, displayName string) (*User, error) {
	// Use email as the username for SAML users
	username := email

	var existingID int64
	err := d.db.QueryRow(`SELECT id FROM users WHERE username = ? AND auth_source = 'saml'`, username).Scan(&existingID)
	if err == nil {
		// User exists — update profile attributes from IdP
		_, err = d.db.Exec(
			`UPDATE users SET email = ?, given_name = ?, surname = ?, display_name = ?, updated_at = CURRENT_TIMESTAMP
			 WHERE id = ?`,
			email, givenName, surname, displayName, existingID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update SAML user: %w", err)
		}
		return d.GetUserByID(existingID)
	}

	// New SAML user — insert
	result, err := d.db.Exec(
		`INSERT INTO users (username, email, given_name, surname, display_name, password_hash, auth_source)
		 VALUES (?, ?, ?, ?, ?, NULL, 'saml')`,
		username, email, givenName, surname, displayName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SAML user: %w", err)
	}

	id, _ := result.LastInsertId()
	return d.GetUserByID(id)
}

func (d *Database) UserCount() (int64, error) {
	var count int64
	err := d.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func isUniqueConstraintError(err error) bool {
	return err != nil && (contains(err.Error(), "UNIQUE constraint failed") || contains(err.Error(), "unique constraint"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
