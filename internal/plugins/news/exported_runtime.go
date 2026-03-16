package newsplugin

import (
	"html/template"
	"path/filepath"
)

func (a *App) Templates() *template.Template {
	return a.templates
}

func (a *App) PageNav(activePath string) []NavItem {
	return a.pageNav(activePath)
}

func (a *App) ProjectName() string {
	return displayProjectName(a.project)
}

func (a *App) VersionString() string {
	return a.version
}

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
