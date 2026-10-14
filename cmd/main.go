package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"workspaces/internal/platform/config"
	"workspaces/internal/platform/database"
	"workspaces/internal/platform/httpserver"
	"workspaces/internal/platform/module"
	"workspaces/projects"
)

const moduleShutdownTimeout = 20 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run вынесен из main: log.Fatal завершает процесс, не выполняя defer,
// а здесь все defer (закрытие БД, остановка модулей) сработают до выхода.
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db := database.GetConnection(&cfg)
	defer db.Close()

	deps := module.Deps{Cfg: &cfg, DB: db}

	modules := []module.Module{
		projects.New(deps),
	}

	r := httpserver.NewRouter("projects-api")
	r.Route("/api/v1", func(r chi.Router) {
		for _, m := range modules {
			log.Printf("registering module: %s", m.Name())
			m.Routes(r)
		}
	})

	runErr := httpserver.Run(ctx, fmt.Sprintf(":%d", cfg.APIPort), r)

	// HTTP остановлен, новых подключений нет. Закрываем сессии модулей и ждём,
	// пока они уберут за собой (контейнеры, статусы в БД). Выполняется до db.Close().
	shutdownCtx, cancel := context.WithTimeout(context.Background(), moduleShutdownTimeout)
	defer cancel()
	for _, m := range modules {
		s, ok := m.(module.Shutdowner)
		if !ok {
			continue
		}
		log.Printf("shutting down module: %s", m.Name())
		if err := s.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown %s: %v", m.Name(), err)
		}
	}

	return runErr
}
