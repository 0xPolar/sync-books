package main

import (
	"go-currently-reading/internal/db"
	"log"
)

func main() {
	println("Hello, world!")

	db, err := db.NewDB("./test.db")
	if err != nil {
		log.Fatal(err)
	}

	db.Close()

}
