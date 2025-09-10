package config

import "time"

type ServerCfg struct {
	ServerPort         string
	ServerHost         string
	PasswordSecret     string
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	S3endpoint         string
	S3accessKey        string
	S3secretKey        string
	S3bucket           string
	S3region           string
	S3useSSL           bool
	S3forcePathStyle   bool
	DBConnectionString string
	MigrationsPath  string
}

func NewConfig() *ServerCfg {
	return &ServerCfg{
		ServerPort: "55441",
		ServerHost: "localhost",
		DBConnectionString : "",
		MigrationsPath: "file://./migrations",
	}
}

func (c *ServerCfg) GetConfig() ServerCfg {
	return *c
}
