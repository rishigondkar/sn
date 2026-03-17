package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"

	"github.com/servicenow/assignment-reference-service/api"
	"github.com/servicenow/assignment-reference-service/config"
	"github.com/servicenow/assignment-reference-service/handler"
	"github.com/servicenow/assignment-reference-service/logging"
	"github.com/servicenow/assignment-reference-service/repository"
	"github.com/servicenow/assignment-reference-service/service"

	pb "github.com/servicenow/assignment-reference-service/proto/assignment_reference_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()
	_, _ = logging.SetupLogging(ctx)

	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db open: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "db ping: %v\n", err)
		os.Exit(1)
	}

	repo := repository.New(db, cfg.DBTimeout)
	if err := repo.SeedPOCIfEmpty(ctx); err != nil {
		// Non-fatal: table may not exist yet (run migrations first); or DB may already be seeded
		fmt.Fprintf(os.Stderr, "seed POC (if empty): %v\n", err)
	}
	var auditPub service.AuditPublisher = service.NoopAuditPublisher{}
	if cfg.AuditServiceAddr != "" {
		auditPub = service.NewGRPCAuditPublisher(cfg.AuditServiceAddr)
	}
	svc := service.New(repo, auditPub)
	grpcHandler := handler.New(svc)
	apiServer := api.NewServer(svc)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(logging.LoggingInterceptor),
	)
	pb.RegisterAssignmentReferenceServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grpc listen: %v\n", err)
		os.Exit(1)
	}
	defer lis.Close()

	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			fmt.Fprintf(os.Stderr, "grpc serve: %v\n", err)
		}
	}()

	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      apiServer.Handler(),
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "http serve: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(shutdownCtx)
}
