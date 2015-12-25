package main

import (
	"btc/data"
	"btc/mon"
	"btc/xmls"
	"database/sql"
	"log"
	"os"
	"time"
)

func install(db *sql.DB) {
	var err error
	if err = mon.Install(db); err != nil {
		log.Fatalf("failed to install data monitor: %s", err)
	}

	var root *xmls.Element
	root, err = xmls.FromFile("tmp/etr.xsd")
	if err != nil {
		log.Fatalf("failed to create xml schema: %s", err)
	}

	schema := mon.NewSchema("etr", "probe ETR-290 checks")
	if err = mon.AddSchema(db, schema, root); err != nil {
		log.Fatalf("failed to install schema: %s", err)
	}

	doc := mon.NewDoc("hw4_172_etr", "etr",
		"http://10.0.30.172/probe/etrdata?inputId=0&tuningSetupId=1",
		60, 86400)
	if err = mon.AddDoc(db, doc); err != nil {
		log.Fatalf("failed to add document: %s", err)
	}
}

func commit(db *sql.DB) {
	file, err := os.Open("tmp/etr.xml")
	if err != nil {
		log.Fatalf("failed to open xml doc: %s", err)
	}
	defer file.Close()

	if err = mon.CommitDoc(db, "hw4_172_etr", file, false); err != nil {
		log.Fatalf("failed to commit doc: %s", err)
	}
}

func checkout(db *sql.DB) {
	timestamp, err := time.Parse(
		time.RFC3339, "2015-12-25T18:26:58+01:00")
	if err != nil {
		log.Fatalf("failed to parse timestamp: %s", err)
	}

	if err := mon.CheckoutDoc(
		db, "hw4_172_etr", timestamp,
		os.Stdout, " ", " "); err != nil {
		log.Fatalf("failed to checkout doc: %s", err)
	}
}

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

	//install(db)
	//commit(db)
	checkout(db)
}
