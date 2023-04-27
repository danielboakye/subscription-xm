package main

import (
	"database/sql"
	"encoding/gob"
	"log"
	"net/http"
	"os"
	"subscription-service/data"
	"sync"
	"time"

	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
)

func openDB() (*sql.DB, error) {
	dsn := os.Getenv("DSN")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initRedis() *redis.Pool {
	redisPool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}
	return redisPool
}

func initSession() *scs.SessionManager {
	gob.Register(data.User{})

	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true
	return session
}

func initApp() (Config, error) {
	// connect to the database
	conn, err := openDB()
	if err != nil {
		return Config{}, err
	}

	// create sessions
	session := initSession()

	// create loggers
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// init signer secret key
	NewURLSigner(os.Getenv("TOKEN_SECRET"))

	// create channels

	// create wait group
	wg := sync.WaitGroup{}

	// set up the application config
	app := Config{
		Session:       session,
		DB:            conn,
		InfoLog:       infoLog,
		ErrorLog:      errorLog,
		Wait:          &wg,
		Models:        data.New(conn),
		Mailer:        Mail{},
		ErrorChan:     make(chan error),
		ErrorChanDone: make(chan bool),
	}

	return app, nil
}
