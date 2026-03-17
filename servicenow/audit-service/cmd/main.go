package main

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/servicenow/audit-service/config"
	"github.com/servicenow/audit-service/consumer"
	"github.com/servicenow/audit-service/handler"
	"github.com/servicenow/audit-service/logging"
	"github.com/servicenow/audit-service/proto/audit_service"
	"github.com/servicenow/audit-service/repository"
	"github.com/servicenow/audit-service/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	_ = logging.SetupLogging(ctx)

	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		panic("database ping: " + err.Error())
	}

	repo := repository.NewRepository(db)
	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	// gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(logging.LoggingInterceptor),
	)
	audit_service.RegisterAuditServiceServer(grpcServer, h)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.GRPCPort))
	if err != nil {
		panic(err)
	}
	go func() {
		if err := grpcServer.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			panic(err)
		}
	}()

	// HTTP server (health only)
	healthHandler := healthHandler(db)
	httpServer := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.HTTPPort),
		Handler:      healthHandler,
		ReadTimeout: cfg.HTTPServerTimeout,
		IdleTimeout: 2 * cfg.HTTPServerTimeout,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	// Event consumer (POC: channel source; production would use SQS/Kafka)
	msgSource := consumer.NewChannelSource(cfg.ConsumerBatchSize)
	cons := &consumer.Consumer{
		Repo:        repo,
		Source:      msgSource,
		MaxRetries:  cfg.ConsumerMaxRetries,
		RetryDelay:  cfg.ConsumerRetryDelay,
	}
	consumerCtx, stopConsumer := context.WithCancel(context.Background())
	go cons.Run(consumerCtx)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	stopConsumer()
	msgSource.Close()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(shutdownCtx)
}

func healthHandler(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if db != nil {
			if err := db.Ping(); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"status":"unavailable","reason":"database"}`))
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	return mux
}

