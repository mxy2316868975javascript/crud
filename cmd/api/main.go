package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mxy2316868975javascript/crud/internal/config"
	"github.com/mxy2316868975javascript/crud/internal/httpapi"
	"github.com/mxy2316868975javascript/crud/internal/repository/postgres"
	"github.com/mxy2316868975javascript/crud/internal/service"
)

// main 读取配置、连接数据库、装配各层并启动 HTTP 服务。
func main() {
	configuration := config.Load()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, configuration.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := postgres.EnsureSchema(ctx, pool); err != nil {
		log.Fatal(err)
	}

	store := postgres.New(pool)
	router := httpapi.NewMux(
		service.NewWordService(store),
		service.NewTestService(store),
		store,
	)

	server := &http.Server{
		Addr:              ":" + configuration.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("listening on :%s", configuration.Port)
	log.Fatal(server.ListenAndServe())
}
