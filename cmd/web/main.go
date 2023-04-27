package main

import (
	"log"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {

	app, err := initApp()
	if err != nil {
		log.Panic("can't connect to postgres")
	}

	// setup mailer
	app.Mailer = app.crateMail()
	go app.listenForMail()

	// listen for shutdown signals
	go app.listenForShutdown()

	// listen for errors
	go app.listenForErrors()

	// listen for web connections
	app.serve()
}
