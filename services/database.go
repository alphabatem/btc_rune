package services

import (
	"github.com/alphabatem/btc_rune"
	"github.com/alphabatem/btc_rune/db"
	"github.com/cloakd/common/services"
)

type DatabaseService struct {
	services.DefaultService

	dbSvc *db.SqliteService
}

const DATABASE_SVC = "database_svc"

func (svc DatabaseService) Id() string {
	return DATABASE_SVC
}

func (svc *DatabaseService) Start() error {
	svc.dbSvc = svc.Service(db.SQLITE_SVC).(*db.SqliteService)

	err := svc.dbSvc.Db().AutoMigrate(&btc_rune.Rune{})
	if err != nil {
		return err
	}

	return nil
}

func (svc *DatabaseService) Function1() error {

	return nil
}
