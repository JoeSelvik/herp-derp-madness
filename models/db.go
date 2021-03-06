package models

import (
	"database/sql"
	"log"
	"os"
)

type DB struct {
	*sql.DB
}

// NewDB returns a connection to new database.
//
// If the given name exists, rename it to *.old, overwriting any existing *.old db.
func NewDB(dbPath string) (*DB, error) {
	if _, err := os.Stat(dbPath); err == nil {
		log.Println("Previous test db existed, renaming to *.old")
		err := os.Rename(dbPath, dbPath+".old")
		if err != nil {
			log.Printf("Error when renaming db: %s\n", err)
			return nil, err
		}
	}
	return OpenDB(dbPath)
}

// OpenDB returns a connection to an existing db and returns a DB struct with an active handle.
func OpenDB(dbPath string) (*DB, error) {
	// sqlite setup and verification
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("Error when opening sqlite3: %s\n", err)
		return nil, err
	}
	if db == nil {
		log.Printf("db nil when opened")
		return nil, err
	}

	// sql.open may validate its arguments without creating a connection to the database
	// call Ping() to verify that the data source name is valid
	err = db.Ping()
	if err != nil {
		log.Printf("Error when pinging db: %s\n", err)
		return nil, err
	}

	return &DB{db}, nil
}
