package repository

import (
	"context"
	"cruder/internal/model"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var ErrUniqueViolation = errors.New("unique constraint violation")

type UserRepository interface {
	GetAll() ([]model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
	GetByUUID(uuid uuid.UUID) (*model.User, error)
	Create(username, email, fullName string) (*model.User, error)
	UpdateByUUID(uuid uuid.UUID, username, email, fullName string) (*model.User, error)
	DeleteByUUID(uuid uuid.UUID) (bool, error)
	UpdateByID(id int64, username, email, fullName string) (*model.User, error)
	DeleteByID(id int64) (bool, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetAll() ([]model.User, error) {
	rows, err := r.db.QueryContext(context.Background(), `SELECT id, uuid, username, email, full_name FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(context.Background(), `SELECT id, uuid, username, email, full_name FROM users WHERE username = $1`, username).
		Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByID(id int64) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(context.Background(), `SELECT id, uuid, username, email, full_name FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByUUID(uuid uuid.UUID) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(context.Background(), `SELECT id, uuid, username, email, full_name FROM users WHERE uuid = $1`, uuid.String()).
		Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(username, email, fullName string) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(
		context.Background(),
		`INSERT INTO users (username, email, full_name) VALUES ($1, $2, $3) RETURNING id, uuid, username, email, full_name`,
		username,
		email,
		fullName,
	).Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
		return nil, mapPQError(err)
	}
	return &u, nil
}

func (r *userRepository) UpdateByUUID(uuid uuid.UUID, username, email, fullName string) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(
		context.Background(),
		`UPDATE users SET username = $1, email = $2, full_name = $3 WHERE uuid = $4 RETURNING id, uuid, username, email, full_name`,
		username,
		email,
		fullName,
		uuid,
	).Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, mapPQError(err)
	}
	return &u, nil
}

func (r *userRepository) DeleteByUUID(uuid uuid.UUID) (bool, error) {
	res, err := r.db.ExecContext(context.Background(), `DELETE FROM users WHERE uuid = $1`, uuid)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (r *userRepository) UpdateByID(id int64, username, email, fullName string) (*model.User, error) {
	var u model.User
	if err := r.db.QueryRowContext(
		context.Background(),
		`UPDATE users SET username = $1, email = $2, full_name = $3 WHERE id = $4 RETURNING id, uuid, username, email, full_name`,
		username,
		email,
		fullName,
		id,
	).Scan(&u.ID, &u.UUID, &u.Username, &u.Email, &u.FullName); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, mapPQError(err)
	}
	return &u, nil
}

func (r *userRepository) DeleteByID(id int64) (bool, error) {
	res, err := r.db.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func mapPQError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrUniqueViolation
	}
	return err
}
