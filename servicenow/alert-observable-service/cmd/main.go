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

	"github.com/org/alert-observable-service/config"
	"github.com/org/alert-observable-service/handler"
	"github.com/org/alert-observable-service/logging"
	"github.com/org/alert-observable-service/repository"
	"github.com/org/alert-observable-service/service"
	pb "github.com/org/alert-observable-service/proto/alert_observable_service"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func main() {
	log := logging.NewLogger()
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DB.DSN())
	if err != nil {
		log.Error("failed to open db", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Error("db ping failed", "error", err)
		os.Exit(1)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	pg := repository.NewPostgres(db)
	ar, al, ob, co, ch, si := repository.NewRepos(pg)
	var auditPub service.AuditPublisher = &service.NoopAudit{}
	if cfg.AuditServiceAddr != "" {
		auditPub = service.NewGRPCAuditPublisher(cfg.AuditServiceAddr)
	}
	svc := &service.Service{
		AlertRuleRepo:       ar,
		AlertRepo:           al,
		ObservableRepo:      ob,
		CaseObservableRepo:  co,
		ChildObservableRepo: ch,
		SimilarIncidentRepo: si,
		TxRunner:            pg,
		Audit:               auditPub,
	}

	recoveryUnary := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if p := recover(); p != nil {
				log.Error("grpc panic recovered", "panic", p, "method", info.FullMethod)
				resp = nil
				err = status.Error(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
	grpcServer := grpc.NewServer(
		grpc.ConnectionTimeout(10*time.Second),
		grpc.UnaryInterceptor(recoveryUnary),
	)
	pb.RegisterAlertObservableServiceServer(grpcServer, handler.NewServer(svc))
	reflection.Register(grpcServer)

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	httpPort := cfg.HTTPPort
	if httpPort <= 0 || httpPort > 65535 {
		httpPort = 8082
	}
	grpcPort := cfg.GRPCPort
	if grpcPort <= 0 || grpcPort > 65535 {
		grpcPort = 50052
	}
	httpServer := &http.Server{Addr: ":" + strconv.Itoa(httpPort), Handler: httpMux}
	grpcAddr := ":" + strconv.Itoa(grpcPort)

	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			log.Error("grpc listen failed", "error", err, "addr", grpcAddr)
			os.Exit(1)
		}
		log.Info("grpc listening", "addr", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Error("grpc serve error", "error", err)
		}
	}()

	go func() {
		log.Info("http listening", "addr", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http serve error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	grpcServer.GracefulStop()
	_ = httpServer.Shutdown(ctx)
	log.Info("shutdown complete")
}
