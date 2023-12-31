package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/poggerr/gophermart/internal/app"
	"github.com/poggerr/gophermart/internal/async"
	"github.com/poggerr/gophermart/internal/config"
	"github.com/poggerr/gophermart/internal/logger"
	"github.com/poggerr/gophermart/internal/routers"
	"github.com/poggerr/gophermart/internal/server"
	"github.com/poggerr/gophermart/internal/storage"
	"log"
)

func main() {

	cfg := config.NewConf()

	db, err := sqlx.Connect("postgres", cfg.DB)
	if err != nil {
		log.Fatalln(err)
	}
	sugaredLogger := logger.Initialize()

	strg := storage.NewStorage(db, cfg)

	newRepo := async.NewRepo(strg)
	go newRepo.WorkerAccrual()

	strg.RestoreDB()

	newApp := app.NewApp(cfg, strg, sugaredLogger, newRepo)
	newApp.AccrualRestore()

	r := routers.Router(newApp)
	server.Server(cfg.ServAddr, r)

}
