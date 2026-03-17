// Package domain defines case and worknote domain types and validation.
package domain

import "time"

// Case states: 8 SIR workflow stages per ServiceNow SIR lifecycle.
const (
	StateDraft     = "Draft"
	StateAnalysis  = "Analysis"
	StateContain   = "Contain"
	StateEradicate = "Eradicate"
	StateRecover   = "Recover"
	StateReview    = "Review"
	StateResolved  = "Resolved"
	StateClosed    = "Closed"
)

// Terminal states; no transition back in v1.
var TerminalStates = map[string]bool{StateResolved: true, StateClosed: true}

// Active states count toward active_duration_seconds (all non-terminal).
var ActiveStates = map[string]bool{
	StateDraft: true, StateAnalysis: true, StateContain: true, StateEradicate: true,
	StateRecover: true, StateReview: true,
}

// Allowed states (the 8 SIR states).
var AllowedStates = map[string]bool{
	StateDraft: true, StateAnalysis: true, StateContain: true, StateEradicate: true,
	StateRecover: true, StateReview: true, StateResolved: true, StateClosed: true,
}

// Priorities P1–P4 per platform contract.
var AllowedPriorities = map[string]bool{"P1": true, "P2": true, "P3": true, "P4": true}

// Severities per platform contract.
var AllowedSeverities = map[string]bool{"critical": true, "high": true, "medium": true, "low": true}

const (
	ShortDescriptionMinLen = 1
	ShortDescriptionMaxLen = 500
	DefaultNoteType        = "worknote"
)

// Case is the domain case entity.
type Case struct {
	ID                     string
	CaseNumber             string
	ShortDescription       string
	Description            string
	State                  string
	Priority               string
	Severity               string
	OpenedByUserID         string
	OpenedTime             time.Time
	EventOccurredTime      *time.Time
	EventReceivedTime      *time.Time
	AffectedUserID         *string
	AssignedUserID         *string
	AssignmentGroupID      *string
	AlertRuleID            *string
	ActiveDurationSeconds  int64
	Accuracy               *string
	Determination          *string
	Impact                 *string
	ClosureCode            *string
	ClosureReason          *string
	ClosedByUserID         *string
	ClosedTime             *time.Time
	IsActive               bool
	VersionNo              int32
	CreatedAt              time.Time
	UpdatedAt              time.Time
	// Incident form fields
	FollowupTime            *time.Time
	Category                *string
	Subcategory             *string
	Source                  *string
	SourceTool              *string
	SourceToolFeature       *string
	ConfigurationItem       *string
	SOCNotes                *string
	NextSteps               *string
	CSIRTClassification     *string
	SOCLeadUserID           *string
	// Incident detail fields
	RequestedByUserID       *string
	EnvironmentLevel        *string
	EnvironmentType         *string
	PDN                     *string
	ImpactedObject          *string
	NotificationTime        *time.Time
	MTTR                    *string
	ReassignmentCount       int32
	AssignedToCount         int32
	IsAffectedUserVIP       bool
	EngineeringDocument     *string
	ResponseDocument        *string
}

// Worknote is the domain worknote entity.
type Worknote struct {
	ID                string
	CaseID            string
	NoteText          string
	NoteType          string
	CreatedByUserID   string
	CreatedAt         time.Time
	UpdatedAt         *time.Time
	IsDeleted         bool
}
