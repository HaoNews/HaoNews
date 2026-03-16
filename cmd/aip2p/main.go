package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"aip2p.org/internal/aip2p"
	"aip2p.org/internal/apphost"
	"aip2p.org/internal/builtin"
	"aip2p.org/internal/host"
	"aip2p.org/internal/scaffold"
	"aip2p.org/internal/themes/directorytheme"
	"aip2p.org/internal/workspace"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}
	switch args[0] {
	case "publish":
		return runPublish(args[1:])
	case "verify":
		return runVerify(args[1:])
	case "show":
		return runShow(args[1:])
	case "sync":
		return runSync(args[1:])
	case "serve":
		return runServe(args[1:])
	case "plugins":
		return runPlugins(args[1:])
	case "themes":
		return runThemes(args[1:])
	case "apps":
		return runApps(args[1:])
	case "create":
		return runCreate(args[1:])
	default:
		return usageError()
	}
}

func runPublish(args []string) error {
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	storeRoot := fs.String("store", ".aip2p", "store root")
	author := fs.String("author", "", "agent author id")
	kind := fs.String("kind", "post", "message kind")
	channel := fs.String("channel", "", "message channel")
	title := fs.String("title", "", "message title")
	body := fs.String("body", "", "message body")
	replyInfoHash := fs.String("reply-infohash", "", "reply target infohash")
	replyMagnet := fs.String("reply-magnet", "", "reply target magnet")
	tagsCSV := fs.String("tags", "", "comma-separated tags")
	extensionsJSON := fs.String("extensions-json", "", "inline JSON object for message extensions")
	extensionsFile := fs.String("extensions-file", "", "path to JSON object file for message extensions")
	if err := fs.Parse(args); err != nil {
		return err
	}

	store, err := aip2p.OpenStore(*storeRoot)
	if err != nil {
		return err
	}

	var replyTo *aip2p.MessageLink
	if strings.TrimSpace(*replyInfoHash) != "" || strings.TrimSpace(*replyMagnet) != "" {
		replyTo = &aip2p.MessageLink{
			InfoHash: strings.TrimSpace(*replyInfoHash),
			Magnet:   strings.TrimSpace(*replyMagnet),
		}
	}
	extensions, err := loadJSONObject(*extensionsJSON, *extensionsFile)
	if err != nil {
		return err
	}

	result, err := aip2p.PublishMessage(store, aip2p.MessageInput{
		Kind:       *kind,
		Author:     *author,
		Channel:    *channel,
		Title:      *title,
		Body:       *body,
		ReplyTo:    replyTo,
		Tags:       splitCSV(*tagsCSV),
		Extensions: extensions,
		CreatedAt:  time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	return writeJSON(result)
}

func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dir := fs.String("dir", "", "content directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*dir) == "" {
		return errors.New("dir is required")
	}
	msg, body, err := aip2p.LoadMessage(*dir)
	if err != nil {
		return err
	}
	return writeJSON(struct {
		Valid   bool          `json:"valid"`
		Message aip2p.Message `json:"message"`
		BodyLen int           `json:"body_len"`
	}{
		Valid:   true,
		Message: msg,
		BodyLen: len(body),
	})
}

func runShow(args []string) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dir := fs.String("dir", "", "content directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*dir) == "" {
		return errors.New("dir is required")
	}
	msg, body, err := aip2p.LoadMessage(*dir)
	if err != nil {
		return err
	}
	return writeJSON(struct {
		Message aip2p.Message `json:"message"`
		Body    string        `json:"body"`
	}{
		Message: msg,
		Body:    body,
	})
}

