package main

import (
	"fmt"
	"os"
)

func main() {
	conn, err := NewDBConnection(
		"user=datamon password=datamon dbname=datamon sslmode=disable")
	if err != nil {
		fmt.Printf("failure: %s\n", err)
		return
	}
	defer conn.Close()

	if err = UninstallDatamon(conn); err != nil {
		fmt.Printf("failure: %s\n", err)
		return
	}

	if err = InstallDatamon(conn); err != nil {
		fmt.Printf("failure: %s\n", err)
		return
	}

	file, err := os.Open("test.xsd")
	if err != nil {
		fmt.Printf("failure: %s\n", err)
		return
	}
	defer file.Close()

	if err = InstallSchema(conn, "test", "blabla", file); err != nil {
		fmt.Printf("failure: %s\n", err)
		return
	}

	fmt.Printf("ok\n")
}
