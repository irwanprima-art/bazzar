package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "bazzar"),
		JWTSecret:  getEnv("JWT_SECRET", "bazzar-makuku-secret-key-2026"),
	}
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func ConnectDB(cfg *Config) *pgxpool.Pool {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		log.Fatalf("Unable to parse database URL: %v", err)
	}

	poolConfig.MaxConns = 20

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	log.Println("✅ Database connected successfully")
	return pool
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
