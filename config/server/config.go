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
	return &ServerCfg{}
}

func (c *ServerCfg) GetConfig() ServerCfg {
	return *c
}
