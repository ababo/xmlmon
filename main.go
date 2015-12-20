package main

import (
	"btc/data"
	"btc/mon"
	"database/sql"
	"log"
	"os"
)

func main() {
	config, err := NewConfig("config.json")
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	var db *sql.DB
	db, err = data.Open(config.DbConnStr)
	if err != nil {
		log.Fatalf("failed to establish db connection: %s", err)
	}
	defer db.Close()

	if err = mon.Install(db); err != nil {
		log.Fatalf("failed to install data monitor: %s", err)
		return
	}

	xsd, err := os.Open("tmp/tmp.xsd")
	if err != nil {
		log.Fatalf("failed to read xsd file: %s", err)
		return
	}
	defer xsd.Close()

	if err = mon.InstallSchema(db, "test", "blabla", xsd); err != nil {
		log.Fatalf("failed to install schema: %s", err)
		return
	}
}
