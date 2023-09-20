package model

import "database/sql"

type Database struct {
	DB *sql.DB
}
