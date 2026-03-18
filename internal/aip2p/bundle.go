package aip2p

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
)

// v0.3: Bundle manifest structures
const BundleManifestFileName = "manifest.json"

type BundleManifest struct {
	Version string       `json:"version"`
	Files   []BundleFile `json:"files"`
}

type BundleFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	Role   string `json:"role"` // main-metadata, main-body, attachment
}

type PublishResult struct {
	InfoHash    string `json:"infohash"`
	Magnet      string `json:"magnet"`
	TorrentFile string `json:"torrent_file"`
	ContentDir  string `json:"content_dir"`
}

func PublishMessage(store *Store, input MessageInput) (PublishResult, error) {
	msg, body, err := BuildMessage(input)
	if err != nil {
		return PublishResult{}, err
	}

	// v0.5: check child identity permissions if identity has children configured
	if input.Identity != nil {
		if err := CheckChildPermission(*input.Identity, msg.Author, msg.Channel, msg.Tags); err != nil {
			return PublishResult{}, err
		}
	}

	contentDir := store.NewContentDir(msg.Title, input.CreatedAt)
	if err := WriteMessage(contentDir, msg, body); err != nil {
		return PublishResult{}, err
	}

	// v0.3: generate bundle manifest
	if err := generateBundleManifest(contentDir, msg); err != nil {
		return PublishResult{}, err
	}

	info := metainfo.Info{
		PieceLength: 32 * 1024,
		Name:        filepath.Base(contentDir),
	}
	if err := info.BuildFromFilePath(contentDir); err != nil {
		return PublishResult{}, err
	}
	infoBytes, err := bencode.Marshal(info)
	if err != nil {
		return PublishResult{}, err
	}

	mi := metainfo.MetaInfo{
		CreationDate: time.Now().Unix(),
		Comment:      "AiP2P message bundle",
		CreatedBy:    "aip2p-go-reference",
		InfoBytes:    infoBytes,
	}
	mi.SetDefaults()

	infoHash := mi.HashInfoBytes().HexString()
	torrentPath := store.TorrentPath(infoHash)
	file, err := os.Create(torrentPath)
	if err != nil {
		return PublishResult{}, err
	}
	defer file.Close()
	if err := mi.Write(file); err != nil {
		return PublishResult{}, err
	}

	magnet := mi.Magnet(nil, &info).String()
	return PublishResult{
		InfoHash:    strings.ToLower(infoHash),
		Magnet:      magnet,
		TorrentFile: torrentPath,
		ContentDir:  contentDir,
	}, nil
}

// v0.3: generateBundleManifest creates manifest.json for a bundle
func generateBundleManifest(contentDir string, msg Message) error {
	files := []BundleFile{}

	// Add message.json (main-metadata)
	messageFile := filepath.Join(contentDir, MessageFileName)
	if stat, err := os.Stat(messageFile); err == nil {
		hash, err := fileSHA256(messageFile)
		if err != nil {
			return err
		}
		files = append(files, BundleFile{
			Path:   MessageFileName,
			SHA256: hash,
			Size:   stat.Size(),
			Role:   "main-metadata",
		})
	}

	// Add body file (main-body)
	bodyFile := filepath.Join(contentDir, msg.BodyFile)
	if stat, err := os.Stat(bodyFile); err == nil {
		hash, err := fileSHA256(bodyFile)
		if err != nil {
			return err
		}
		files = append(files, BundleFile{
			Path:   msg.BodyFile,
			SHA256: hash,
			Size:   stat.Size(),
			Role:   "main-body",
		})
	}

	// Scan for additional files (attachments)
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == MessageFileName || name == msg.BodyFile || name == BundleManifestFileName {
			continue
		}
		filePath := filepath.Join(contentDir, name)
		stat, err := os.Stat(filePath)
		if err != nil {
			continue
		}
		hash, err := fileSHA256(filePath)
		if err != nil {
			continue
		}
		files = append(files, BundleFile{
			Path:   name,
			SHA256: hash,
			Size:   stat.Size(),
			Role:   "attachment",
		})
	}

	manifest := BundleManifest{
		Version: "v0.3-bundle-1",
		Files:   files,
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	manifestPath := filepath.Join(contentDir, BundleManifestFileName)
	return os.WriteFile(manifestPath, data, 0o644)
}

// fileSHA256 computes SHA-256 hash of a file
func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// v0.3: VerifyBundleManifest validates bundle integrity
func VerifyBundleManifest(contentDir string) error {
	manifestPath := filepath.Join(contentDir, BundleManifestFileName)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		// v0.3: manifest is optional for backward compatibility
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var manifest BundleManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}

	// Verify each file's SHA-256
	for _, file := range manifest.Files {
		if file.Path == BundleManifestFileName {
			continue // skip self
		}
		filePath := filepath.Join(contentDir, file.Path)
		actualHash, err := fileSHA256(filePath)
		if err != nil {
			return err
		}
		if actualHash != file.SHA256 {
			return errors.New("manifest SHA-256 mismatch for " + file.Path)
		}
	}

	return nil
}

// v0.3: LoadBundleManifest loads manifest.json if present
func LoadBundleManifest(contentDir string) (*BundleManifest, error) {
	manifestPath := filepath.Join(contentDir, BundleManifestFileName)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // optional
		}
		return nil, err
	}

	var manifest BundleManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}
