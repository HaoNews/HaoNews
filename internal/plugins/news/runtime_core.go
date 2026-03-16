package newsplugin

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"aip2p.org/internal/apphost"
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

func displayProjectName(project string) string {
	if strings.EqualFold(strings.TrimSpace(project), "aip2p.news") {
		return "AiP2P News Public"
	}
	return strings.TrimSpace(project)
}

func loadThemeAssets(theme apphost.WebTheme, funcs template.FuncMap) (*template.Template, fs.FS, error) {
	if theme != nil {
		tmpl, err := theme.ParseTemplates(funcs)
		if err != nil {
			return nil, nil, err
		}
		staticFS, err := theme.StaticFS()
		if err != nil {
			return nil, nil, err
		}
		return tmpl, staticFS, nil
	}
	tmpl, err := template.New("").Funcs(funcs).ParseFS(webFS, "web/templates/*.html")
	if err != nil {
		return nil, nil, err
	}
	staticFS, err := fs.Sub(webFS, "web/static")
	if err != nil {
		return nil, nil, err
	}
	return tmpl, staticFS, nil
}

func ensureRuntimeLayout(storeRoot, archiveRoot, rulesPath, writerPath, netPath string) error {
	storeRoot = strings.TrimSpace(storeRoot)
	if storeRoot != "" {
		for _, dir := range []string{
			filepath.Join(storeRoot, "data"),
			filepath.Join(storeRoot, "torrents"),
		} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
		}
	}
	archiveRoot = strings.TrimSpace(archiveRoot)
	if archiveRoot != "" {
		if err := os.MkdirAll(archiveRoot, 0o755); err != nil {
			return err
		}
	}
	runtimeRoot := strings.TrimSpace(filepath.Dir(archiveRoot))
	if runtimeRoot != "" {
		for _, dir := range []string{
			filepath.Join(runtimeRoot, "bin"),
			filepath.Join(runtimeRoot, "identities"),
			filepath.Join(runtimeRoot, "delegations"),
			filepath.Join(runtimeRoot, "revocations"),
		} {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
		}
	}
	rulesPath = strings.TrimSpace(rulesPath)
	if rulesPath != "" {
		if err := os.MkdirAll(filepath.Dir(rulesPath), 0o755); err != nil {
			return err
		}
		if err := ensureFileIfMissing(rulesPath, []byte(defaultSubscriptionsJSON)); err != nil {
			return err
		}
	}
	writerPath = strings.TrimSpace(writerPath)
	if writerPath != "" {
		if err := os.MkdirAll(filepath.Dir(writerPath), 0o755); err != nil {
			return err
		}
		if err := ensureWriterPolicyFile(writerPath); err != nil {
			return err
		}
		if err := ensureFileIfMissing(filepath.Join(filepath.Dir(writerPath), writerWhitelistINFName), []byte(defaultWriterWhitelistINF)); err != nil {
			return err
		}
		if err := ensureFileIfMissing(filepath.Join(filepath.Dir(writerPath), writerBlacklistINFName), []byte(defaultWriterBlacklistINF)); err != nil {
			return err
		}
	}
	netPath = strings.TrimSpace(netPath)
	if netPath != "" {
		if err := os.MkdirAll(filepath.Dir(netPath), 0o755); err != nil {
			return err
		}
		if _, err := os.Stat(netPath); errors.Is(err, os.ErrNotExist) {
			content, err := buildDefaultLatestNetINF()
			if err != nil {
				return err
			}
			if err := ensureFileIfMissing(netPath, []byte(content)); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if err := ensureFileIfMissing(filepath.Join(filepath.Dir(netPath), "Trackerlist.inf"), []byte(defaultTrackerListINF)); err != nil {
			return err
		}
		if err := appendNetworkIDIfMissing(netPath, latestOrgNetworkID); err != nil {
			return err
		}
		if err := appendLANPeerIfMissing(netPath, defaultLANPeer); err != nil {
			return err
		}
		if err := appendLANTorrentPeerIfMissing(netPath, defaultLANPeer); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(payload)
}
