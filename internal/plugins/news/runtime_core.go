package newsplugin

import (
	"html/template"
	"net/http"
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

func (a *App) ProjectID() string {
	return a.project
}

func (a *App) VersionString() string {
	return a.version
}

func (a *App) StoreRoot() string {
	return a.storeRoot
}

func (a *App) Index() (Index, error) {
	return a.index()
}

func (a *App) SubscriptionRules() (SubscriptionRules, error) {
	return a.subscriptionRules()
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	writeJSON(w, status, payload)
}
