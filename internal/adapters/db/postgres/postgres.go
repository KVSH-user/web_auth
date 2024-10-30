package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"log/slog"
	"web_auth/internal/models"
	"web_auth/internal/modules/auth"

	"web_auth/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose"
)

const migrationDir = "db/migrations/postgres"

type Storage struct {
	db  *pgx.Conn
	log *slog.Logger
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*Storage, error) {
	const op = "postgres.New"

	log = log.With(slog.String("op", op))

	url := dbStringConverter(cfg)

	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		log.Error("can`t connect to db: ", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := conn.Ping(ctx); err != nil {
		log.Error("failed ping db: ", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := applyMigrations(ctx, conn, migrationDir); err != nil {
		log.Error("can't migrate up: ", "err", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db:  conn,
		log: log,
	}, nil
}

func (s *Storage) Close(ctx context.Context) error {
	return s.db.Close(ctx)
}

func applyMigrations(ctx context.Context, conn *pgx.Conn, migrationsDir string) error {
	db := stdlib.OpenDB(*conn.Config())
	defer db.Close()

	return goose.Up(db, migrationsDir)
}

func dbStringConverter(cfg *config.Config) string {
	urlConn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DBName,
	)

	return urlConn
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error) {
	const op = "postgres.SaveUser"

	query := `
		INSERT INTO users (email, password) 
		VALUES ($1, $2) 
		RETURNING id;
		`

	err = s.db.QueryRow(ctx, query, email, passHash).Scan(&uid)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, auth.ErrUserExists
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return uid, nil
}

func (s *Storage) ProvideUser(ctx context.Context, email string) (*models.User, error) {
	const op = "postgres.ProvideUser"

	query := `
		SELECT users.id,
		users.password
		FROM users
		WHERE email = $1
		LIMIT 1;
		`
	var user models.User

	err := s.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.PasswordHashed)
	if errors.Is(err, pgx.ErrNoRows) {
		return &user, auth.ErrUserNotFound
	} else if err != nil {
		return &user, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) ListUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	const op = "postgres.ListUsers"

	query := `
		SELECT id, email, created_at, is_active
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;
	`

	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt, &user.IsActive); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return users, nil
}

func (s *Storage) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	const op = "postgres.GetUserByID"

	query := `
		SELECT id, email, password, created_at, is_active
		FROM users
		WHERE id = $1
		LIMIT 1;
	`

	var user models.User
	err := s.db.QueryRow(ctx, query, userID).Scan(&user.ID, &user.Email, &user.PasswordHashed, &user.CreatedAt, &user.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, auth.ErrUserNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) BlockUserByID(ctx context.Context, userID int64) error {
	const op = "postgres.BlockUserByID"

	query := `
		UPDATE users
		SET is_active = false
		WHERE id = $1;
	`

	cmdTag, err := s.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return auth.ErrUserNotFound
	}

	return nil
}

func (s *Storage) GetUserMessages(ctx context.Context, userID int64, limit, offset int) ([]models.Message, error) {
	const op = "postgres.GetUserMessages"

	query := `
		SELECT id, user_id, message_text, sender_type, created_at
		FROM user_messages
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3;
	`

	rows, err := s.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.UserID, &message.MessageText, &message.SenderType, &message.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return messages, nil
}

func (s *Storage) SaveMessage(ctx context.Context, message models.Message) error {
	query := `
		INSERT INTO user_messages (user_id, message_text, sender_type, created_at)
		VALUES ($1, $2, $3, $4);
	`
	_, err := s.db.Exec(ctx, query, message.UserID, message.MessageText, message.SenderType, message.CreatedAt)
	return err
}