func runSync(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	storeRoot := fs.String("store", ".aip2p", "store root")
	queuePath := fs.String("queue", "", "line-based magnet/infohash queue file")
	netPath := fs.String("net", "./aip2p_net.inf", "network bootstrap config")
	trackersPath := fs.String("trackers", "", "tracker list file; defaults to Trackerlist.inf next to the net config")
	subscriptionsPath := fs.String("subscriptions", "", "subscription rules file for pubsub topic joins")
	listenAddr := fs.String("listen", "0.0.0.0:0", "bittorrent listen address")
	magnets := fs.String("magnet", "", "comma-separated magnets or infohashes to sync immediately")
	poll := fs.Duration("poll", 30*time.Second, "queue polling interval")
	timeout := fs.Duration("timeout", 20*time.Second, "per-ref sync timeout")
	once := fs.Bool("once", false, "run one sync pass and exit")
	seed := fs.Bool("seed", true, "seed after download while daemon is running")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return aip2p.RunSync(ctx, aip2p.SyncOptions{
		StoreRoot:         *storeRoot,
		QueuePath:         *queuePath,
		NetPath:           *netPath,
		TrackerListPath:   *trackersPath,
		SubscriptionsPath: *subscriptionsPath,
		ListenAddr:        *listenAddr,
		Refs:              splitCSV(*magnets),
		PollInterval:      *poll,
		Timeout:           *timeout,
		Once:              *once,
		Seed:              *seed,
	}, log.Printf)
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	listenAddr := fs.String("listen", "0.0.0.0:51818", "http listen address")
	appID := fs.String("app", "", "built-in application id; defaults to the built-in sample app")
	appDir := fs.String("app-dir", "", "application directory containing aip2p.app.json and optional themes/plugins folders")
	pluginID := fs.String("plugin", "", "single built-in plugin id; ignored when --plugins is set")
	pluginsCSV := fs.String("plugins", "", "comma-separated built-in plugin ids to compose; overrides --plugin")
	pluginDirsCSV := fs.String("plugin-dir", "", "comma-separated external plugin directories containing aip2p.plugin.json")
	themeID := fs.String("theme", "", "theme id; defaults to the plugin default theme")
	themeDir := fs.String("theme-dir", "", "directory theme override; expects aip2p.theme.json plus templates/static")
	project := fs.String("project", "", "project id override")
	version := fs.String("version", "dev", "host version label")
	runtimeRoot := fs.String("runtime-root", "", "application runtime root")
	storeRoot := fs.String("store", "", "store root override")
	archiveRoot := fs.String("archive", "", "archive root override")
	rulesPath := fs.String("subscriptions", "", "subscription rules path override")
	writerPolicy := fs.String("writer-policy", "", "writer policy path override")
	netPath := fs.String("net", "", "network bootstrap config override")
	trackersPath := fs.String("trackers", "", "tracker list override")
	syncMode := fs.String("sync-mode", "", "sync mode override")
	syncBinary := fs.String("sync-binary", "", "managed sync binary override")
	syncStaleAfter := fs.Duration("sync-stale-after", 2*time.Minute, "managed sync stale restart threshold")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	instance, err := host.New(ctx, host.Config{
		App:              *appID,
		AppDir:           *appDir,
		Plugin:           *pluginID,
		Plugins:          splitCSV(*pluginsCSV),
		PluginDirs:       splitCSV(*pluginDirsCSV),
		Theme:            *themeID,
		ThemeDir:         *themeDir,
		Project:          *project,
		Version:          *version,
		ListenAddr:       *listenAddr,
		RuntimeRoot:      *runtimeRoot,
		StoreRoot:        *storeRoot,
		ArchiveRoot:      *archiveRoot,
		RulesPath:        *rulesPath,
		WriterPolicyPath: *writerPolicy,
		NetPath:          *netPath,
		TrackerPath:      *trackersPath,
		SyncMode:         *syncMode,
		SyncBinaryPath:   *syncBinary,
		SyncStaleAfter:   *syncStaleAfter,
		Logf:             log.Printf,
	})
	if err != nil {
		return err
	}
	log.Printf("AiP2P host serving plugin=%s theme=%s on http://%s", instance.Site().Manifest.ID, instance.Site().Theme.ID, *listenAddr)
	return instance.ListenAndServe(ctx)
}

func runPlugins(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p plugins <list|inspect>")
	}
	switch args[0] {
	case "list":
		registry := builtin.DefaultRegistry()
		return writeJSON(struct {
			Plugins []any `json:"plugins"`
		}{
			Plugins: manifestsToAny(registry.PluginManifests()),
		})
	case "inspect":
		fs := flag.NewFlagSet("plugins inspect", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		dir := fs.String("dir", "", "plugin directory containing aip2p.plugin.json")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		registry := builtin.DefaultRegistry()
		bundle, err := workspace.LoadPluginBundleDir(*dir)
		if err != nil {
			return err
		}
		_, manifest, err := workspace.LoadPluginDir(*dir, registry)
		if err != nil {
			return err
		}
		resolved, err := workspace.ValidatePluginManifest(manifest, registry)
		if err != nil {
			return err
		}
		resolved.Root = bundle.Root
		resolved.Config = bundle.Config
		return writeJSON(struct {
			Dir      string                   `json:"dir"`
			Manifest apphost.PluginManifest   `json:"manifest"`
			Config   map[string]any           `json:"config,omitempty"`
			Resolved workspace.ResolvedPlugin `json:"resolved"`
		}{
			Dir:      *dir,
			Manifest: manifest,
			Config:   bundle.Config,
			Resolved: resolved,
		})
	default:
		return errors.New("usage: aip2p plugins <list|inspect>")
	}
}

