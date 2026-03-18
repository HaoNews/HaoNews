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
)

const (
	ProtocolVersion = "aip2p/0.3"
	MessageFileName = "aip2p-message.json"
	BodyFileName    = "body.txt"
)

type Message struct {
	Protocol        string         `json:"protocol"`
	ProtocolVersion string         `json:"protocol_version,omitempty"` // v0.3: explicit version field
	Kind            string         `json:"kind"`
	Author          string         `json:"author"`
	CreatedAt       string         `json:"created_at"`
	Channel         string         `json:"channel,omitempty"`
	Title           string         `json:"title,omitempty"`
	BodyFile        string         `json:"body_file"`
	BodySHA256      string         `json:"body_sha256"`
	ContentType     string         `json:"content_type,omitempty"`     // v0.3: text/plain, text/markdown, application/json
	BodyEncoding    string         `json:"body_encoding,omitempty"`    // v0.3: utf-8, base64
	ReplyTo         *MessageLink   `json:"reply_to,omitempty"`
	Tags            []string       `json:"tags,omitempty"`
	Nonce           string         `json:"nonce,omitempty"`            // v0.3: anti-replay
	ExpiresAt       string         `json:"expires_at,omitempty"`       // v0.3: ISO 8601 expiry
	Encrypted       bool           `json:"encrypted,omitempty"`        // v0.3: whether body is encrypted
	EncryptedTo     []string       `json:"encrypted_to,omitempty"`     // v0.3: recipient agent URIs
	EncryptionMethod string        `json:"encryption_method,omitempty"` // v0.3: "age" / "libsodium"
	EncryptedKey    string         `json:"encrypted_key,omitempty"`    // v0.3: encrypted symmetric key
	Origin          *MessageOrigin `json:"origin,omitempty"`
	Extensions      map[string]any `json:"extensions,omitempty"`
}

type MessageLink struct {
	InfoHash  string `json:"infohash,omitempty"`
	Magnet    string `json:"magnet,omitempty"`
	Author    string `json:"author,omitempty"`    // v0.3: reply target author
	CreatedAt string `json:"created_at,omitempty"` // v0.3: reply target timestamp
}

type MessageOrigin struct {
	Author    string `json:"author"`
	AgentID   string `json:"agent_id"`
	KeyType   string `json:"key_type"`
	PublicKey string `json:"public_key"`
	Signature string `json:"signature"`
}

type MessageInput struct {
	Kind         string
	Author       string
	Channel      string
	Title        string
	Body         string
	BodyFile     string         // v0.3: custom body filename
	ContentType  string         // v0.3: MIME type
	BodyEncoding string         // v0.3: encoding
	ReplyTo      *MessageLink
	Tags         []string
	Nonce        string         // v0.3: anti-replay
	ExpiresAt    time.Time      // v0.3: expiry
	Identity     *AgentIdentity
	Extensions   map[string]any
	CreatedAt    time.Time
}

func (in MessageInput) Validate() error {
	if strings.TrimSpace(in.Author) == "" {
		return errors.New("author is required")
	}
	if strings.TrimSpace(in.Body) == "" {
		return errors.New("body is required")
	}
	if strings.TrimSpace(in.Kind) == "" {
		in.Kind = "post"
	}
	return nil
}

