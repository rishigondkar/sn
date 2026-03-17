package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/soc-platform/attachment-service/api"
	"github.com/soc-platform/attachment-service/audit"
	"github.com/soc-platform/attachment-service/config"
	"github.com/soc-platform/attachment-service/handler"
	"github.com/soc-platform/attachment-service/logging"
	"github.com/soc-platform/attachment-service/repository"
	"github.com/soc-platform/attachment-service/service"
	"github.com/soc-platform/attachment-service/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/soc-platform/attachment-service/proto/attachment_service"
)

func main() {
	_ = logging.Setup()
	cfg := config.Load()

	ctx := context.Background()
	if cfg.DatabaseURL == "" {
		panic("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		panic("database: " + err.Error())
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		panic("database ping: " + err.Error())
	}
	repo := repository.New(pool)

	store := storage.NewMemoryStore()
	var auditPub audit.Publisher = audit.NoopPublisher{}
	if cfg.AuditServiceAddr != "" {
		auditPub = audit.NewGRPCPublisher(cfg.AuditServiceAddr)
	}
	svc := &service.Service{
		Repo:     repo,
		Storage:  store,
		Audit:    auditPub,
		Bucket:   cfg.StorageBucket,
		MaxBytes: cfg.MaxFileSizeBytes,
		Allowed:  cfg.AllowedTypes,
		Denied:   cfg.DeniedTypes,
	}

	grpcHandler := &handler.Handler{Service: svc}
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(logging.LoggingInterceptor),
	)
	pb.RegisterAttachmentServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	restAPI := api.New(svc)
	httpHandler := restAPI.Handler()
	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      httpHandler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		panic("grpc listen: " + err.Error())
	}
	defer lis.Close()

	go func() {
		_ = grpcServer.Serve(lis)
	}()
	go func() {
		_ = httpServer.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(shutdownCtx)
}
