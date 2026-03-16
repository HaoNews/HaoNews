package newsplugin

import "path/filepath"

func (a *App) WriterPolicyPath() string {
	return a.writerPath
}

func (a *App) GovernanceSummary() []SummaryStat {
	return a.governanceSummary()
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
