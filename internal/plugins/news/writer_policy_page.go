package newsplugin

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type WriterPolicyPageData struct {
	Project                   string
	Version                   string
	PageNav                   []NavItem
	NodeStatus                NodeStatus
	PolicyPath                string
	WhitelistPath             string
	BlacklistPath             string
	Saved                     bool
	Error                     string
	SyncMode                  string
	AllowUnsigned             bool
	DefaultCapability         string
	RelayDefaultTrust         string
	TrustedAuthoritiesText    string
	SharedRegistriesText      string
	AgentCapabilitiesText     string
	PublicKeyCapabilitiesText string
	AllowedAgentIDsText       string
	AllowedPublicKeysText     string
	BlockedAgentIDsText       string
	BlockedPublicKeysText     string
	RelayPeerTrustText        string
	RelayHostTrustText        string
	EffectiveSummary          []SummaryStat
}

func (a *App) handleWriterPolicy(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/writer-policy" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.renderWriterPolicyPage(w, r, "", r.URL.Query().Get("saved") == "1")
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			a.renderWriterPolicyPage(w, r, err.Error(), false)
			return
		}
		policy, err := writerPolicyFromForm(r)
		if err != nil {
			a.renderWriterPolicyPage(w, r, err.Error(), false)
			return
		}
		policy.normalize()
		data, err := json.MarshalIndent(policy, "", "  ")
		if err != nil {
			a.renderWriterPolicyPage(w, r, err.Error(), false)
			return
		}
		data = append(data, '\n')
		if err := os.WriteFile(a.writerPath, data, 0o644); err != nil {
			a.renderWriterPolicyPage(w, r, err.Error(), false)
			return
		}
		http.Redirect(w, r, "/writer-policy?saved=1", http.StatusSeeOther)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *App) renderWriterPolicyPage(w http.ResponseWriter, r *http.Request, formErr string, saved bool) {
	policy, err := loadLocalWriterPolicy(a.writerPath)
	if err != nil {
		formErr = err.Error()
		policy = defaultWriterPolicy()
	}
	if r.Method == http.MethodPost {
		if posted, postErr := writerPolicyFromForm(r); postErr == nil {
			policy = posted
		}
	}
	effective, effectiveErr := a.loadWriter(a.writerPath)
	summary := writerPolicySummary(effective, effectiveErr)
	data := WriterPolicyPageData{
		Project:                   displayProjectName(a.project),
		Version:                   a.version,
		PageNav:                   a.pageNav("/writer-policy"),
		NodeStatus:                NodeStatus{Summary: "policy", SummaryTone: "good"},
		PolicyPath:                a.writerPath,
		WhitelistPath:             filepath.Join(filepath.Dir(a.writerPath), writerWhitelistINFName),
		BlacklistPath:             filepath.Join(filepath.Dir(a.writerPath), writerBlacklistINFName),
		Saved:                     saved,
		Error:                     formErr,
		SyncMode:                  string(policy.SyncMode),
		AllowUnsigned:             policy.AllowUnsigned,
		DefaultCapability:         string(policy.DefaultCapability),
		RelayDefaultTrust:         string(policy.RelayDefaultTrust),
		TrustedAuthoritiesText:    formatStringMap(policy.TrustedAuthorities),
		SharedRegistriesText:      strings.Join(policy.SharedRegistries, "\n"),
		AgentCapabilitiesText:     formatCapabilityMap(policy.AgentCapabilities),
		PublicKeyCapabilitiesText: formatCapabilityMap(policy.PublicKeyCapabilities),
		AllowedAgentIDsText:       strings.Join(policy.AllowedAgentIDs, "\n"),
		AllowedPublicKeysText:     strings.Join(policy.AllowedPublicKeys, "\n"),
		BlockedAgentIDsText:       strings.Join(policy.BlockedAgentIDs, "\n"),
		BlockedPublicKeysText:     strings.Join(policy.BlockedPublicKeys, "\n"),
		RelayPeerTrustText:        formatRelayTrustMap(policy.RelayPeerTrust),
		RelayHostTrustText:        formatRelayTrustMap(policy.RelayHostTrust),
		EffectiveSummary:          summary,
	}
	if err := a.templates.ExecuteTemplate(w, "writer_policy.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadLocalWriterPolicy(path string) (WriterPolicy, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return defaultWriterPolicy(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultWriterPolicy(), nil
		}
		return WriterPolicy{}, err
	}
	var policy WriterPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return WriterPolicy{}, err
	}
	policy.normalize()
	return policy, nil
}

