package main

import (
	config "github.com/Maxim-Ba/information-keeper/config/client"
	"github.com/Maxim-Ba/information-keeper/internal/client"
	"github.com/Maxim-Ba/information-keeper/internal/client/tui"
	"github.com/Maxim-Ba/information-keeper/pkg/logger"
)

func main() {
	cfg := config.NewConfig()
	lconfig := logger.DefaultConfig()
	lconfig.FilePath = "myapp.log" // TODO : change to config
	lconfig.Level = logger.LevelDebug

	if err := logger.InitLogger(lconfig); err != nil {
		panic(err)
	}
	
	clnt, err := client.NewGRPCClient(cfg.GetConfig().ServerHost + ":" + cfg.GetConfig().ServerPort)
	if err != nil {
		panic(err)
	}
	defer clnt.Close()

	tui.StartTUI(clnt)

}
