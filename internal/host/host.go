package host

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"aip2p.org/internal/apphost"
	"aip2p.org/internal/builtin"
	"aip2p.org/internal/themes/directorytheme"
	"aip2p.org/internal/workspace"
)

type Config struct {
	App              string
	Plugin           string
	Plugins          []string
	Theme            string
	ThemeDir         string
	AppDir           string
	Project          string
	Version          string
	ListenAddr       string
	RuntimeRoot      string
	StoreRoot        string
	ArchiveRoot      string
	RulesPath        string
	WriterPolicyPath string
	NetPath          string
	TrackerPath      string
	SyncMode         string
	SyncBinaryPath   string
	SyncStaleAfter   time.Duration
	Logf             func(string, ...any)
}

type Instance struct {
	config Config
	site   *apphost.Site
	server *http.Server
}

func New(ctx context.Context, cfg Config) (*Instance, error) {
	cfg = normalizeConfig(cfg)
	appDirExplicit := strings.TrimSpace(cfg.AppDir) != ""
	themeExplicit := strings.TrimSpace(cfg.Theme) != ""
	var bundle workspace.AppBundle
	if appDirExplicit {
		var err error
		bundle, err = workspace.LoadAppBundle(cfg.AppDir)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(cfg.App) == "" {
			cfg.App = bundle.App.ID
		}
		if len(cfg.Plugins) == 0 && strings.TrimSpace(cfg.Plugin) == "" {
			cfg.Plugins = append([]string(nil), bundle.App.Plugins...)
		}
		if !themeExplicit && strings.TrimSpace(cfg.ThemeDir) == "" {
			cfg.Theme = bundle.App.Theme
		}
	}
	if strings.TrimSpace(cfg.App) != "" {
		app, err := builtin.ResolveApp(cfg.App)
		if err != nil {
			if !appDirExplicit || !strings.EqualFold(strings.TrimSpace(cfg.App), strings.TrimSpace(bundle.App.ID)) {
				return nil, err
			}
		} else {
			if len(cfg.Plugins) == 0 && strings.TrimSpace(cfg.Plugin) == "" {
				cfg.Plugins = append([]string(nil), app.Plugins...)
			}
			if !themeExplicit && strings.TrimSpace(cfg.ThemeDir) == "" && strings.TrimSpace(cfg.Theme) == "" {
				cfg.Theme = app.Theme
			}
		}
	}
	registry := builtin.DefaultRegistry()
	for _, theme := range bundle.Themes {
		if err := registry.RegisterTheme(theme); err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(cfg.ThemeDir) != "" {
		theme, err := directorytheme.Load(cfg.ThemeDir)
		if err != nil {
			return nil, err
		}
		if err := registry.RegisterTheme(theme); err != nil {
			return nil, err
		}
		if !themeExplicit {
			cfg.Theme = theme.Manifest().ID
		}
	}
	site, err := registry.Build(ctx, apphost.Config{
		Plugin:           cfg.Plugin,
		Plugins:          cfg.Plugins,
		Theme:            cfg.Theme,
		Project:          cfg.Project,
		Version:          cfg.Version,
		ListenAddr:       cfg.ListenAddr,
		RuntimeRoot:      cfg.RuntimeRoot,
		StoreRoot:        cfg.StoreRoot,
		ArchiveRoot:      cfg.ArchiveRoot,
		RulesPath:        cfg.RulesPath,
		WriterPolicyPath: cfg.WriterPolicyPath,
		NetPath:          cfg.NetPath,
		TrackerPath:      cfg.TrackerPath,
		SyncMode:         cfg.SyncMode,
		SyncBinaryPath:   cfg.SyncBinaryPath,
		SyncStaleAfter:   cfg.SyncStaleAfter,
		Logf:             cfg.Logf,
	})
	if err != nil {
		return nil, err
	}
	return &Instance{
		config: cfg,
		site:   site,
		server: &http.Server{
			Addr:    cfg.ListenAddr,
			Handler: site.Handler,
		},
	}, nil
}

func (i *Instance) ListenAndServe(ctx context.Context) error {
	if i == nil || i.server == nil {
		return errors.New("host instance is not initialized")
	}
	errCh := make(chan error, 1)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = i.server.Shutdown(shutdownCtx)
		_ = i.site.Shutdown(shutdownCtx)
	}()
	go func() {
		errCh <- i.server.ListenAndServe()
	}()
	err := <-errCh
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (i *Instance) Site() *apphost.Site {
	if i == nil {
		return nil
	}
	return i.site
}

func normalizeConfig(cfg Config) Config {
	if strings.TrimSpace(cfg.AppDir) == "" && strings.TrimSpace(cfg.App) == "" && len(cfg.Plugins) == 0 && strings.TrimSpace(cfg.Plugin) == "" {
		cfg.App = "default-news"
	}
	if strings.TrimSpace(cfg.ListenAddr) == "" {
		cfg.ListenAddr = "0.0.0.0:51818"
	}
	if strings.TrimSpace(cfg.Version) == "" {
		cfg.Version = "dev"
	}
	if cfg.SyncStaleAfter <= 0 {
		cfg.SyncStaleAfter = 2 * time.Minute
	}
	if cfg.Logf == nil {
		cfg.Logf = log.Printf
	}
	return cfg
}

func (i *Instance) String() string {
	if i == nil || i.site == nil {
		return "aip2p host"
	}
	return fmt.Sprintf("%s on %s", i.site.Manifest.Name, i.config.ListenAddr)
}
