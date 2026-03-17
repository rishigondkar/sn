package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/servicenow/case-service/api"
	"github.com/servicenow/case-service/audit"
	"github.com/servicenow/case-service/config"
	"github.com/servicenow/case-service/handler"
	"github.com/servicenow/case-service/logging"
	"github.com/servicenow/case-service/repository"
	"github.com/servicenow/case-service/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/servicenow/case-service/proto/case_service"
)

func main() {
	_ = logging.Setup()
	cfg := config.Load()

	ctx := context.Background()
	if cfg.DBURL == "" {
		fmt.Fprintf(os.Stderr, "DATABASE_URL is required\n")
		os.Exit(1)
	}
	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "database ping: %v\n", err)
		os.Exit(1)
	}
	repo := repository.New(pool)
	var auditPub audit.Publisher = &audit.NoopPublisher{}
	if cfg.AuditServiceAddr != "" {
		// LazyPublisher connects on first Publish and reconnects after connection errors,
		// so audit events are sent even if the audit service wasn't ready at startup.
		auditPub = audit.NewLazyPublisher(cfg.AuditServiceAddr)
	}
	svc := service.New(repo, auditPub)
	h := handler.New(svc)
	apiHandler := api.New(svc)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.LoggingInterceptor,
		),
	)
	pb.RegisterCaseServiceServer(grpcServer, h)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		fmt.Fprintf(os.Stderr, "grpc listen: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      apiHandler.Handler(),
		ReadTimeout:  cfg.ServerTimeout,
		WriteTimeout: cfg.ServerTimeout,
	}

	go func() {
		if err := grpcServer.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			fmt.Fprintf(os.Stderr, "grpc serve: %v\n", err)
		}
	}()
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "http serve: %v\n", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownGrace)
	defer cancel()
	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(shutdownCtx)
	time.Sleep(100 * time.Millisecond)
}
