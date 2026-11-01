package render

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"

	"github.com/uidbz/fynerisor/scene"
)

// LocalAssets loads file: and http(s) assets with a per-instance cache. tie:
// references need a host resolver (tie-browser supplies one that streams from
// the filehost and serves thumbnails).
type LocalAssets struct {
	Client   *http.Client
	MaxBytes int64

	mu    sync.Mutex
	cache map[string]fyne.Resource
}

// NewLocalAssets returns a resolver with a 30-second HTTP timeout.
func NewLocalAssets() *LocalAssets {
	return &LocalAssets{
		Client:   &http.Client{Timeout: 30 * time.Second},
		MaxBytes: 64 << 20,
		cache:    make(map[string]fyne.Resource),
	}
}

// Resolve resolves an asset reference against the document base.
func (a *LocalAssets) Resolve(base, ref string) (string, error) {
	return scene.ResolveRef(base, ref)
}

func (a *LocalAssets) client() *http.Client {
	if a.Client != nil {
		return a.Client
	}
	return http.DefaultClient
}

// Open loads an asset. It performs I/O and must run off the UI thread.
func (a *LocalAssets) Open(ctx context.Context, ref string) (fyne.Resource, error) {
	a.mu.Lock()
	if r, ok := a.cache[ref]; ok {
		a.mu.Unlock()
		return r, nil
	}
	a.mu.Unlock()

	u, err := url.Parse(ref)
	if err != nil {
		return nil, fmt.Errorf("scene/render: bad asset reference %q: %w", ref, err)
	}
	var data []byte
	switch strings.ToLower(u.Scheme) {
	case "", "file":
		path := ref
		if u.Scheme == "file" {
			path = filepath.FromSlash(u.Path)
		}
		data, err = os.ReadFile(path)
	case "http", "https":
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, ref, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Set("User-Agent", "fynerisor-scene/1")
		resp, reqErr := a.client().Do(req)
		if reqErr != nil {
			return nil, fmt.Errorf("scene/render: fetch %s: %w", ref, reqErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("scene/render: fetch %s: %s", ref, resp.Status)
		}
		limit := a.MaxBytes
		if limit <= 0 {
			limit = 64 << 20
		}
		data, err = io.ReadAll(io.LimitReader(resp.Body, limit+1))
		if int64(len(data)) > limit {
			return nil, fmt.Errorf("scene/render: asset %s exceeds %d bytes", ref, limit)
		}
	case "tie":
		return nil, fmt.Errorf("scene/render: tie asset %q needs a tie asset resolver", ref)
	default:
		return nil, fmt.Errorf("scene/render: unsupported asset scheme %q", u.Scheme)
	}
	if err != nil {
		return nil, fmt.Errorf("scene/render: load %s: %w", ref, err)
	}
	res := fyne.NewStaticResource(filepath.Base(u.Path), data)
	a.mu.Lock()
	if a.cache == nil {
		a.cache = make(map[string]fyne.Resource)
	}
	a.cache[ref] = res
	a.mu.Unlock()
	return res, nil
}

// MemoryAssets is a simple resolver for tests and embedded pages.
type MemoryAssets struct {
	mu   sync.Mutex
	data map[string][]byte
	Base string
}

// NewMemoryAssets returns an empty resolver.
func NewMemoryAssets() *MemoryAssets {
	return &MemoryAssets{data: make(map[string][]byte)}
}

// Put registers an asset.
func (m *MemoryAssets) Put(ref string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[ref] = data
}

// Resolve resolves an asset reference.
func (m *MemoryAssets) Resolve(base, ref string) (string, error) {
	if base == "" {
		base = m.Base
	}
	return scene.ResolveRef(base, ref)
}

// Open returns a registered asset.
func (m *MemoryAssets) Open(_ context.Context, ref string) (fyne.Resource, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.data[ref]
	if !ok {
		return nil, fmt.Errorf("scene/render: asset %q not found", ref)
	}
	return fyne.NewStaticResource(filepath.Base(ref), data), nil
}
