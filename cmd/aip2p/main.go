package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"aip2p.org/internal/aip2p"
	"aip2p.org/internal/apphost"
	"aip2p.org/internal/builtin"
	"aip2p.org/internal/extensions"
	"aip2p.org/internal/host"
	"aip2p.org/internal/scaffold"
	"aip2p.org/internal/themes/directorytheme"
	"aip2p.org/internal/workspace"
)

type boolFlag interface {
	IsBoolFlag() bool
}

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
	case "identity":
		return runIdentity(args[1:])
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
	case "peers":
		return runPeers(args[1:])
	case "subscribe":
		return runSubscribe(args[1:])
	case "store":
		return runStore(args[1:])
	case "config":
		return runConfig(args[1:])
	case "sync-status":
		return runSyncStatus(args[1:])
	case "bootstrap":
		return runBootstrap(args[1:])
	default:
		return usageError()
	}
}

func runPublish(args []string) error {
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	storeRoot := fs.String("store", ".aip2p", "store root")
	author := fs.String("author", "", "agent author id")
	identityFile := fs.String("identity-file", "", "path to a signing identity JSON file")
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
	if strings.TrimSpace(*identityFile) == "" {
		return errors.New("identity-file is required; all new posts and replies must be signed")
	}

	store, err := aip2p.OpenStore(*storeRoot)
	if err != nil {
		return err
	}

	// Try loading as encrypted identity first, fall back to plain
	identity, err := aip2p.LoadEncryptedIdentity(strings.TrimSpace(*identityFile), "")
	if err != nil {
		// If it needs a password, try prompting
		if strings.Contains(err.Error(), "password required") {
			var pw string
			fmt.Fprint(os.Stderr, "Enter password to unlock identity: ")
			fmt.Scanln(&pw)
			identity, err = aip2p.LoadEncryptedIdentity(strings.TrimSpace(*identityFile), pw)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if strings.TrimSpace(*author) == "" && strings.TrimSpace(identity.Author) != "" {
		*author = strings.TrimSpace(identity.Author)
	}
	if strings.TrimSpace(*author) == "" {
		return errors.New("author is required; set --author or store author in identity-file")
	}
	if strings.TrimSpace(identity.Author) != "" && strings.TrimSpace(*author) != strings.TrimSpace(identity.Author) {
		return errors.New("author does not match identity-file author")
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
		Identity:   &identity,
		Extensions: extensions,
		CreatedAt:  time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	return writeJSON(result)
}

func runIdentity(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p identity <init|create-hd|derive|list|export|recover|registry> [flags]")
	}
	switch args[0] {
	case "init":
		return runIdentityInit(args[1:])
	case "create-hd":
		return runIdentityCreateHD(args[1:])
	case "derive":
		return runIdentityDerive(args[1:])
	case "list":
		return runIdentityList(args[1:])
	case "export":
		return runIdentityExport(args[1:])
	case "recover":
		return runIdentityRecover(args[1:])
	case "registry":
		return runIdentityRegistry(args[1:])
	default:
		return errors.New("usage: aip2p identity <init|create-hd|derive|list|export|recover|registry> [flags]")
	}
}

func runIdentityInit(args []string) error {
	fs := flag.NewFlagSet("identity init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	agentID := fs.String("agent-id", "", "stable agent id")
	author := fs.String("author", "", "default author for this identity")
	out := fs.String("out", "", "identity file output path; defaults to ~/.aip2p-news/identities/<sanitized-agent-id>.json")
	force := fs.Bool("force", false, "overwrite output file if it exists")
	if err := fs.Parse(args); err != nil {
		return err
	}
	outputPath, err := defaultIdentityOutputPath(*agentID, *out)
	if err != nil {
		return err
	}
	if !*force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("identity file already exists: %s", outputPath)
		}
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	identity, err := aip2p.NewAgentIdentity(*agentID, *author, time.Now().UTC())
	if err != nil {
		return err
	}
	if err := aip2p.SaveAgentIdentity(outputPath, identity); err != nil {
		return err
	}
	return writeJSON(map[string]any{
		"agent_id":   identity.AgentID,
		"author":     identity.Author,
		"key_type":   identity.KeyType,
		"public_key": identity.PublicKey,
		"created_at": identity.CreatedAt,
		"file":       outputPath,
	})
}

func runIdentityCreateHD(args []string) error {
	fs := flag.NewFlagSet("identity create-hd", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	agentID := fs.String("agent-id", "", "stable agent id")
	author := fs.String("author", "", "author URI (e.g., agent://alice)")
	out := fs.String("out", "", "identity file output path")
	force := fs.Bool("force", false, "overwrite output file if it exists")
	encrypt := fs.Bool("encrypt", false, "encrypt mnemonic with password")
	password := fs.String("password", "", "password for mnemonic encryption (prompted if --encrypt and not provided)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	outputPath, err := defaultIdentityOutputPath(*agentID, *out)
	if err != nil {
		return err
	}
	if !*force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("identity file already exists: %s", outputPath)
		}
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	identity, err := aip2p.NewHDIdentity(*agentID, *author, time.Now().UTC())
	if err != nil {
		return err
	}

	mnemonic := identity.Mnemonic // save before encryption clears it

	if *encrypt {
		pw := *password
		if pw == "" {
			fmt.Fprint(os.Stderr, "Enter password to encrypt mnemonic: ")
			fmt.Scanln(&pw)
			if pw == "" {
				return errors.New("password is required for encryption")
			}
		}
		if err := aip2p.SaveEncryptedIdentity(outputPath, identity, pw); err != nil {
			return err
		}
	} else {
		if err := aip2p.SaveAgentIdentity(outputPath, identity); err != nil {
			return err
		}
	}

	return writeJSON(map[string]any{
		"agent_id":        identity.AgentID,
		"author":          identity.Author,
		"key_type":        identity.KeyType,
		"public_key":      identity.PublicKey,
		"master_pubkey":   identity.MasterPubKey,
		"derivation_path": identity.DerivationPath,
		"mnemonic":        mnemonic,
		"encrypted":       *encrypt,
		"created_at":      identity.CreatedAt,
		"file":            outputPath,
		"warning":         "SAVE YOUR MNEMONIC! It cannot be recovered.",
	})
}

func runIdentityDerive(args []string) error {
	fs := flag.NewFlagSet("identity derive", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	parent := fs.String("parent", "", "parent identity file path")
	path := fs.String("path", "", "derivation path (e.g., m/0/0)")
	out := fs.String("out", "", "output file path")
	force := fs.Bool("force", false, "overwrite output file if it exists")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *parent == "" {
		return errors.New("--parent is required")
	}
	if *path == "" {
		return errors.New("--path is required")
	}

	parentIdentity, err := aip2p.LoadAgentIdentity(*parent)
	if err != nil {
		return fmt.Errorf("failed to load parent identity: %w", err)
	}

	childIdentity, err := aip2p.DeriveChildIdentity(parentIdentity, *path)
	if err != nil {
		return fmt.Errorf("failed to derive child identity: %w", err)
	}

	outputPath := *out
	if outputPath == "" {
		outputPath = filepath.Join(filepath.Dir(*parent), childIdentity.AgentID+".json")
	}

	if !*force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("output file already exists: %s", outputPath)
		}
	}

	if err := aip2p.SaveAgentIdentity(outputPath, childIdentity); err != nil {
		return err
	}

	return writeJSON(map[string]any{
		"agent_id":        childIdentity.AgentID,
		"author":          childIdentity.Author,
		"parent":          childIdentity.Parent,
		"derivation_path": childIdentity.DerivationPath,
		"public_key":      childIdentity.PublicKey,
		"file":            outputPath,
	})
}

func runIdentityList(args []string) error {
	fs := flag.NewFlagSet("identity list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dir := fs.String("dir", "", "identities directory (default: ~/.aip2p/identities)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	identitiesDir := *dir
	if identitiesDir == "" {
		home, _ := os.UserHomeDir()
		identitiesDir = filepath.Join(home, ".aip2p", "identities")
	}

	entries, err := os.ReadDir(identitiesDir)
	if err != nil {
		return err
	}

	var identities []map[string]any
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(identitiesDir, entry.Name())
		identity, err := aip2p.LoadAgentIdentity(path)
		if err != nil {
			continue
		}
		identities = append(identities, map[string]any{
			"agent_id":        identity.AgentID,
			"author":          identity.Author,
			"hd_enabled":      identity.HDEnabled,
			"parent":          identity.Parent,
			"derivation_path": identity.DerivationPath,
			"file":            path,
		})
	}

	return writeJSON(identities)
}

func runIdentityExport(args []string) error {
	return errors.New("identity export not yet implemented")
}

func runIdentityRecover(args []string) error {
	fs := flag.NewFlagSet("identity recover", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	mnemonic := fs.String("mnemonic", "", "BIP39 mnemonic (24 words)")
	agentID := fs.String("agent-id", "", "agent ID")
	author := fs.String("author", "", "author URI")
	out := fs.String("out", "", "output file path")
	force := fs.Bool("force", false, "overwrite output file if it exists")
	encrypt := fs.Bool("encrypt", false, "encrypt mnemonic with password")
	password := fs.String("password", "", "password for mnemonic encryption")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *mnemonic == "" {
		return errors.New("--mnemonic is required")
	}
	if *agentID == "" {
		return errors.New("--agent-id is required")
	}

	// Validate mnemonic
	seed, err := aip2p.MnemonicToSeed(*mnemonic, "")
	if err != nil {
		return fmt.Errorf("invalid mnemonic: %w", err)
	}

	// Derive m/0 key
	masterKey, err := aip2p.NewMasterKey(seed)
	if err != nil {
		return err
	}
	childKey, err := masterKey.DeriveChild(0)
	if err != nil {
		return err
	}

	publicKey := childKey.PublicKey()
	privateKey := childKey.PrivateKey()

	identity := aip2p.AgentIdentity{
		AgentID:        *agentID,
		Author:         *author,
		KeyType:        aip2p.KeyTypeEd25519,
		PublicKey:      fmt.Sprintf("%x", publicKey),
		PrivateKey:     fmt.Sprintf("%x", privateKey),
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		HDEnabled:      true,
		Mnemonic:       *mnemonic,
		MasterPubKey:   fmt.Sprintf("%x", masterKey.PublicKey()),
		DerivationPath: "m/0",
	}

	outputPath, err := defaultIdentityOutputPath(*agentID, *out)
	if err != nil {
		return err
	}
	if !*force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("output file already exists: %s", outputPath)
		}
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	if *encrypt {
		pw := *password
		if pw == "" {
			fmt.Fprint(os.Stderr, "Enter password to encrypt mnemonic: ")
			fmt.Scanln(&pw)
			if pw == "" {
				return errors.New("password is required for encryption")
			}
		}
		if err := aip2p.SaveEncryptedIdentity(outputPath, identity, pw); err != nil {
			return err
		}
	} else {
		if err := aip2p.SaveAgentIdentity(outputPath, identity); err != nil {
			return err
		}
	}

	return writeJSON(map[string]any{
		"agent_id":        identity.AgentID,
		"author":          identity.Author,
		"derivation_path": identity.DerivationPath,
		"encrypted":       *encrypt,
		"file":            outputPath,
		"message":         "Identity recovered successfully",
	})
}

func runIdentityRegistry(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p identity registry <add|list|remove> [flags]")
	}
	home, _ := os.UserHomeDir()
	registryPath := filepath.Join(home, ".aip2p", "identity_registry.json")

	switch args[0] {
	case "add":
		fs := flag.NewFlagSet("identity registry add", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		author := fs.String("author", "", "author URI (e.g., agent://alice)")
		pubkey := fs.String("pubkey", "", "master public key (ed25519:...)")
		trust := fs.String("trust", "known", "trust level (trusted, known, unknown)")
		notes := fs.String("notes", "", "optional notes")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *author == "" {
			return errors.New("--author is required")
		}
		if *pubkey == "" {
			return errors.New("--pubkey is required")
		}
		registry, err := aip2p.LoadIdentityRegistry(registryPath)
		if err != nil {
			return err
		}
		registry.Add(*author, *pubkey, *trust, *notes)
		if err := registry.Save(registryPath); err != nil {
			return err
		}
		return writeJSON(map[string]any{
			"action": "added",
			"author": *author,
			"pubkey": *pubkey,
			"trust":  *trust,
			"file":   registryPath,
		})

	case "list":
		registry, err := aip2p.LoadIdentityRegistry(registryPath)
		if err != nil {
			return err
		}
		entries := make([]map[string]any, 0)
		for author, entry := range registry.List() {
			entries = append(entries, map[string]any{
				"author":        author,
				"master_pubkey": entry.MasterPubKey,
				"trust_level":   entry.TrustLevel,
				"added_at":      entry.AddedAt,
				"notes":         entry.Notes,
			})
		}
		return writeJSON(entries)

	case "remove":
		fs := flag.NewFlagSet("identity registry remove", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		author := fs.String("author", "", "author URI to remove")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *author == "" {
			return errors.New("--author is required")
		}
		registry, err := aip2p.LoadIdentityRegistry(registryPath)
		if err != nil {
			return err
		}
		if !registry.Remove(*author) {
			return fmt.Errorf("identity %s not found in registry", *author)
		}
		if err := registry.Save(registryPath); err != nil {
			return err
		}
		return writeJSON(map[string]any{
			"action": "removed",
			"author": *author,
		})

	default:
		return errors.New("usage: aip2p identity registry <add|list|remove> [flags]")
	}
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
	extensionsRoot := fs.String("extensions-root", "", "installed extensions root; defaults to ~/.aip2p/extensions")
	pluginID := fs.String("plugin", "", "single built-in plugin id; ignored when --plugins is set")
	pluginsCSV := fs.String("plugins", "", "comma-separated built-in plugin ids to compose; overrides --plugin")
	pluginDirsCSV := fs.String("plugin-dir", "", "comma-separated external plugin directories containing aip2p.plugin.json")
	themeID := fs.String("theme", "", "theme id; defaults to the plugin default theme")
	themeDir := fs.String("theme-dir", "", "directory theme override; expects aip2p.theme.json plus templates/static")
	project := fs.String("project", "", "project id override")
	version := fs.String("version", "dev", "host version label")
	runtimeRoot := fs.String("runtime-root", "", "application runtime root")
	storeRoot := fs.String("store", "", "store root override")
	apiAddr := fs.String("api-addr", "127.0.0.1:51819", "agent API listen address (localhost only)")
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
		ExtensionsRoot:   *extensionsRoot,
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
	log.Printf("AiP2P host serving plugin=%s theme=%s on http://%s", instance.Site().Manifest.ID, instance.Site().Theme.ID, instance.ListenAddr())

	// Start Agent API server on localhost
	apiStoreRoot := *storeRoot
	if apiStoreRoot == "" {
		home, _ := os.UserHomeDir()
		apiStoreRoot = filepath.Join(home, ".aip2p", "store")
	}
	apiStore, err := aip2p.OpenStore(apiStoreRoot)
	if err != nil {
		log.Printf("warning: agent API store init failed: %v", err)
	} else {
		apiServer := aip2p.NewAPIServer(apiStore)
		apiHTTP := &http.Server{
			Addr:    *apiAddr,
			Handler: apiServer.Handler(),
		}
		go func() {
			log.Printf("AiP2P agent API on http://%s", *apiAddr)
			if err := apiHTTP.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("agent API error: %v", err)
			}
		}()
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			apiHTTP.Shutdown(shutdownCtx)
		}()
	}

	return instance.ListenAndServe(ctx)
}

func runPlugins(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p plugins <list|inspect|install|link|remove>")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("plugins list", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		registry := builtin.DefaultRegistry()
		store, err := extensions.Open(*root)
		if err != nil {
			return err
		}
		installed, err := store.ListPlugins()
		if err != nil {
			return err
		}
		plugins := make([]any, 0, len(registry.PluginManifests())+len(installed))
		for _, manifest := range registry.PluginManifests() {
			plugins = append(plugins, map[string]any{
				"source":   "builtin",
				"manifest": manifest,
			})
		}
		for _, entry := range installed {
			plugins = append(plugins, map[string]any{
				"source":   "installed",
				"root":     entry.Root,
				"manifest": entry.Manifest,
				"config":   entry.Config,
				"metadata": entry.Metadata,
			})
		}
		return writeJSON(struct {
			Plugins []any `json:"plugins"`
		}{
			Plugins: plugins,
		})
	case "inspect":
		fs := flag.NewFlagSet("plugins inspect", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		dir := fs.String("dir", "", "plugin directory containing aip2p.plugin.json")
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		registry := builtin.DefaultRegistry()
		store, err := extensions.Open(*root)
		if err != nil {
			return err
		}
		if _, err := store.RegisterIntoRegistry(registry, "", "", ""); err != nil {
			return err
		}
		if strings.TrimSpace(*dir) != "" {
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
		}
		if fs.NArg() == 0 {
			return errors.New("plugin id or --dir is required")
		}
		id := fs.Arg(0)
		if entry, err := store.GetPlugin(id); err == nil {
			resolved, err := workspace.ValidatePluginManifest(entry.Manifest, registry)
			if err != nil {
				return err
			}
			resolved.Root = entry.Root
			resolved.Config = entry.Config
			return writeJSON(struct {
				Source   string                     `json:"source"`
				Root     string                     `json:"root"`
				Manifest apphost.PluginManifest     `json:"manifest"`
				Config   map[string]any             `json:"config,omitempty"`
				Metadata extensions.InstallMetadata `json:"metadata"`
				Resolved workspace.ResolvedPlugin   `json:"resolved"`
			}{
				Source:   "installed",
				Root:     entry.Root,
				Manifest: entry.Manifest,
				Config:   entry.Config,
				Metadata: entry.Metadata,
				Resolved: resolved,
			})
		}
		_, manifest, err := registry.ResolvePlugin(id)
		if err != nil {
			return err
		}
		return writeJSON(struct {
			Source   string                 `json:"source"`
			Manifest apphost.PluginManifest `json:"manifest"`
		}{
			Source:   "builtin",
			Manifest: manifest,
		})
	case "install", "link":
		fs := flag.NewFlagSet("plugins "+args[0], flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		dir := fs.String("dir", "", "plugin directory containing aip2p.plugin.json")
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		store, err := extensions.Open(*root)
		if err != nil {
			return err
		}
		entry, err := store.InstallPlugin(*dir, args[0] == "link")
		if err != nil {
			return err
		}
		return writeJSON(entry)
	case "remove":
		fs := flag.NewFlagSet("plugins remove", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() == 0 {
			return errors.New("plugin id is required")
		}
		store, err := extensions.Open(*root)
		if err != nil {
			return err
		}
		if err := store.RemovePlugin(fs.Arg(0)); err != nil {
			return err
		}
		return writeJSON(map[string]any{"removed": fs.Arg(0)})
	default:
		return errors.New("usage: aip2p plugins <list|inspect|install|link|remove>")
	}
}

func runApps(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p apps <list|inspect|validate|install|link|remove>")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("apps list", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		store, err := extensions.Open(*root)
		if err != nil {
			return err
		}
		installed, err := store.ListApps()
		if err != nil {
			return err
		}
		apps := make([]any, 0, len(builtin.DefaultApps())+len(installed))
		for _, app := range builtin.DefaultApps() {
			apps = append(apps, map[string]any{
				"source": "builtin",
				"app":    app,
			})
		}
		for _, entry := range installed {
			apps = append(apps, map[string]any{
				"source":   "installed",
				"root":     entry.Root,
				"app":      entry.Manifest,
				"config":   entry.Config,
				"metadata": entry.Metadata,
			})
		}
		return writeJSON(struct {
			Apps []any `json:"apps"`
		}{
			Apps: apps,
		})
	case "inspect":
		fs := flag.NewFlagSet("apps inspect", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		dir := fs.String("dir", "", "application directory containing aip2p.app.json")
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*dir) != "" {
			bundle, report, err := inspectAppDir(*dir, *root)
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
		}
		if fs.NArg() == 0 {
			return errors.New("app id or --dir is required")
		}
		store, err := extensions.Open(*root)
		if err != nil {
			return err
		}
		entry, err := store.GetApp(fs.Arg(0))
		if err != nil {
			return err
		}
		bundle, report, err := inspectAppDir(entry.Root, *root)
		if err != nil {
			return err
		}
		return writeJSON(struct {
			Source     string                     `json:"source"`
			Root       string                     `json:"root"`
			Metadata   extensions.InstallMetadata `json:"metadata"`
			App        apphost.AppManifest        `json:"app"`
			Config     workspace.AppConfig        `json:"config"`
			Plugins    []apphost.PluginManifest   `json:"plugins"`
			Themes     []apphost.ThemeManifest    `json:"themes"`
			Validation workspace.ValidationReport `json:"validation"`
		}{
			Source:     "installed",
			Root:       entry.Root,
			Metadata:   entry.Metadata,
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
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		target := strings.TrimSpace(*dir)
		if target == "" {
			if fs.NArg() == 0 {
				return errors.New("app id or --dir is required")
			}
			store, err := extensions.Open(*root)
			if err != nil {
				return err
			}
			entry, err := store.GetApp(fs.Arg(0))
			if err != nil {
				return err
			}
			target = entry.Root
		}
		_, report, err := inspectAppDir(target, *root)
		if err != nil {
			return err
		}
		return writeJSON(report)
	case "install", "link":
		fs := flag.NewFlagSet("apps "+args[0], flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		dir := fs.String("dir", "", "application directory containing aip2p.app.json")
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		store, err := extensions.Open(*root)
		if err != nil {
			return err
		}
		entry, err := store.InstallApp(*dir, args[0] == "link")
		if err != nil {
			return err
		}
		return writeJSON(entry)
	case "remove":
		fs := flag.NewFlagSet("apps remove", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() == 0 {
			return errors.New("app id is required")
		}
		store, err := extensions.Open(*root)
		if err != nil {
			return err
		}
		if err := store.RemoveApp(fs.Arg(0)); err != nil {
			return err
		}
		return writeJSON(map[string]any{"removed": fs.Arg(0)})
	default:
		return errors.New("usage: aip2p apps <list|inspect|validate|install|link|remove>")
	}
}

func runCreate(args []string) error {
	if len(args) < 2 {
		return errors.New("usage: aip2p create <plugin|theme|app> <name> [--out dir]")
	}
	kind := strings.TrimSpace(args[0])
	target := strings.TrimSpace(args[1])
	if target == "" {
		return errors.New("name is required")
	}
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	outDir := fs.String("out", "", "output directory")
	if err := fs.Parse(args[2:]); err != nil {
		return err
	}
	name, resolvedOut, err := resolveCreateTarget(target, *outDir)
	if err != nil {
		return err
	}

	var (
		files []scaffold.File
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
	if err := scaffold.WriteFiles(resolvedOut, files); err != nil {
		return err
	}
	return writeJSON(map[string]any{
		"kind":   kind,
		"name":   name,
		"output": resolvedOut,
		"files":  filePaths(files),
	})
}

func resolveCreateTarget(target, explicitOut string) (string, string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", "", errors.New("name is required")
	}
	explicitOut = strings.TrimSpace(explicitOut)
	if explicitOut != "" {
		return targetBaseName(target), explicitOut, nil
	}
	if looksLikePath(target) {
		return targetBaseName(target), target, nil
	}
	return target, scaffold.Slug(target), nil
}

func looksLikePath(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	if filepath.IsAbs(value) {
		return true
	}
	switch value {
	case ".", "..":
		return true
	}
	return strings.Contains(value, "/") || strings.Contains(value, `\`)
}

func targetBaseName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	base := filepath.Base(filepath.Clean(value))
	base = strings.TrimSpace(base)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return value
	}
	return base
}

func defaultIdentityOutputPath(agentID, explicitOut string) (string, error) {
	explicitOut = strings.TrimSpace(explicitOut)
	if explicitOut != "" {
		return explicitOut, nil
	}
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return "", errors.New("agent-id is required")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	home = strings.TrimSpace(home)
	if home == "" {
		return "", errors.New("user home directory is empty")
	}
	return filepath.Join(home, ".aip2p-news", "identities", sanitizeAgentIDForFilename(agentID)+".json"), nil
}

func sanitizeAgentIDForFilename(agentID string) string {
	agentID = strings.ToLower(strings.TrimSpace(agentID))
	if agentID == "" {
		return "identity"
	}
	var b strings.Builder
	lastDash := false
	for _, r := range agentID {
		isAlnum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlnum {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	value := strings.Trim(b.String(), "-")
	if value == "" {
		return "identity"
	}
	return value
}

func parseFlagSetInterspersed(fs *flag.FlagSet, args []string) error {
	reordered := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positionals = append(positionals, arg)
			continue
		}
		reordered = append(reordered, arg)
		if strings.Contains(arg, "=") {
			continue
		}
		name := strings.TrimLeft(arg, "-")
		if name == "" {
			continue
		}
		info := fs.Lookup(name)
		if info == nil {
			continue
		}
		if bf, ok := info.Value.(boolFlag); ok && bf.IsBoolFlag() {
			continue
		}
		if i+1 < len(args) {
			i++
			reordered = append(reordered, args[i])
		}
	}
	reordered = append(reordered, positionals...)
	return fs.Parse(reordered)
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
		return errors.New("usage: aip2p themes <list|inspect|install|link|remove>")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("themes list", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		registry := builtin.DefaultRegistry()
		store, err := extensions.Open(*root)
		if err != nil {
			return err
		}
		installed, err := store.ListThemes()
		if err != nil {
			return err
		}
		themes := make([]any, 0, len(registry.ThemeManifests())+len(installed))
		for _, manifest := range registry.ThemeManifests() {
			themes = append(themes, map[string]any{
				"source":   "builtin",
				"manifest": manifest,
			})
		}
		for _, entry := range installed {
			themes = append(themes, map[string]any{
				"source":   "installed",
				"root":     entry.Root,
				"manifest": entry.Manifest,
				"metadata": entry.Metadata,
			})
		}
		return writeJSON(struct {
			Themes []any `json:"themes"`
		}{
			Themes: themes,
		})
	case "inspect":
		fs := flag.NewFlagSet("themes inspect", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		dir := fs.String("dir", "", "theme directory containing aip2p.theme.json")
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		if strings.TrimSpace(*dir) != "" {
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
		}
		if fs.NArg() == 0 {
			return errors.New("theme id or --dir is required")
		}
		store, err := extensions.Open(*root)
		if err != nil {
			return err
		}
		if entry, err := store.GetTheme(fs.Arg(0)); err == nil {
			return writeJSON(struct {
				Source   string                     `json:"source"`
				Root     string                     `json:"root"`
				Manifest apphost.ThemeManifest      `json:"manifest"`
				Metadata extensions.InstallMetadata `json:"metadata"`
			}{
				Source:   "installed",
				Root:     entry.Root,
				Manifest: entry.Manifest,
				Metadata: entry.Metadata,
			})
		}
		registry := builtin.DefaultRegistry()
		_, manifest, err := registry.ResolveTheme(fs.Arg(0))
		if err != nil {
			return err
		}
		return writeJSON(struct {
			Source   string                `json:"source"`
			Manifest apphost.ThemeManifest `json:"manifest"`
		}{
			Source:   "builtin",
			Manifest: manifest,
		})
	case "install", "link":
		fs := flag.NewFlagSet("themes "+args[0], flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		dir := fs.String("dir", "", "theme directory containing aip2p.theme.json")
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		store, err := extensions.Open(*root)
		if err != nil {
			return err
		}
		entry, err := store.InstallTheme(*dir, args[0] == "link")
		if err != nil {
			return err
		}
		return writeJSON(entry)
	case "remove":
		fs := flag.NewFlagSet("themes remove", flag.ContinueOnError)
		fs.SetOutput(os.Stderr)
		root := fs.String("root", "", "extensions root override")
		if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
			return err
		}
		if fs.NArg() == 0 {
			return errors.New("theme id is required")
		}
		store, err := extensions.Open(*root)
		if err != nil {
			return err
		}
		if err := store.RemoveTheme(fs.Arg(0)); err != nil {
			return err
		}
		return writeJSON(map[string]any{"removed": fs.Arg(0)})
	default:
		return errors.New("usage: aip2p themes <list|inspect|install|link|remove>")
	}
}

func manifestsToAny[T any](items []T) []any {
	out := make([]any, 0, len(items))
	for _, item := range items {
		out = append(out, item)
	}
	return out
}

func inspectAppDir(dir, extensionsRoot string) (workspace.AppBundle, workspace.ValidationReport, error) {
	bundle, err := workspace.LoadAppBundle(dir)
	if err != nil {
		return workspace.AppBundle{}, workspace.ValidationReport{}, err
	}
	registry := builtin.DefaultRegistry()
	store, err := extensions.Open(extensionsRoot)
	if err != nil {
		return workspace.AppBundle{}, workspace.ValidationReport{}, err
	}
	if _, err := store.RegisterIntoRegistry(registry, "", "", bundle.App.ID); err != nil {
		return workspace.AppBundle{}, workspace.ValidationReport{}, err
	}
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
	return errors.New(`usage: aip2p <command> [flags]

Commands:
  identity     Manage agent identity (init)
  publish      Publish a signed message bundle
  verify       Verify a message bundle
  show         Show a message bundle
  sync         Run sync daemon
  serve        Start HTTP server
  plugins      Manage plugins (list, inspect, install, link, remove)
  themes       Manage themes (list, inspect, install, link, remove)
  apps         Manage apps (list, inspect, install, link, remove)
  create       Scaffold new plugin, theme, or app
  peers        Show connected peers (list, health)
  subscribe    Manage subscriptions (add, remove, list)
  store        Manage bundle storage (status, clean)
  config       Manage configuration (get, set, show)
  sync-status  Show sync daemon status
  bootstrap    Manage bootstrap nodes (refresh, list, add)`)
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

// v0.3: peers command - show connected peers and network health
func runPeers(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p peers <list|health>")
	}
	fs := flag.NewFlagSet("peers", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	storeRoot := fs.String("store", ".aip2p", "store root")
	if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
		return err
	}

	statusPath := filepath.Join(*storeRoot, "sync", "status.json")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		return fmt.Errorf("sync daemon not running or status unavailable: %w", err)
	}

	var status map[string]any
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}

	switch args[0] {
	case "list":
		libp2p, _ := status["libp2p"].(map[string]any)
		if libp2p == nil {
			fmt.Println("No libp2p status available.")
			return nil
		}
		fmt.Printf("Peer ID: %v\n", libp2p["peer_id"])
		fmt.Printf("Connected peers: %v\n", libp2p["connected_peers"])
		fmt.Printf("Routing table: %v\n", libp2p["routing_table_peers"])
		if peers, ok := libp2p["bootstrap_peers"].([]any); ok {
			fmt.Printf("\nBootstrap peers (%d):\n", len(peers))
			for _, p := range peers {
				if pm, ok := p.(map[string]any); ok {
					connected := "disconnected"
					if c, _ := pm["connected"].(bool); c {
						connected = "connected"
					}
					fmt.Printf("  %v  %v  [%s]\n", pm["peer_id"], pm["address"], connected)
				}
			}
		}
		if mdnsStatus, ok := libp2p["mdns"].(map[string]any); ok {
			fmt.Printf("\nmDNS discovered: %v  connected: %v\n", mdnsStatus["discovered_peers"], mdnsStatus["connected_peers"])
		}

	case "health":
		fmt.Println("=== Network Health ===")
		libp2p, _ := status["libp2p"].(map[string]any)
		if libp2p != nil {
			connected, _ := libp2p["connected_peers"].(float64)
			routing, _ := libp2p["routing_table_peers"].(float64)
			fmt.Printf("Connected peers: %.0f\n", connected)
			fmt.Printf("Routing table:   %.0f\n", routing)
			if mdnsStatus, ok := libp2p["mdns"].(map[string]any); ok {
				fmt.Printf("mDNS discovered: %v\n", mdnsStatus["discovered_peers"])
			}
		}
		bt, _ := status["bittorrent_dht"].(map[string]any)
		if bt != nil {
			fmt.Printf("BT DHT good nodes: %v\n", bt["good_nodes"])
		}
		activity, _ := status["sync_activity"].(map[string]any)
		if activity != nil {
			fmt.Printf("Imported: %v  Skipped: %v  Failed: %v\n",
				activity["imported_count"], activity["skipped_count"], activity["failed_count"])
		}
		fmt.Printf("Updated at: %v\n", status["updated_at"])

	default:
		return fmt.Errorf("unknown peers subcommand: %s (use list or health)", args[0])
	}
	return nil
}

// v0.3: subscribe command - manage topic subscriptions
func runSubscribe(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p subscribe <add|remove|list>")
	}
	fs := flag.NewFlagSet("subscribe", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	rulesPath := fs.String("rules", "", "path to subscriptions.json")
	if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
		return err
	}

	path := strings.TrimSpace(*rulesPath)
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, ".aip2p-news", "subscriptions.json")
	}

	switch args[0] {
	case "list":
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No subscriptions configured (subscribing to all).")
				return nil
			}
			return err
		}
		var rules map[string]any
		if err := json.Unmarshal(data, &rules); err != nil {
			return err
		}
		out, _ := json.MarshalIndent(rules, "", "  ")
		fmt.Println(string(out))

	case "add":
		positional := fs.Args()
		if len(positional) == 0 {
			return errors.New("usage: aip2p subscribe add <topic>")
		}
		topic := strings.TrimSpace(positional[0])
		rules := loadOrCreateRules(path)
		topics, _ := rules["topics"].([]any)
		for _, t := range topics {
			if strings.EqualFold(fmt.Sprint(t), topic) {
				fmt.Printf("Topic %q already subscribed.\n", topic)
				return nil
			}
		}
		rules["topics"] = append(topics, topic)
		return saveRules(path, rules)

	case "remove":
		positional := fs.Args()
		if len(positional) == 0 {
			return errors.New("usage: aip2p subscribe remove <topic>")
		}
		topic := strings.TrimSpace(positional[0])
		rules := loadOrCreateRules(path)
		topics, _ := rules["topics"].([]any)
		filtered := make([]any, 0, len(topics))
		for _, t := range topics {
			if !strings.EqualFold(fmt.Sprint(t), topic) {
				filtered = append(filtered, t)
			}
		}
		rules["topics"] = filtered
		return saveRules(path, rules)

	default:
		return fmt.Errorf("unknown subscribe subcommand: %s (use add, remove, or list)", args[0])
	}
	return nil
}

func loadOrCreateRules(path string) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{}
	}
	var rules map[string]any
	if err := json.Unmarshal(data, &rules); err != nil {
		return map[string]any{}
	}
	return rules
}

func saveRules(path string, rules map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	fmt.Printf("Saved %s\n", path)
	return nil
}

// v0.3: store command - manage local bundle storage
func runStore(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p store <status|clean>")
	}
	fs := flag.NewFlagSet("store", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	storeRoot := fs.String("store", ".aip2p", "store root")
	if err := parseFlagSetInterspersed(fs, args[1:]); err != nil {
		return err
	}

	switch args[0] {
	case "status":
		dataDir := filepath.Join(*storeRoot, "data")
		torrentDir := filepath.Join(*storeRoot, "torrents")

		bundleCount := 0
		var totalSize int64
		if entries, err := os.ReadDir(dataDir); err == nil {
			for _, e := range entries {
				if e.IsDir() {
					bundleCount++
					dirPath := filepath.Join(dataDir, e.Name())
					if files, err := os.ReadDir(dirPath); err == nil {
						for _, f := range files {
							if info, err := f.Info(); err == nil {
								totalSize += info.Size()
							}
						}
					}
				}
			}
		}

		torrentCount := 0
		if entries, err := os.ReadDir(torrentDir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".torrent") {
					torrentCount++
				}
			}
		}

		fmt.Printf("Store root:  %s\n", *storeRoot)
		fmt.Printf("Bundles:     %d\n", bundleCount)
		fmt.Printf("Torrents:    %d\n", torrentCount)
		fmt.Printf("Total size:  %.2f MB\n", float64(totalSize)/(1024*1024))

	case "clean":
		dataDir := filepath.Join(*storeRoot, "data")
		entries, err := os.ReadDir(dataDir)
		if err != nil {
			return err
		}
		removed := 0
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			dirPath := filepath.Join(dataDir, e.Name())
			msgPath := filepath.Join(dirPath, "aip2p-message.json")
			data, err := os.ReadFile(msgPath)
			if err != nil {
				continue
			}
			var msg struct {
				ExpiresAt string `json:"expires_at"`
			}
			if err := json.Unmarshal(data, &msg); err != nil {
				continue
			}
			if msg.ExpiresAt == "" {
				continue
			}
			expires, err := time.Parse(time.RFC3339, msg.ExpiresAt)
			if err != nil {
				continue
			}
			if time.Now().UTC().After(expires) {
				if err := os.RemoveAll(dirPath); err == nil {
					removed++
					fmt.Printf("Removed expired bundle: %s\n", e.Name())
				}
			}
		}
		fmt.Printf("Cleaned %d expired bundles.\n", removed)

	default:
		return fmt.Errorf("unknown store subcommand: %s (use status or clean)", args[0])
	}
	return nil
}

// v0.3: bootstrap command - manage bootstrap node list
func runBootstrap(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p bootstrap <refresh|list|add>")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	bootstrapPath := filepath.Join(home, ".aip2p", "bootstrap.list")

	switch args[0] {
	case "refresh":
		fmt.Println("Fetching bootstrap list from remote sources...")
		urls := []string{
			"https://raw.githubusercontent.com/AiP2P/AiP2P/main/bootstrap.list",
			"https://cdn.gh-proxy.org/https://raw.githubusercontent.com/AiP2P/AiP2P/main/bootstrap.list",
		}
		var fetched []string
		for _, url := range urls {
			nodes, err := fetchBootstrapList(url)
			if err != nil {
				fmt.Printf("  [skip] %s: %v\n", url, err)
				continue
			}
			fetched = append(fetched, nodes...)
			fmt.Printf("  [ok] %s: %d nodes\n", url, len(nodes))
			break
		}
		if len(fetched) == 0 {
			fmt.Println("No remote bootstrap nodes fetched. Using defaults.")
			fetched = defaultBootstrapNodes()
		}
		// Merge with existing
		existing := loadBootstrapList(bootstrapPath)
		merged := mergeBootstrapNodes(existing, fetched)
		if err := saveBootstrapList(bootstrapPath, merged); err != nil {
			return err
		}
		fmt.Printf("Saved %d bootstrap nodes to %s\n", len(merged), bootstrapPath)

	case "list":
		nodes := loadBootstrapList(bootstrapPath)
		if len(nodes) == 0 {
			fmt.Println("No bootstrap nodes configured. Run 'aip2p bootstrap refresh' to fetch.")
			return nil
		}
		fmt.Printf("Bootstrap nodes (%d):\n", len(nodes))
		for _, n := range nodes {
			fmt.Printf("  %s\n", n)
		}

	case "add":
		if len(args) < 2 {
			return errors.New("usage: aip2p bootstrap add <multiaddr>")
		}
		node := strings.TrimSpace(args[1])
		existing := loadBootstrapList(bootstrapPath)
		for _, n := range existing {
			if n == node {
				fmt.Println("Node already in bootstrap list.")
				return nil
			}
		}
		existing = append(existing, node)
		if err := saveBootstrapList(bootstrapPath, existing); err != nil {
			return err
		}
		fmt.Printf("Added %s to bootstrap list.\n", node)

	default:
		return fmt.Errorf("unknown bootstrap subcommand: %s (use refresh, list, or add)", args[0])
	}
	return nil
}

func fetchBootstrapList(url string) ([]string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	data := make([]byte, 0, 4096)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		data = append(data, buf[:n]...)
		if err != nil {
			break
		}
	}
	var nodes []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		nodes = append(nodes, line)
	}
	return nodes, nil
}

func defaultBootstrapNodes() []string {
	return []string{
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	}
}

func loadBootstrapList(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var nodes []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			nodes = append(nodes, line)
		}
	}
	return nodes
}

func mergeBootstrapNodes(existing, fetched []string) []string {
	seen := make(map[string]struct{})
	var merged []string
	for _, n := range existing {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		if _, ok := seen[n]; !ok {
			seen[n] = struct{}{}
			merged = append(merged, n)
		}
	}
	for _, n := range fetched {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		if _, ok := seen[n]; !ok {
			seen[n] = struct{}{}
			merged = append(merged, n)
		}
	}
	return merged
}

func saveBootstrapList(path string, nodes []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content := "# AiP2P bootstrap nodes\n# Updated by 'aip2p bootstrap refresh'\n\n"
	for _, n := range nodes {
		content += n + "\n"
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// v0.3: config command - manage ~/.aip2p/config.json
func runConfig(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p config <get|set|show>")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(home, ".aip2p", "config.json")

	switch args[0] {
	case "show":
		data, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("{}")
				return nil
			}
			return err
		}
		fmt.Println(string(data))

	case "get":
		if len(args) < 2 {
			return errors.New("usage: aip2p config get <key>")
		}
		key := strings.TrimSpace(args[1])
		cfg := loadOrCreateConfig(configPath)
		value, ok := cfg[key]
		if !ok {
			return fmt.Errorf("key %q not found", key)
		}
		switch v := value.(type) {
		case string:
			fmt.Println(v)
		default:
			out, _ := json.MarshalIndent(v, "", "  ")
			fmt.Println(string(out))
		}

	case "set":
		if len(args) < 3 {
			return errors.New("usage: aip2p config set <key> <value>")
		}
		key := strings.TrimSpace(args[1])
		rawValue := strings.TrimSpace(args[2])
		cfg := loadOrCreateConfig(configPath)
		var value any
		if err := json.Unmarshal([]byte(rawValue), &value); err != nil {
			value = rawValue
		}
		cfg[key] = value
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			return err
		}
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}
		data = append(data, '\n')
		if err := os.WriteFile(configPath, data, 0o644); err != nil {
			return err
		}
		fmt.Printf("Set %s = %v\n", key, value)

	default:
		return fmt.Errorf("unknown config subcommand: %s (use get, set, or show)", args[0])
	}
	return nil
}

func loadOrCreateConfig(path string) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{}
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return map[string]any{}
	}
	return cfg
}

// v0.3: sync-status command - show sync daemon status
func runSyncStatus(args []string) error {
	fs := flag.NewFlagSet("sync-status", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	storeRoot := fs.String("store", ".aip2p", "store root")
	if err := fs.Parse(args); err != nil {
		return err
	}

	statusPath := filepath.Join(*storeRoot, "sync", "status.json")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Sync daemon is not running (no status file found).")
			return nil
		}
		return err
	}

	var status map[string]any
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}

	fmt.Println("=== Sync Status ===")
	if v, ok := status["mode"]; ok {
		fmt.Printf("Mode:        %v\n", v)
	}
	if v, ok := status["started_at"]; ok {
		fmt.Printf("Started:     %v\n", v)
	}
	if v, ok := status["updated_at"]; ok {
		fmt.Printf("Updated:     %v\n", v)
	}
	if v, ok := status["pid"]; ok {
		fmt.Printf("PID:         %.0f\n", v)
	}
	if v, ok := status["network_id"]; ok {
		nid := fmt.Sprint(v)
		if len(nid) > 16 {
			nid = nid[:16] + "..."
		}
		fmt.Printf("Network:     %s\n", nid)
	}
	if v, ok := status["seed"]; ok {
		fmt.Printf("Seeding:     %v\n", v)
	}

	// Sync activity
	if activity, ok := status["sync_activity"].(map[string]any); ok {
		fmt.Println("\n--- Activity ---")
		fmt.Printf("Queue refs:  %v\n", activity["queue_refs_count"])
		fmt.Printf("Imported:    %v\n", activity["imported_count"])
		fmt.Printf("Skipped:     %v\n", activity["skipped_count"])
		fmt.Printf("Failed:      %v\n", activity["failed_count"])
		if v, ok := activity["last_status"]; ok && v != nil {
			fmt.Printf("Last status: %v\n", v)
		}
	}

	// PubSub
	if pubsub, ok := status["pubsub"].(map[string]any); ok {
		fmt.Println("\n--- PubSub ---")
		fmt.Printf("Published:   %v\n", pubsub["published_count"])
		fmt.Printf("Received:    %v\n", pubsub["received_count"])
		fmt.Printf("Enqueued:    %v\n", pubsub["enqueued_count"])
		if topics, ok := pubsub["joined_topics"].([]any); ok {
			fmt.Printf("Topics:      %d joined\n", len(topics))
		}
	}

	return nil
}
