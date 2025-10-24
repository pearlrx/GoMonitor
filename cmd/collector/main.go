package main

import (
	"GoMonitor/internal/collector"
	"GoMonitor/internal/config"
	"GoMonitor/internal/storage"
	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	cfgPath := "config.yaml"
	if p := os.Getenv("GM_CONFIG"); p != "" {
		cfgPath = p
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	dbCfg := config.DataBaseConfig{
		Host:       cfg.DataBase.Host,
		Port:       cfg.DataBase.Port,
		User:       cfg.DataBase.User,
		Password:   cfg.DataBase.Password,
		Name:       cfg.DataBase.Name,
		Migrations: cfg.DataBase.Migrations,
	}

	pg, err := storage.NewPostgres(dbCfg)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pg.Close()

	if err = pg.RunMigrations(context.Background(), "/app/migrations/001_init.sql"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	interval := time.Duration(cfg.Collector.IntervalSeconds) * time.Second
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// 1. Сначала добавляем все серверы в БД
	serverIDs := make(map[string]int)
	for _, srv := range cfg.Servers {
		id, err := pg.AddServerIfNotExist(ctx, srv.Name, srv.Address, srv.Description)
		if err != nil {
			log.Printf("failed to register server %s: %v", srv.Name, err)
			continue
		}
		serverIDs[srv.Name] = id
		log.Printf("✅ Server registered: %s -> ID %d", srv.Name, id)
	}

	// 2. Небольшая пауза чтобы убедиться, что все серверы добавлены
	time.Sleep(100 * time.Millisecond)

	// 3. Теперь запускаем сборщики метрик
	for _, srv := range cfg.Servers {
		id, exists := serverIDs[srv.Name]
		if !exists {
			log.Printf("⚠️  Skipping collector for server %s - not registered in DB", srv.Name)
			continue // пропускаем серверы, которые не удалось добавить
		}

		go func(serverID int, s config.Server) {
			c := collector.New(pg, "server", serverID, interval)
			log.Printf("[collector started] for %s (%s) with ID %d", s.Name, s.Address, serverID)
			c.Start(ctx)
		}(id, srv)
	}

	log.Printf("🎯 All collectors started. Monitoring %d servers", len(serverIDs))

	<-ctx.Done()
	log.Println("shutting down collectors...")
	time.Sleep(500 * time.Millisecond)
}
