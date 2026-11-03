package scene

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Store errors. Callers must use errors.Is, never string matching.
var (
	ErrNotFound = errors.New("scene: not found")
	ErrConflict = errors.New("scene: version conflict")
	ErrReadOnly = errors.New("scene: store is read-only")
)

// Snapshot is the result of loading a document. Version is opaque to the
// caller and is sent back on Save for optimistic concurrency control.
type Snapshot struct {
	Ref      string
	Version  string
	Data     []byte
	Writable bool
}

// Store loads and saves scene documents. Resolve turns a reference relative to
// a document into a canonical reference the store understands.
type Store interface {
	Load(ctx context.Context, ref string) (*Snapshot, error)
	Save(ctx context.Context, ref string, data []byte, baseVersion string) (*Snapshot, error)
	Resolve(base, ref string) (string, error)
}

// VersionOf returns the content hash used as a version by the file, memory and
// HTTP stores.
func VersionOf(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// ResolveRef resolves ref against base the way a browser resolves a link.
// base and ref may be URLs (file, http(s), tie) or plain file paths.
func ResolveRef(base, ref string) (string, error) {
	if strings.TrimSpace(ref) == "" {
		return "", fmt.Errorf("scene: empty reference")
	}
	r, err := url.Parse(ref)
	if err == nil && r.IsAbs() {
		return r.String(), nil
	}
	if base == "" {
		return ref, nil
	}
	b, err := url.Parse(base)
	if err == nil && b.IsAbs() {
		out := b.ResolveReference(r).String()
		// Go prints a scheme with an empty host as scheme:///path; keep the
		// scheme:path form that tie uses. file:/// must keep its slashes.
		if b.Scheme != "file" {
			out = strings.Replace(out, b.Scheme+":///", b.Scheme+":/", 1)
		}
		return out, nil
	}
	if filepath.IsAbs(ref) {
		return filepath.Clean(ref), nil
	}
	return filepath.Clean(filepath.Join(filepath.Dir(base), ref)), nil
}

// FilePath converts a file reference into a local path.
func FilePath(ref string) (string, error) {
	u, err := url.Parse(ref)
	if err == nil && u.IsAbs() {
		switch u.Scheme {
		case "file":
			p, err := url.PathUnescape(u.Path)
			if err != nil {
				return "", fmt.Errorf("scene: bad file URL %q: %w", ref, err)
			}
			return filepath.FromSlash(p), nil
		case "":
			// fall through to the plain path below
		default:
			return "", fmt.Errorf("scene: %q is not a local file reference", ref)
		}
	}
	return ref, nil
}

// FileStore stores scene documents on the local filesystem. Save is atomic
// (temporary file + rename) and refuses to overwrite a newer version.
type FileStore struct{}

// Resolve resolves a file reference.
func (FileStore) Resolve(base, ref string) (string, error) {
	out, err := ResolveRef(base, ref)
	if err != nil {
		return "", err
	}
	if _, err := FilePath(out); err != nil {
		return "", err
	}
	return out, nil
}

// Load reads a document from disk.
func (FileStore) Load(ctx context.Context, ref string) (*Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := FilePath(ref)
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, path)
		}
		return nil, fmt.Errorf("scene: read %s: %w", path, err)
	}
	return &Snapshot{Ref: path, Version: VersionOf(data), Data: data, Writable: true}, nil
}

// Save atomically writes a document.
func (FileStore) Save(ctx context.Context, ref string, data []byte, baseVersion string) (*Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := FilePath(ref)
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}
	if baseVersion != "" {
		cur, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("%w: %s was removed", ErrConflict, path)
			}
			return nil, fmt.Errorf("scene: read %s: %w", path, err)
		}
		if VersionOf(cur) != baseVersion {
			return nil, fmt.Errorf("%w: %s changed on disk", ErrConflict, path)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("scene: create directory for %s: %w", path, err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".scene-*.tmp")
	if err != nil {
		return nil, fmt.Errorf("scene: create temporary file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return nil, fmt.Errorf("scene: write %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return nil, fmt.Errorf("scene: close %s: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return nil, fmt.Errorf("scene: replace %s: %w", path, err)
	}
	return &Snapshot{Ref: path, Version: VersionOf(data), Data: data, Writable: true}, nil
}

// MemoryStore is an in-memory store for tests and previews.
type MemoryStore struct {
	mu     sync.Mutex
	files  map[string]*Snapshot
	Writes int // number of successful saves, for tests
	Fail   error
}

// NewMemoryStore returns an empty memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{files: make(map[string]*Snapshot)}
}

// Put seeds a document without conflict checking.
func (m *MemoryStore) Put(ref string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[ref] = &Snapshot{Ref: ref, Version: VersionOf(data), Data: append([]byte(nil), data...), Writable: true}
}

// Resolve resolves a reference.
func (m *MemoryStore) Resolve(base, ref string) (string, error) { return ResolveRef(base, ref) }

// Load returns a stored document.
func (m *MemoryStore) Load(_ context.Context, ref string) (*Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.files[ref]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, ref)
	}
	cp := *s
	cp.Data = append([]byte(nil), s.Data...)
	return &cp, nil
}

