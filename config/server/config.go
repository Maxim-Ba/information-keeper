package config

import "time"

type ServerCfg struct {
	ServerPort         string
	ServerHost         string
	PasswordSecret     string
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

func NewConfig() *ServerCfg {
	return &ServerCfg{
		ServerPort: "55441",
		ServerHost: "localhost",
	}
}

func (c *ServerCfg) GetConfig() ServerCfg {
	return *c
}
