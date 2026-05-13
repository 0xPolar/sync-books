package main

import (
	"context"
	"go-currently-reading/internal/db"
	"go-currently-reading/internal/kavita"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func run(ctx context.Context) error {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	apiKey := os.Getenv("KAVITA_KEY")

	db, err := db.NewDB("./books.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	client, err := kavita.New("https://kavita.aoaknode.xyz", apiKey, "sync-books")
	if err != nil {
		log.Fatal(err)
	}

	err = client.Authenticate(ctx)
	if err != nil {
		log.Fatal(err)
	}

	me, err := client.Me(ctx)
	if err != nil {
		log.Fatal(err)
	}

	println("Hello, " + me.Username)

	books, err := client.CurrentlyReadingBooks(ctx, 0, 0, time.Time{})
	if err != nil {
		log.Fatal(err)
	}

	if err := db.UpsertBooks(books); err != nil {
		log.Fatal(err)
	}

	for _, book := range books {
		log.Printf("Book: %s, ISBN: %s, Authors: %v, Progress: %.2f%%\n", book.Title, book.ISBN, book.Authors, book.ProgressPct)
	}

	return nil
}

func main() {
	println("Hello, world!")
	ctx := context.Background()

	err := run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