// Save stores a document with optimistic concurrency.
func (m *MemoryStore) Save(_ context.Context, ref string, data []byte, baseVersion string) (*Snapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Fail != nil {
		return nil, m.Fail
	}
	if cur, ok := m.files[ref]; ok && cur.Version != baseVersion {
		return nil, fmt.Errorf("%w: %s", ErrConflict, ref)
	}
	if _, ok := m.files[ref]; !ok && baseVersion != "" {
		return nil, fmt.Errorf("%w: %s was removed", ErrConflict, ref)
	}
	s := &Snapshot{Ref: ref, Version: VersionOf(data), Data: append([]byte(nil), data...), Writable: true}
	m.files[ref] = s
	m.Writes++
	cp := *s
	cp.Data = append([]byte(nil), s.Data...)
	return &cp, nil
}

// HTTPStore loads scene documents over HTTP(S) and can save them with PUT
// (WebDAV-style) when Writable is set. Relative references must stay on the
// same origin, mirroring the script import policy.
type HTTPStore struct {
	Client   *http.Client
	Writable bool
}

// NewHTTPStore returns an HTTP store with a bounded client.
func NewHTTPStore(writable bool) *HTTPStore {
	return &HTTPStore{Client: &http.Client{Timeout: 30 * time.Second}, Writable: writable}
}

// Resolve resolves a reference and enforces same-origin for saving.
func (s *HTTPStore) Resolve(base, ref string) (string, error) {
	out, err := ResolveRef(base, ref)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(out)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("scene: %q is not an HTTP reference", out)
	}
	if base != "" {
		b, err := url.Parse(base)
		if err == nil && b.IsAbs() && origin(b) != origin(u) {
			return "", fmt.Errorf("scene: cross-origin reference %q (base %q)", out, base)
		}
	}
	return out, nil
}

func (s *HTTPStore) client() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return http.DefaultClient
}

// Load fetches a document. 401/403 map to ErrReadOnly, 404 to ErrNotFound.
func (s *HTTPStore) Load(ctx context.Context, ref string) (*Snapshot, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ref, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("scene: fetch %s: %w", ref, err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, fmt.Errorf("%w: %s", ErrNotFound, ref)
	case resp.StatusCode == http.StatusUnauthorized, resp.StatusCode == http.StatusForbidden:
		return nil, fmt.Errorf("%w: %s", ErrReadOnly, ref)
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("scene: fetch %s: %s", ref, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, MaxDocumentBytes+1))
	if err != nil {
		return nil, fmt.Errorf("scene: read %s: %w", ref, err)
	}
	if len(data) > MaxDocumentBytes {
		return nil, fmt.Errorf("scene: %s exceeds %d bytes", ref, MaxDocumentBytes)
	}
	version := resp.Header.Get("ETag")
	if version != "" {
		version = strings.Trim(version, `"`)
	} else {
		version = VersionOf(data)
	}
	return &Snapshot{Ref: ref, Version: version, Data: data, Writable: s.Writable}, nil
}

// Save PUTs a document with If-Match when a base version is known.
func (s *HTTPStore) Save(ctx context.Context, ref string, data []byte, baseVersion string) (*Snapshot, error) {
	if !s.Writable {
		return nil, fmt.Errorf("%w: %s", ErrReadOnly, ref)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, ref, strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if baseVersion != "" {
		req.Header.Set("If-Match", `"`+baseVersion+`"`)
	}
	resp, err := s.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("scene: save %s: %w", ref, err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusPreconditionFailed:
		return nil, fmt.Errorf("%w: %s changed on the server", ErrConflict, ref)
	case resp.StatusCode == http.StatusUnauthorized, resp.StatusCode == http.StatusForbidden:
		return nil, fmt.Errorf("%w: %s", ErrReadOnly, ref)
	case resp.StatusCode < 200 || resp.StatusCode > 299:
		return nil, fmt.Errorf("scene: save %s: %s", ref, resp.Status)
	}
	version := strings.Trim(resp.Header.Get("ETag"), `"`)
	if version == "" {
		version = VersionOf(data)
	}
	return &Snapshot{Ref: ref, Version: version, Data: data, Writable: true}, nil
}

func origin(u *url.URL) string {
	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)
}
