package db

import (
	"errors"
	"fmt"

	"github.com/cloakd/common/context"
	"github.com/cloakd/common/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

type SqliteService struct {
	services.DefaultService
	db *gorm.DB

	username string
	password string
	database string
	host     string
}

const SQLITE_SVC = "sqlite_svc"

// Id returns Service ID
func (ds SqliteService) Id() string {
	return SQLITE_SVC
}

// Db Access to raw SqliteService db
func (ds SqliteService) Db() *gorm.DB {
	return ds.db
}

// Configure the service
func (ds *SqliteService) Configure(ctx *context.Context) error {
	ds.database = fmt.Sprintf("%s", os.Getenv("DB_DATABASE"))

	return ds.DefaultService.Configure(ctx)
}

// Start the service and open connection to the database
// Migrate any tables that have changed since last runtime
func (ds *SqliteService) Start() (err error) {
	ds.db, err = gorm.Open(sqlite.Open(ds.database), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Error),
		PrepareStmt: true,
	})
	if err != nil {
		return err
	}

	return nil
}

// Find returns the db query for a statement
func (ds *SqliteService) Find(out interface{}, where string, args ...interface{}) error {
	return ds.error(ds.db.Find(out, where, args).Error)
}

// Create a new item in the SqliteService
func (ds *SqliteService) Create(val interface{}) (interface{}, error) {
	err := ds.db.Create(val).Error
	return val, ds.error(err)
}

// Update an existing item
func (ds *SqliteService) Update(old interface{}, new interface{}) (interface{}, error) {
	err := ds.db.Model(old).Updates(new).Error
	return new, ds.error(err)
}

// Delete an existing item
func (ds *SqliteService) Delete(val interface{}) error {
	err := ds.db.Delete(val).Error
	return ds.error(err)
}

// Migrate creates any new tables needed
func (ds *SqliteService) Migrate(values ...interface{}) error {
	err := ds.db.AutoMigrate(values).Error()
	if err != "" {
		return errors.New(err)
	}
	return nil
}

// Shutdown Gracefully close the database connection
func (ds *SqliteService) Shutdown() {
	//
}

// Parse an error returned from the database into a more contextual error that can be used with http response codes
func (ds *SqliteService) error(err error) error {
	if err == nil {
		return nil
	}

	var code int

	switch err {
	case gorm.ErrRecordNotFound:
		code = 404
		break
	default:
		code = 500
	}

	log.Println(code) //TODO implement
	return err
}
