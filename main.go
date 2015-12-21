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
	}

	xsd, err := os.Open("tmp/tmp.xsd")
	if err != nil {
		log.Fatalf("failed to read xsd file: %s", err)
	}
	defer xsd.Close()

	schema := mon.NewSchema("etr", "probe ETR-290 checks")
	if err = mon.AddSchema(db, schema, xsd); err != nil {
		log.Fatalf("failed to install schema: %s", err)
	}

	doc := mon.NewDoc("hw4_172_etr", "etr",
		"http://10.0.30.172/probe/data/AnaEtrDetails?inputId=0&"+
			"etrEngineNo=0&detailsId=-1&showDisabledChecks=true",
		60, 86400)
	if err = mon.AddDoc(db, doc); err != nil {
		log.Fatalf("failed to add document: %s", err)
	}
}
