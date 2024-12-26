package main

import (
	"log"

	"github.com/umeh-promise/social/internal/db"
	"github.com/umeh-promise/social/internal/env"
	"github.com/umeh-promise/social/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://user:password@localhost:5432/social?sslmode=disable")

	conn, err := db.New(addr, 10, 10, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	store := store.NewStore(conn)
	db.Seed(store)
}
