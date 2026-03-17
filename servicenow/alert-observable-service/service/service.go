package service

import (
	"context"

	"github.com/org/alert-observable-service/repository"
)

// TxRunner runs repository operations in a transaction.
type TxRunner interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context, tx *repository.Tx) error) error
}

// Service holds business logic dependencies.
type Service struct {
	AlertRuleRepo        repository.AlertRuleRepo
	AlertRepo            repository.AlertRepo
	ObservableRepo       repository.ObservableRepo
	CaseObservableRepo   repository.CaseObservableRepo
	ChildObservableRepo  repository.ChildObservableRepo
	SimilarIncidentRepo  repository.SimilarIncidentRepo
	TxRunner             TxRunner
	Audit                AuditPublisher
}

// Actor holds identity and tracing from gRPC metadata.
type Actor struct {
	UserID       string
	UserName     string
	RequestID    string
	CorrelationID string
}