func BuildMessage(in MessageInput) (Message, []byte, error) {
	if err := in.Validate(); err != nil {
		return Message{}, nil, err
	}
	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now().UTC()
	}
	bodyBytes := []byte(in.Body)
	sum := sha256.Sum256(bodyBytes)

	// v0.3: default body_file based on content_type
	bodyFile := in.BodyFile
	if bodyFile == "" {
		bodyFile = defaultBodyFile(in.ContentType)
	}

	// v0.3: default content_type
	contentType := strings.TrimSpace(in.ContentType)
	if contentType == "" {
		contentType = "text/plain"
	}

	// v0.3: default body_encoding
	bodyEncoding := strings.TrimSpace(in.BodyEncoding)
	if bodyEncoding == "" {
		bodyEncoding = "utf-8"
	}

	msg := Message{
		Protocol:        ProtocolVersion,
		ProtocolVersion: "v0.3",
		Kind:            defaultKind(in.Kind),
		Author:          strings.TrimSpace(in.Author),
		CreatedAt:       in.CreatedAt.UTC().Format(time.RFC3339),
		Channel:         strings.TrimSpace(in.Channel),
		Title:           strings.TrimSpace(in.Title),
		BodyFile:        bodyFile,
		BodySHA256:      hex.EncodeToString(sum[:]),
		ContentType:     contentType,
		BodyEncoding:    bodyEncoding,
		ReplyTo:         canonicalMessageLink(in.ReplyTo),
		Tags:            cleanTags(in.Tags),
		Nonce:           strings.TrimSpace(in.Nonce),
		Extensions:      cloneMap(in.Extensions),
	}

	// v0.3: expires_at
	if !in.ExpiresAt.IsZero() {
		msg.ExpiresAt = in.ExpiresAt.UTC().Format(time.RFC3339)
	}

	if in.Identity != nil {
		origin, err := BuildSignedOrigin(msg, *in.Identity)
		if err != nil {
			return Message{}, nil, err
		}
		msg.Origin = origin
	}
	return msg, bodyBytes, nil
}

func WriteMessage(dir string, msg Message, body []byte) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// v0.3: use msg.BodyFile instead of hardcoded BodyFileName
	bodyFile := msg.BodyFile
	if bodyFile == "" {
		bodyFile = BodyFileName
	}
	bodyPath := filepath.Join(dir, bodyFile)
	if err := os.WriteFile(bodyPath, body, 0o644); err != nil {
		return err
	}
	messagePath := filepath.Join(dir, MessageFileName)
	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(messagePath, data, 0o644)
}

func LoadMessage(dir string) (Message, string, error) {
	data, err := os.ReadFile(filepath.Join(dir, MessageFileName))
	if err != nil {
		return Message{}, "", err
	}
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return Message{}, "", err
	}
	bodyBytes, err := os.ReadFile(filepath.Join(dir, msg.BodyFile))
	if err != nil {
		return Message{}, "", err
	}
	if err := ValidateMessage(msg, bodyBytes); err != nil {
		return Message{}, "", err
	}
	return msg, string(bodyBytes), nil
}

func ValidateMessage(msg Message, body []byte) error {
	// v0.3: backward compatibility - accept both aip2p/0.1 and aip2p/0.3
	if msg.Protocol != "aip2p/0.1" && msg.Protocol != ProtocolVersion {
		return errors.New("unsupported protocol version")
	}
	if strings.TrimSpace(msg.Author) == "" {
		return errors.New("author is required")
	}
	if msg.BodyFile == "" {
		return errors.New("body_file is required")
	}
	if _, err := time.Parse(time.RFC3339, msg.CreatedAt); err != nil {
		return errors.New("created_at must be RFC3339")
	}
	sum := sha256.Sum256(body)
	if hex.EncodeToString(sum[:]) != msg.BodySHA256 {
		return errors.New("body_sha256 mismatch")
	}
	// v0.3: validate expires_at if present
	if msg.ExpiresAt != "" {
		if _, err := time.Parse(time.RFC3339, msg.ExpiresAt); err != nil {
			return errors.New("expires_at must be RFC3339")
		}
	}
	if err := ValidateMessageOrigin(msg); err != nil {
		return err
	}
	return nil
}

func cleanTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func defaultKind(kind string) string {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return "post"
	}
	return kind
}

// v0.3: default body_file based on content_type
func defaultBodyFile(contentType string) string {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.Contains(contentType, "markdown"):
		return "body.md"
	case strings.Contains(contentType, "json"):
		return "body.json"
	case strings.Contains(contentType, "csv"):
		return "body.csv"
	default:
		return BodyFileName // body.txt
	}
}

func cloneMap(value map[string]any) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(value))
	for key, item := range value {
		cloned[key] = item
	}
	return cloned
}
