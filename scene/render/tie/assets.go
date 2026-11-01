// Package tieassets resolves scene asset references through a tie collection.
//
// It is the bridge between `tie:<hash>` scene references and the renderer:
// blobs are downloaded through the filehost, optionally preferring the
// server-cached thumbnail that tie-view/tie-fm write for an item.
package tieassets

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"github.com/uidbz/tie/client"

	"github.com/uidbz/fynerisor/scene"
)

// Assets implements render.AssetResolver for tie references.
type Assets struct {
	Client *client.TieClient
	// PreferThumbnails asks the item's `thumbnail` relation first. The
	// thumbnail is usually much smaller than a camera original.
	PreferThumbnails bool

	mu    sync.Mutex
	cache map[string]fyne.Resource
}

// New returns a tie asset resolver.
func New(tc *client.TieClient) *Assets {
	return &Assets{Client: tc, PreferThumbnails: true, cache: make(map[string]fyne.Resource)}
}

// Resolve resolves an asset reference against the document base.
func (a *Assets) Resolve(base, ref string) (string, error) {
	return scene.ResolveRef(base, ref)
}

// Open downloads a tie: blob (or its thumbnail) and returns it as a resource.
// It performs network I/O and must run off the UI thread.
func (a *Assets) Open(ctx context.Context, ref string) (fyne.Resource, error) {
	if a.Client == nil {
		return nil, fmt.Errorf("tieassets: no tie client")
	}
	if !strings.HasPrefix(ref, "tie:") {
		return nil, fmt.Errorf("tieassets: %q is not a tie: reference", ref)
	}
	hash := strings.TrimPrefix(ref, "tie:")
	if hash == "" {
		return nil, fmt.Errorf("tieassets: empty hash")
	}
	a.mu.Lock()
	if r, ok := a.cache[hash]; ok {
		a.mu.Unlock()
		return r, nil
	}
	a.mu.Unlock()

	useThumb := ""
	if a.PreferThumbnails {
		if row, err := a.Client.GetContext(ctx, hash); err == nil {
			useThumb = client.RowFirst(row, "thumbnail")
		}
	}
	downloadHash := hash
	if useThumb != "" {
		downloadHash = useThumb
	}
	tmp, err := os.CreateTemp("", "fynerisor-asset-*")
	if err != nil {
		return nil, err
	}
	name := tmp.Name()
	tmp.Close()
	defer os.Remove(name)
	if err := a.Client.DownloadContext(ctx, "", downloadHash, name); err != nil {
		// A stale thumbnail must not hide the original.
		if useThumb != "" {
			if err := a.Client.DownloadContext(ctx, "", hash, name); err == nil {
				downloadHash = hash
			} else {
				return nil, fmt.Errorf("tieassets: download %s: %w", hash, err)
			}
		} else {
			return nil, fmt.Errorf("tieassets: download %s: %w", hash, err)
		}
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	res := fyne.NewStaticResource(downloadHash[:minInt(8, len(downloadHash))]+".bin", data)
	a.mu.Lock()
	a.cache[hash] = res
	a.mu.Unlock()
	return res, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
