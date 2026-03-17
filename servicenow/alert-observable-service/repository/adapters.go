package repository

import (
	"context"

	"github.com/google/uuid"
)

// Ensure Postgres implements all repo interfaces via adapters.

type alertRuleRepo struct{ *Postgres }
type alertRepo struct{ *Postgres }
type observableRepo struct{ *Postgres }
type caseObservableRepo struct{ *Postgres }
type childObservableRepo struct{ *Postgres }
type similarIncidentRepo struct{ *Postgres }

func (r *alertRuleRepo) Create(ctx context.Context, row *AlertRule) error {
	return r.CreateAlertRule(ctx, row)
}
func (r *alertRuleRepo) GetByID(ctx context.Context, id uuid.UUID) (*AlertRule, error) {
	return r.GetAlertRuleByID(ctx, id)
}

func (r *alertRepo) Create(ctx context.Context, a *Alert) error {
	return r.CreateAlert(ctx, a)
}
func (r *alertRepo) GetByID(ctx context.Context, id uuid.UUID) (*Alert, error) {
	return r.GetAlertByID(ctx, id)
}
func (r *alertRepo) ListByCaseID(ctx context.Context, caseID uuid.UUID, opts ListOpts) ([]*Alert, string, error) {
	return r.ListAlertsByCaseID(ctx, caseID, opts)
}

func (r *observableRepo) Create(ctx context.Context, o *Observable) error {
	return r.CreateObservable(ctx, o)
}
func (r *observableRepo) GetByID(ctx context.Context, id uuid.UUID) (*Observable, error) {
	return r.GetObservableByID(ctx, id)
}
func (r *observableRepo) GetByTypeAndNormalized(ctx context.Context, observableType, normalizedValue string) (*Observable, error) {
	return r.GetObservableByTypeAndNormalized(ctx, observableType, normalizedValue)
}
func (r *observableRepo) Update(ctx context.Context, o *Observable) error {
	return r.Postgres.UpdateObservable(ctx, o)
}
func (r *observableRepo) List(ctx context.Context, opts ListOpts, searchQuery *string) ([]*Observable, string, error) {
	return r.ListObservables(ctx, opts, searchQuery)
}

func (r *caseObservableRepo) Create(ctx context.Context, co *CaseObservable) error {
	return r.CreateCaseObservable(ctx, co)
}
func (r *caseObservableRepo) Update(ctx context.Context, co *CaseObservable) error {
	return r.UpdateCaseObservable(ctx, co)
}
func (r *caseObservableRepo) GetByID(ctx context.Context, id uuid.UUID) (*CaseObservable, error) {
	return r.GetCaseObservableByID(ctx, id)
}
func (r *caseObservableRepo) GetByCaseAndObservable(ctx context.Context, caseID, observableID uuid.UUID) (*CaseObservable, error) {
	return r.GetCaseObservableByCaseAndObservable(ctx, caseID, observableID)
}
func (r *caseObservableRepo) ListByCaseID(ctx context.Context, caseID uuid.UUID, opts ListOpts, filterType, filterStatus *string) ([]*CaseObservable, string, error) {
	return r.ListCaseObservablesByCaseID(ctx, caseID, opts, filterType, filterStatus)
}
func (r *caseObservableRepo) ListCaseObservablesWithDetailsByCaseID(ctx context.Context, caseID uuid.UUID, opts ListOpts, filterType, filterStatus *string) ([]*CaseObservableWithDetails, string, error) {
	return r.Postgres.ListCaseObservablesWithDetailsByCaseID(ctx, caseID, opts, filterType, filterStatus)
}
func (r *caseObservableRepo) CaseIDsByObservable(ctx context.Context, observableID uuid.UUID, opts ListOpts) ([]uuid.UUID, string, error) {
	return r.ListCaseIDsByObservable(ctx, observableID, opts)
}
func (r *caseObservableRepo) ObservableIDsByCaseID(ctx context.Context, caseID uuid.UUID) ([]uuid.UUID, error) {
	return r.GetObservableIDsByCaseID(ctx, caseID)
}
func (r *caseObservableRepo) DeleteByCaseAndObservable(ctx context.Context, caseID, observableID uuid.UUID) error {
	return r.DeleteCaseObservableByCaseAndObservable(ctx, caseID, observableID)
}

func (r *childObservableRepo) Create(ctx context.Context, c *ChildObservable) error {
	return r.CreateChildObservable(ctx, c)
}
func (r *childObservableRepo) GetByParentChildType(ctx context.Context, parentID, childID uuid.UUID, relationshipType string) (*ChildObservable, error) {
	return r.GetChildObservableByParentChildType(ctx, parentID, childID, relationshipType)
}
func (r *childObservableRepo) ListByParentID(ctx context.Context, parentID uuid.UUID, opts ListOpts) ([]*ChildObservable, string, error) {
	return r.ListChildObservablesByParentID(ctx, parentID, opts)
}

func (r *similarIncidentRepo) Upsert(ctx context.Context, s *SimilarIncident) error {
	return r.UpsertSimilarIncident(ctx, s)
}
func (r *similarIncidentRepo) DeleteByCaseAndSimilarCase(ctx context.Context, caseID, similarCaseID uuid.UUID) error {
	return r.DeleteSimilarIncidentByCaseAndSimilarCase(ctx, caseID, similarCaseID)
}
func (r *similarIncidentRepo) ListByCaseID(ctx context.Context, caseID uuid.UUID, opts ListOpts) ([]*SimilarIncident, string, error) {
	return r.ListSimilarIncidentsByCaseID(ctx, caseID, opts)
}

// NewRepos returns repository interfaces from a Postgres instance.
func NewRepos(pg *Postgres) (
	AlertRuleRepo, AlertRepo, ObservableRepo, CaseObservableRepo, ChildObservableRepo, SimilarIncidentRepo,
) {
	return &alertRuleRepo{pg}, &alertRepo{pg}, &observableRepo{pg}, &caseObservableRepo{pg}, &childObservableRepo{pg}, &similarIncidentRepo{pg}
}
