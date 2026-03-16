package newsplugin

import "strconv"

import "path/filepath"

func (a *App) WriterPolicyPath() string {
	return a.writerPath
}

func (a *App) GovernanceSummary() []SummaryStat {
	return a.governanceSummary()
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
		{Label: "Trusted authorities", Value: strconv.Itoa(len(policy.TrustedAuthorities))},
		{Label: "Shared registries", Value: strconv.Itoa(len(policy.SharedRegistries))},
		{Label: "Writer caps", Value: strconv.Itoa(len(policy.AgentCapabilities) + len(policy.PublicKeyCapabilities))},
		{Label: "Relay trust rules", Value: strconv.Itoa(len(policy.RelayPeerTrust) + len(policy.RelayHostTrust))},
	}
}

func DefaultWriterPolicy() WriterPolicy {
	return defaultWriterPolicy()
}

func WriterWhitelistPath(writerPolicyPath string) string {
	return filepath.Join(filepath.Dir(writerPolicyPath), writerWhitelistINFName)
}

func WriterBlacklistPath(writerPolicyPath string) string {
	return filepath.Join(filepath.Dir(writerPolicyPath), writerBlacklistINFName)
}
