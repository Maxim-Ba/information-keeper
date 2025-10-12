package config

type ClientCfg struct {
	ServerPort string
	ServerHost string
}

func NewConfig() *ClientCfg {
	return &ClientCfg{
		ServerPort: "55441",
		ServerHost: "localhost",
	}
}

func (c *ClientCfg) GetConfig() ClientCfg {
	return *c
}
