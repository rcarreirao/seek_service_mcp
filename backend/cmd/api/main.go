// Command api is the REST entrypoint of the Seek Service application.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/rcarreirao/seek_service_mcp/internal/shared/config"
	"github.com/rcarreirao/seek_service_mcp/internal/shared/database"
	"github.com/rcarreirao/seek_service_mcp/internal/shared/httpx"
)

func main() {
	app := fx.New(
		fx.Provide(
			config.Load,
			newLogger,
			newDatabase,
			newRouter,
			newServer,
		),
		fx.Invoke(registerHooks),
	)
	app.Run()
}

func newLogger() *slog.Logger {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	return log.With(slog.String("component", "api"))
}

func newDatabase(lc fx.Lifecycle, cfg config.Config) (*gorm.DB, error) {
	db, err := database.Open(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return database.AutoMigrate(ctx, db)
		},
		OnStop: func(ctx context.Context) error {
			return database.Close(db)
		},
	})
	return db, nil
}

func newRouter(log *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(httpx.RequestID)
	r.Use(httpx.BodyLimit)
	r.Use(httpx.Recovery(log))
	r.Use(httpx.Metrics)
	r.Use(httpx.Logging(log))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Method(http.MethodGet, "/metrics", promhttp.Handler())
	return r
}

func newServer(lc fx.Lifecycle, cfg config.Config, handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", cfg.HTTPAddr)
			if err != nil {
				return err
			}
			go func() {
				if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
					slog.Error("server stopped", slog.String("error", err.Error()))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
	return srv
}

func registerHooks(lc fx.Lifecycle, log *slog.Logger, cfg config.Config, _ *http.Server, _ *gorm.DB) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("api starting",
				slog.String("addr", cfg.HTTPAddr),
			)
			return nil
		},
	})
}