package main

import (
	"github.com/alphabatem/btc_rune/db"
	"github.com/alphabatem/btc_rune/services"
	"github.com/babilu-online/common/context"
	"github.com/joho/godotenv"
	"log"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx, err := context.NewCtx(
		&db.SqliteService{},
		&services.DatabaseService{},
		&services.BTCService{},
		&services.RuneService{},
		&services.ChainSyncService{},
		&services.HttpService{},
	)

	if err != nil {
		log.Fatal(err)
		return
	}

	err = ctx.Run()
	log.Fatal(err)
}
