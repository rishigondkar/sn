package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/servicenow/api-gateway/api"
	"github.com/servicenow/api-gateway/clients"
	"github.com/servicenow/api-gateway/config"
	"github.com/servicenow/api-gateway/logging"
	"github.com/servicenow/api-gateway/orchestrator"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	log := logging.NewLogger()

	// All downstream services required; no stubs. Dial each and exit on first failure.
	caseCmd, caseQuery, closeCase, err := clients.NewCaseClients(cfg.CaseServiceAddr, cfg.DownstreamTimeout)
	if err != nil {
		log.Error("case service dial failed", "addr", cfg.CaseServiceAddr, "error", err)
		os.Exit(1)
	}

	ref, closeRef, err := clients.NewReferenceClients(cfg.ReferenceServiceAddr, cfg.DownstreamTimeout)
	if err != nil {
		log.Error("reference service dial failed", "addr", cfg.ReferenceServiceAddr, "error", err)
		closeCase()
		os.Exit(1)
	}

	audit, closeAudit, err := clients.NewAuditClients(cfg.AuditServiceAddr, cfg.DownstreamTimeout)
	if err != nil {
		log.Error("audit service dial failed", "addr", cfg.AuditServiceAddr, "error", err)
		closeCase()
		closeRef()
		os.Exit(1)
	}

	attQuery, attCmd, closeAtt, err := clients.NewAttachmentClients(cfg.AttachmentServiceAddr, cfg.DownstreamTimeout)
	if err != nil {
		log.Error("attachment service dial failed", "addr", cfg.AttachmentServiceAddr, "error", err)
		closeCase()
		closeRef()
		closeAudit()
		os.Exit(1)
	}

	enrichment, threatLookup, closeEnrichment, err := clients.NewEnrichmentClients(cfg.EnrichmentServiceAddr, cfg.DownstreamTimeout)
	if err != nil {
		log.Error("enrichment service dial failed", "addr", cfg.EnrichmentServiceAddr, "error", err)
		closeCase()
		closeRef()
		closeAudit()
		closeAtt()
		os.Exit(1)
	}

	obsCmd, obsQuery, closeObs, err := clients.NewObservableClients(cfg.ObservableServiceAddr, cfg.DownstreamTimeout)
	if err != nil {
		log.Error("observable service dial failed", "addr", cfg.ObservableServiceAddr, "error", err)
		closeCase()
		closeRef()
		closeAudit()
		closeAtt()
		closeEnrichment()
		os.Exit(1)
	}

	orch := orchestrator.New(caseCmd, caseQuery, obsCmd, obsQuery, enrichment, threatLookup, ref, attQuery, attCmd, audit)
	handler := api.NewHandler(orch)

	chain := handler.Router()
	chain = api.Logging(log)(chain)
	chain = api.Auth(chain)
	chain = api.CorrelationID(chain)
	chain = api.RequestID(chain)
	chain = api.CORS(cfg.CORSAllowedOrigins, cfg.CORSAllowedMethods)(chain)
	chain = api.Recovery(log)(chain)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      chain,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down")
	closeObs()
	closeEnrichment()
	closeAtt()
	closeAudit()
	closeRef()
	closeCase()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", "error", err)
		os.Exit(1)
	}
	log.Info("shutdown complete")
}
