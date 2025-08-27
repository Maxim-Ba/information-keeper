package config

type ServerCfg struct {
	ServerPort     string
	ServerHost     string
	PasswordSecret string
}

func NewConfig() *ServerCfg {
	return &ServerCfg{}
}

func (c *ServerCfg) GetConfig() ServerCfg {
	return *c
}
