package repository

import (
	"Iu5-web/internal/app/minioClient"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Кастомные ошибки для слоя репозитория
var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("record with given parameters already exists")
	ErrNotAllowed    = errors.New("action is not allowed")
)

type Repository struct {
	db *gorm.DB
	mc *minioClient.Client
	rd *redis.Client
}

func New(dsn string, mc *minioClient.Client) (*Repository, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return nil, errors.New("REDIS_ADDR environment variable is not set")
	}

	rd := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if _, err := rd.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Repository{
		db: db,
		mc: mc,
		rd: rd,
	}, nil
}

// --- Работа с черным списком токенов в Redis ---

func (r *Repository) AddTokenToBlacklist(ctx context.Context, tokenString string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	key := "blacklist:" + tokenString
	return r.rd.Set(ctx, key, "1", ttl).Err()
}

func (r *Repository) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	key := "blacklist:" + tokenString
	val, err := r.rd.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}
