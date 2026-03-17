package orchestrator

import (
	"github.com/servicenow/api-gateway/clients"
)

// Orchestrator holds downstream client interfaces and composes responses.
type Orchestrator struct {
	CaseCmd     clients.CaseCommandService
	CaseQuery   clients.CaseQueryService
	ObsCmd      clients.ObservableCommandService
	ObsQuery    clients.ObservableQueryService
	Enrichment  clients.EnrichmentQueryService
	ThreatLookup clients.ThreatLookupQueryService
	Reference   clients.ReferenceQueryService
	Attachment  clients.AttachmentQueryService
	AttachmentCmd clients.AttachmentCommandService
	Audit       clients.AuditQueryService
}

// New returns an Orchestrator with the given clients.
func New(
	caseCmd clients.CaseCommandService,
	caseQuery clients.CaseQueryService,
	obsCmd clients.ObservableCommandService,
	obsQuery clients.ObservableQueryService,
	enrichment clients.EnrichmentQueryService,
	threatLookup clients.ThreatLookupQueryService,
	ref clients.ReferenceQueryService,
	att clients.AttachmentQueryService,
	attCmd clients.AttachmentCommandService,
	audit clients.AuditQueryService,
) *Orchestrator {
	return &Orchestrator{
		CaseCmd:       caseCmd,
		CaseQuery:     caseQuery,
		ObsCmd:        obsCmd,
		ObsQuery:      obsQuery,
		Enrichment:    enrichment,
		ThreatLookup:  threatLookup,
		Reference:     ref,
		Attachment:    att,
		AttachmentCmd: attCmd,
		Audit:         audit,
	}
}
