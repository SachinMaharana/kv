package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/sachinmaharana/kv/db"
)

type config struct {
	port int
	db   struct {
		dsn string
	}
}
type application struct {
	config config
	logger *log.Logger
	db     interface {
		Get(string) (string, error)
		Set(string, interface{}) error
		Search(string) ([]string, error)
		Total() int
	}
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "my-release-redis-master.default.svc.cluster.local:6379", "Redis DSN")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	client, err := createClient(cfg)

	db := db.NewRedisRepository(client)

	app := &application{
		config: cfg,
		logger: logger,
		db:     db,
	}

	if err != nil {
		logger.Fatal(err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	logger.Printf("starting server on %s", srv.Addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	<-shutdown

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Server Stopped")
	os.Exit(0)
}

func createClient(cfg config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.db.dsn,
		// Password: "cQ2FfYrN2E",
		DB: 0,
	})
	if err := client.Ping().Err(); err != nil {
		return nil, err
	}
	return client, nil
}
