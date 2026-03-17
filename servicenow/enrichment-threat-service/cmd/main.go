package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/servicenow/enrichment-threat-service/api"
	"github.com/servicenow/enrichment-threat-service/config"
	"github.com/servicenow/enrichment-threat-service/handler"
	"github.com/servicenow/enrichment-threat-service/logging"
	"github.com/servicenow/enrichment-threat-service/proto/enrichment_threat_service"
	"github.com/servicenow/enrichment-threat-service/repository"
	"github.com/servicenow/enrichment-threat-service/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
)

func main() {
	cfg := config.Load()
	logger := logging.Setup()
	_ = logger

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		slog.Error("database connection failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		slog.Error("database ping failed", "err", err)
		os.Exit(1)
	}

	repo := repository.NewRepository(pool)
	var auditPub service.AuditEventPublisher = &service.StubAuditPublisher{}
	if cfg.AuditServiceAddr != "" {
		auditPub = service.NewGRPCAuditPublisher(cfg.AuditServiceAddr)
	}
	svc := service.NewService(repo, auditPub, cfg)
	grpcHandler := handler.NewHandler(svc)
	restAPI := api.NewAPI(svc)

	// gRPC server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(logging.LoggingInterceptor),
	)
	enrichment_threat_service.RegisterEnrichmentThreatServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.GRPCPort))
	if err != nil {
		slog.Error("gRPC listen failed", "err", err)
		os.Exit(1)
	}
	go func() {
		slog.Info("gRPC server listening", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server error", "err", err)
		}
	}()

	// HTTP server (health + REST)
	httpHandler := restAPI.Handler()
	httpServer := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.HTTPPort),
		Handler:      httpHandler,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
	}
	go func() {
		slog.Info("HTTP server listening", "port", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "err", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP shutdown error", "err", err)
	}
	slog.Info("shutdown complete")
}
