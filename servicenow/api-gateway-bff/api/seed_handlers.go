package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/servicenow/api-gateway/clients"
)

// Assignment group IDs from migration (SOC L1, SOC L2, CSIRT).
var presentationGroupIDs = []string{
	"b0000000-0000-4000-8000-000000000002", // SOC L1
	"b0000000-0000-4000-8000-000000000003", // SOC L2
	"b0000000-0000-4000-8000-000000000004", // CSIRT
}

// SeedPresentation creates 50 cases with 200 observables total (4 per case) for demo/presentation.
// POST /api/v1/seed/presentation
func (h *Handler) SeedPresentation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r)
		return
	}

	ctx := r.Context()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	caseTitles := presentationCaseTitles()
	observableSpecs := presentationObservableSpecs()

	var casesCreated, observablesCreated int
	for i := 0; i < 50; i++ {
		title := caseTitles[i%len(caseTitles)]
		groupID := presentationGroupIDs[i%len(presentationGroupIDs)]
		priority := []string{"P1", "P2", "P3", "P4"}[i%4]
		severity := []string{"critical", "high", "medium", "low"}[i%4]

		c, err := h.Orch.CaseCmd.CreateCase(ctx, &clients.CreateCaseRequest{
			Title:             title,
			Description:       fmt.Sprintf("Presentation seed case for demo. Priority %s, severity %s.", priority, severity),
			Priority:          priority,
			Severity:          severity,
			AssignmentGroupID: groupID,
		})
		if err != nil {
			writeDownstreamError(w, r, err)
			return
		}
		casesCreated++

		// 4 observables per case → 200 total
		for j := 0; j < 4; j++ {
			spec := observableSpecs[(i*4+j)%len(observableSpecs)]
			value := spec.value(rng, i, j)
			err = h.Orch.ObsCmd.CreateAndLinkObservable(ctx, c.ID, spec.obsType, value)
			if err != nil {
				// log but continue so we still return counts
				continue
			}
			observablesCreated++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"message":             "Presentation seed completed.",
		"cases_created":       casesCreated,
		"observables_created": observablesCreated,
	})
}

type obsSpec struct {
	obsType string
	value   func(rng *rand.Rand, caseIdx, obsIdx int) string
}

func presentationObservableSpecs() []obsSpec {
	return []obsSpec{
		{"ipv4", func(r *rand.Rand, i, j int) string {
			return fmt.Sprintf("192.168.%d.%d", 1+r.Intn(254), 1+r.Intn(254))
		}},
		{"domain", func(r *rand.Rand, i, j int) string {
			domains := []string{"example.com", "suspicious.net", "phish-site.org", "malware-host.co", "c2-server.io"}
			return fmt.Sprintf("case%d-%s", i, domains[r.Intn(len(domains))])
		}},
		{"email", func(r *rand.Rand, i, j int) string {
			return fmt.Sprintf("user%d.alert%d@example.com", i, j)
		}},
		{"url", func(r *rand.Rand, i, j int) string {
			return fmt.Sprintf("https://example.com/path/case-%d/obs-%d", i, j)
		}},
		{"md5", func(r *rand.Rand, i, j int) string {
			const hex = "0123456789abcdef"
			b := make([]byte, 32)
			for k := range b {
				b[k] = hex[r.Intn(16)]
			}
			return string(b)
		}},
		{"sha256", func(r *rand.Rand, i, j int) string {
			const hex = "0123456789abcdef"
			b := make([]byte, 64)
			for k := range b {
				b[k] = hex[r.Intn(16)]
			}
			return string(b)
		}},
	}
}

func presentationCaseTitles() []string {
	return []string{
		"Phishing email reported by user",
		"Suspicious login from unknown IP",
		"Malware detected on endpoint",
		"Data exfiltration attempt blocked",
		"C2 callback to external host",
		"Privilege escalation alert",
		"Brute force login attempts",
		"Ransomware file hash detection",
		"Unusual outbound traffic volume",
		"Credential stuffing campaign",
		"Insider threat - bulk download",
		"Domain fronting abuse",
		"Malicious macro in document",
		"Supply chain compromise",
		"Zero-day exploit attempt",
		"Lateral movement detected",
		"DNS tunneling activity",
		"API key leaked in repo",
		"Impossible travel login",
		"Anomalous PowerShell execution",
		"Web shell upload",
		"SQL injection attempt",
		"Port scan from internal host",
		"Disabled security control",
		"Unpatched critical vulnerability",
		"Stolen cookie reuse",
		"Fake login page hosting",
		"BEC wire transfer request",
		"Cloud bucket misconfiguration",
		"Container escape attempt",
		"Kernel module load",
		"Scheduled task persistence",
		"Registry run key modification",
		"Network share enumeration",
		"Password spray campaign",
		"Token theft - LSASS",
		"Shadow copy deletion",
		"Backup deletion attempt",
		"Log tampering detected",
		"New admin account created",
		"Service account abuse",
		"VPN credential compromise",
		"Mobile device jailbreak",
		"USB device exfiltration",
		"Print nightmare exploitation",
		"ProxyLogon exploitation",
		"Log4j exploitation attempt",
		"Citrix vulnerability abuse",
		"Pulse Secure VPN exploit",
		"Fortinet SSL VPN exploit",
	}
}
