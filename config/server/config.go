package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type ServerCfg struct {
	ServerPort         string        `env:"SERVER_PORT" envDefault:"55441"`
	ServerHost         string        `env:"SERVER_HOST" envDefault:"localhost"`
	PasswordSecret     string        `env:"PASSWORD_SECRET" envDefault:"default-password-secret-change-in-production"`
	JWTSecret          string        `env:"JWT_SECRET" envDefault:"default-jwt-secret-change-in-production"`
	AccessTokenExpiry  time.Duration `env:"ACCESS_TOKEN_EXPIRY" envDefault:"15m"`
	RefreshTokenExpiry time.Duration `env:"REFRESH_TOKEN_EXPIRY" envDefault:"24h"`
	S3endpoint         string        `env:"S3_ENDPOINT" envDefault:"localhost:9000"`
	S3accessKey        string        `env:"S3_ACCESS_KEY" envDefault:"minioadmin"`
	S3secretKey        string        `env:"S3_SECRET_KEY" envDefault:"minioadmin"`
	S3bucket           string        `env:"S3_BUCKET" envDefault:"your-bucket"`
	S3region           string        `env:"S3_REGION" envDefault:"us-east-1"`
	S3useSSL           bool          `env:"S3_USE_SSL" envDefault:"false"`
	S3forcePathStyle   bool          `env:"S3_FORCE_PATH_STYLE" envDefault:"true"`
	DBConnectionString string        `env:"DB_CONNECTION_STRING" envDefault:"postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"`
	MigrationsPath     string        `env:"MIGRATIONS_PATH" envDefault:"file://./migrations"`
}

func NewConfig() (*ServerCfg, error) {
	cfg := &ServerCfg{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *ServerCfg) GetConfig() ServerCfg {
	return *c
}
