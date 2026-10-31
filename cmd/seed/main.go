package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mxy2316868975javascript/crud/internal/config"
	"github.com/mxy2316868975javascript/crud/internal/repository/postgres"
	"github.com/mxy2316868975javascript/crud/internal/seed"
)

// main 读取配置、确保 schema 存在，然后导入内嵌词库。
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

	count, err := seed.Import(ctx, postgres.New(pool))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("seeded %d words", count)
}
