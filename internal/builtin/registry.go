package builtin

import (
	_ "embed"
	"fmt"
	"strings"

	"aip2p.org/internal/apphost"
	newsarchive "aip2p.org/internal/plugins/newsarchive"
	newscontent "aip2p.org/internal/plugins/newscontent"
	newsgovernance "aip2p.org/internal/plugins/newsgovernance"
	newsops "aip2p.org/internal/plugins/newsops"
	"aip2p.org/internal/themes/defaultnews"
)

//go:embed news-demo.app.json
var newsDemoAppJSON []byte

func DefaultRegistry() *apphost.Registry {
	registry := apphost.NewRegistry()
	registry.MustRegisterTheme(defaultnews.Theme{})
	registry.MustRegisterPlugin(newscontent.Plugin{})
	registry.MustRegisterPlugin(newsarchive.Plugin{})
	registry.MustRegisterPlugin(newsgovernance.Plugin{})
	registry.MustRegisterPlugin(newsops.Plugin{})
	return registry
}

func DefaultApps() []apphost.AppManifest {
	return []apphost.AppManifest{
		apphost.MustLoadAppManifestJSON(newsDemoAppJSON),
	}
}

func ResolveApp(id string) (apphost.AppManifest, error) {
	id = strings.ToLower(strings.TrimSpace(id))
	for _, app := range DefaultApps() {
		if strings.ToLower(strings.TrimSpace(app.ID)) == id {
			return app, nil
		}
	}
	return apphost.AppManifest{}, fmt.Errorf("app %q not found", id)
}