func runApps(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p apps <list|inspect|validate>")
	}
	switch args[0] {
	case "list":
		return writeJSON(struct {
			Apps []apphost.AppManifest `json:"apps"`
		}{
			Apps: builtin.DefaultApps(),
		})
	case "inspect":
		fs := flag.NewFlagSet("apps inspect", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		dir := fs.String("dir", "", "application directory containing aip2p.app.json")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		bundle, report, err := inspectAppDir(*dir)
		if err != nil {
			return err
		}
		return writeJSON(struct {
			Dir        string                     `json:"dir"`
			App        apphost.AppManifest        `json:"app"`
			Config     workspace.AppConfig        `json:"config"`
			Plugins    []apphost.PluginManifest   `json:"plugins"`
			Themes     []apphost.ThemeManifest    `json:"themes"`
			Validation workspace.ValidationReport `json:"validation"`
		}{
			Dir:        *dir,
			App:        bundle.App,
			Config:     bundle.Config,
			Plugins:    bundle.PluginManifests,
			Themes:     bundle.ThemeManifests,
			Validation: report,
		})
	case "validate":
		fs := flag.NewFlagSet("apps validate", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		dir := fs.String("dir", "", "application directory containing aip2p.app.json")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		_, report, err := inspectAppDir(*dir)
		if err != nil {
			return err
		}
		return writeJSON(report)
	default:
		return errors.New("usage: aip2p apps <list|inspect|validate>")
	}
}

func runCreate(args []string) error {
	if len(args) < 2 {
		return errors.New("usage: aip2p create <plugin|theme|app> <name> [--out dir]")
	}
	kind := strings.TrimSpace(args[0])
	name := strings.TrimSpace(args[1])
	if name == "" {
		return errors.New("name is required")
	}
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	outDir := fs.String("out", scaffold.Slug(name), "output directory")
	if err := fs.Parse(args[2:]); err != nil {
		return err
	}

	var (
		files []scaffold.File
		err   error
	)
	switch kind {
	case "plugin":
		files, err = scaffold.PluginFiles(name)
	case "theme":
		files, err = scaffold.ThemeFiles(name)
	case "app":
		files, err = scaffold.AppFiles(name)
	default:
		return errors.New("usage: aip2p create <plugin|theme|app> <name> [--out dir]")
	}
	if err != nil {
		return err
	}
	if err := scaffold.WriteFiles(*outDir, files); err != nil {
		return err
	}
	return writeJSON(map[string]any{
		"kind":   kind,
		"name":   name,
		"output": *outDir,
		"files":  filePaths(files),
	})
}

func filePaths(files []scaffold.File) []string {
	out := make([]string, 0, len(files))
	for _, file := range files {
		out = append(out, file.Path)
	}
	return out
}

func runThemes(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p themes <list|inspect>")
	}
	switch args[0] {
	case "list":
		registry := builtin.DefaultRegistry()
		return writeJSON(struct {
			Themes []any `json:"themes"`
		}{
			Themes: manifestsToAny(registry.ThemeManifests()),
		})
	case "inspect":
		fs := flag.NewFlagSet("themes inspect", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		dir := fs.String("dir", "", "theme directory containing aip2p.theme.json")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		theme, err := directorytheme.Load(*dir)
		if err != nil {
			return err
		}
		return writeJSON(struct {
			Dir      string                `json:"dir"`
			Manifest apphost.ThemeManifest `json:"manifest"`
		}{
			Dir:      *dir,
			Manifest: theme.Manifest(),
		})
	default:
		return errors.New("usage: aip2p themes <list|inspect>")
	}
}

func manifestsToAny[T any](items []T) []any {
	out := make([]any, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func inspectAppDir(dir string) (workspace.AppBundle, workspace.ValidationReport, error) {
	bundle, err := workspace.LoadAppBundle(dir)
	if err != nil {
		return workspace.AppBundle{}, workspace.ValidationReport{}, err
	}
	registry := builtin.DefaultRegistry()
	plugins, _, err := workspace.LoadPlugins(filepath.Join(bundle.Root, "plugins"), registry)
	if err != nil {
		return workspace.AppBundle{}, workspace.ValidationReport{}, err
	}
	for _, plugin := range plugins {
		if err := registry.RegisterPlugin(plugin); err != nil {
			return workspace.AppBundle{}, workspace.ValidationReport{}, err
		}
	}
	for _, theme := range bundle.Themes {
		if err := registry.RegisterTheme(theme); err != nil {
			return workspace.AppBundle{}, workspace.ValidationReport{}, err
		}
	}
	report, err := workspace.ValidateAppBundle(bundle, registry, registry)
	if err != nil {
		return workspace.AppBundle{}, workspace.ValidationReport{}, err
	}
	return bundle, report, nil
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
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

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func usageError() error {
	return errors.New("usage: aip2p <publish|verify|show|sync|serve|plugins|themes|apps|create> [flags]")
}

func loadJSONObject(inline, path string) (map[string]any, error) {
	inline = strings.TrimSpace(inline)
	path = strings.TrimSpace(path)
	if inline != "" && path != "" {
		return nil, errors.New("use only one of extensions-json or extensions-file")
	}
	if inline == "" && path == "" {
		return map[string]any{}, nil
	}
	var data []byte
	var err error
	if inline != "" {
		data = []byte(inline)
	} else {
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("parse extensions json: %w", err)
	}
	if value == nil {
		value = map[string]any{}
	}
	return value, nil
}
