package main

import (
	"log"

	"github.com/umeh-promise/social/internal/db"
	"github.com/umeh-promise/social/internal/env"
	"github.com/umeh-promise/social/internal/store"
)

const version = "1.0.0"

func main() {
	config := config{
		addr: env.GetString("ADDR", ":8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://user:password@localhost:5432/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
	}

	db, err := db.New(
		config.db.addr,
		config.db.maxOpenConns,
		config.db.maxIdleConns,
		config.db.maxIdleTime,
	)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Printf("DB connected successfully")

	store := store.NewStore(db)

	app := &application{
		config: config,
		store:  store,
	}

	router := app.mount()

	log.Fatal(app.run(router))
}
