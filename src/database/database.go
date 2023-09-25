package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	DB *sql.DB
}

func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", "file:"+path+"?cache=shared")
	db.SetMaxOpenConns(1)
	if err != nil {
		return nil, err
	}

	database := &Database{DB: db}

	err = database.RunMigrations()
	if err != nil {
		return nil, err
	}

	return database, nil
}

func (db *Database) RunMigrations() error {
	err := db.CreateExploitsTable()
	if err != nil {
		return err
	}
	err = db.CreateTargetsTable()
	if err != nil {
		return err
	}
	err = db.CreateFlagsTable()
	if err != nil {
		return err
	}
	return nil
}