func writerPolicyFromForm(r *http.Request) (WriterPolicy, error) {
	policy := WriterPolicy{
		SyncMode:              WriterSyncMode(r.FormValue("sync_mode")),
		AllowUnsigned:         r.FormValue("allow_unsigned") == "on",
		DefaultCapability:     WriterCapability(r.FormValue("default_capability")),
		RelayDefaultTrust:     RelayTrust(r.FormValue("relay_default_trust")),
		TrustedAuthorities:    parseStringMap(r.FormValue("trusted_authorities")),
		SharedRegistries:      parseList(r.FormValue("shared_registries")),
		AgentCapabilities:     parseCapabilityMap(r.FormValue("agent_capabilities")),
		PublicKeyCapabilities: parseCapabilityMap(r.FormValue("public_key_capabilities")),
		AllowedAgentIDs:       parseList(r.FormValue("allowed_agent_ids")),
		AllowedPublicKeys:     parseList(r.FormValue("allowed_public_keys")),
		BlockedAgentIDs:       parseList(r.FormValue("blocked_agent_ids")),
		BlockedPublicKeys:     parseList(r.FormValue("blocked_public_keys")),
		RelayPeerTrust:        parseRelayTrustMap(r.FormValue("relay_peer_trust")),
		RelayHostTrust:        parseRelayTrustMap(r.FormValue("relay_host_trust")),
	}
	policy.normalize()
	return policy, nil
}

func parseList(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ','
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func parseStringMap(raw string) map[string]string {
	lines := parseList(raw)
	if len(lines) == 0 {
		return nil
	}
	out := make(map[string]string, len(lines))
	for _, line := range lines {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		out[key] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseCapabilityMap(raw string) map[string]WriterCapability {
	lines := parseList(raw)
	if len(lines) == 0 {
		return nil
	}
	out := make(map[string]WriterCapability, len(lines))
	for _, line := range lines {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		out[key] = WriterCapability(value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseRelayTrustMap(raw string) map[string]RelayTrust {
	lines := parseList(raw)
	if len(lines) == 0 {
		return nil
	}
	out := make(map[string]RelayTrust, len(lines))
	for _, line := range lines {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		out[key] = RelayTrust(value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func formatStringMap(items map[string]string) string {
	if len(items) == 0 {
		return ""
	}
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, key+"="+items[key])
	}
	return strings.Join(lines, "\n")
}

func formatCapabilityMap(items map[string]WriterCapability) string {
	if len(items) == 0 {
		return ""
	}
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, key+"="+string(items[key]))
	}
	return strings.Join(lines, "\n")
}

func formatRelayTrustMap(items map[string]RelayTrust) string {
	if len(items) == 0 {
		return ""
	}
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, key+"="+string(items[key]))
	}
	return strings.Join(lines, "\n")
}

func writerPolicySummary(policy WriterPolicy, err error) []SummaryStat {
	if err != nil {
		return []SummaryStat{
			{Label: "Effective policy", Value: "load error"},
			{Label: "Reason", Value: err.Error()},
		}
	}
	return []SummaryStat{
		{Label: "Sync mode", Value: string(policy.SyncMode)},
		{Label: "Trusted authorities", Value: itoa(len(policy.TrustedAuthorities))},
		{Label: "Shared registries", Value: itoa(len(policy.SharedRegistries))},
		{Label: "Writer caps", Value: itoa(len(policy.AgentCapabilities) + len(policy.PublicKeyCapabilities))},
		{Label: "Relay trust rules", Value: itoa(len(policy.RelayPeerTrust) + len(policy.RelayHostTrust))},
	}
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
