package storage

import (
	"GoMonitor/internal/config"
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/url"
	"strings"
	"time"
)

type Postgres struct {
	pool *pgxpool.Pool
}

// Инициализация подключения
func NewPostgres(cfg config.DataBaseConfig) (*Postgres, error) {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   cfg.Name,
	}
	connStr := u.String()

	pool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	log.Printf("✅ Connected to PostgreSQL: %s:%s", cfg.Host, cfg.Port)
	return &Postgres{pool: pool}, nil
}

func (pg *Postgres) Close() {
	if pg.pool != nil {
		pg.pool.Close()
		log.Println("✅ PostgreSQL connection closed")
	}
}

// AddServerIfNotExist добавляет сервер в БД, если его там нет
func (pg *Postgres) AddServerIfNotExist(ctx context.Context, name, ip, description string) (int, error) {
	var id int

	log.Printf("[AddServerIfNotExist] Проверка существующего сервера: name=%s, ip=%s", name, ip)

	// Сначала пытаемся найти существующий сервер
	err := pg.pool.QueryRow(ctx,
		`SELECT id FROM servers WHERE name = $1 AND ip = $2`,
		name, ip).Scan(&id)

	if err == nil {
		log.Printf("[AddServerIfNotExist] Сервер уже существует: id=%d, name=%s, ip=%s", id, name, ip)
		return id, nil
	}

	log.Printf("[AddServerIfNotExist] Сервер не найден, добавляем новый: name=%s, ip=%s", name, ip)

	// Если сервер не найден, добавляем новый
	err = pg.pool.QueryRow(ctx,
		`INSERT INTO servers (name, ip, description)
         VALUES ($1, $2, $3)
         RETURNING id`,
		name, ip, description).Scan(&id)

	if err != nil {
		// Если возникла ошибка конфликта (дубликат), пытаемся снова найти
		if isDuplicateKeyError(err) {
			log.Printf("[AddServerIfNotExist] Обнаружен дубликат, повторный поиск: name=%s", name)
			return pg.AddServerIfNotExist(ctx, name, ip, description)
		}
		log.Printf("[AddServerIfNotExist] Ошибка при вставке сервера: %v", err)
		return 0, fmt.Errorf("insert server: %w", err)
	}

	log.Printf("[AddServerIfNotExist] Новый сервер добавлен: id=%d, name=%s, ip=%s", id, name, ip)
	return id, nil
}

// SaveMetric сохраняет метрику в базу данных
func (pg *Postgres) SaveMetric(ctx context.Context, objectType string, objectID int, metricName string, value float64, ts time.Time) error {
	_, err := pg.pool.Exec(ctx,
		`INSERT INTO metrics (object_type, object_id, metric_name, value, timestamp)
         VALUES ($1, $2, $3, $4, $5)`,
		objectType, objectID, metricName, value, ts)

	if err != nil {
		log.Printf("[SaveMetric] Ошибка сохранения метрики: object_type=%s, object_id=%d, metric=%s, error=%v",
			objectType, objectID, metricName, err)
		return fmt.Errorf("save metric: %w", err)
	}

	return nil
}

// Проверка на ошибку дубликата ключа в PostgreSQL
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()
	// Проверяем различные варианты ошибок дубликата
	return strings.Contains(errorStr, "duplicate key value") ||
		strings.Contains(errorStr, "23505") || // PostgreSQL error code for unique violation
		strings.Contains(errorStr, "already exists")
}

// Дополнительные методы для проверки состояния

// CheckServerExists проверяет существование сервера по ID
func (pg *Postgres) CheckServerExists(ctx context.Context, serverID int) (bool, error) {
	var exists bool
	err := pg.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM servers WHERE id = $1)`,
		serverID).Scan(&exists)
	return exists, err
}

// GetServerID получает ID сервера по имени и IP
func (pg *Postgres) GetServerID(ctx context.Context, name, ip string) (int, error) {
	var id int
	err := pg.pool.QueryRow(ctx,
		`SELECT id FROM servers WHERE name = $1 AND ip = $2`,
		name, ip).Scan(&id)
	return id, err
}
