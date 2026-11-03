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

	mu       sync.Mutex
	cache    map[string]fyne.Resource
	order    []string
	max      int
	inflight map[string]*assetCall
}

// assetCall deduplicates concurrent loads of the same hash (a gallery and a
// picker can ask for the same thumbnail at the same time).
type assetCall struct {
	done chan struct{}
	res  fyne.Resource
	err  error
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
	if call, ok := a.inflight[hash]; ok {
		a.mu.Unlock()
		select {
		case <-call.done:
			return call.res, call.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if a.inflight == nil {
		a.inflight = make(map[string]*assetCall)
	}
	call := &assetCall{done: make(chan struct{})}
	a.inflight[hash] = call
	a.mu.Unlock()

	res, err := a.load(ctx, hash)
	a.mu.Lock()
	call.res, call.err = res, err
	delete(a.inflight, hash)
	a.mu.Unlock()
	close(call.done)
	return res, err
}

// load fetches one asset (thumbnail when available) and caches it.
func (a *Assets) load(ctx context.Context, hash string) (fyne.Resource, error) {
	useThumb := ""
	name := hash + ".bin"
	// One Get gives both the thumbnail pointer and a real filename. The name
	// matters: Fyne picks SVG rendering from the suffix, and a ".bin" name
	// made SVG assets decode as bitmaps.
	if row, err := a.Client.GetContext(ctx, hash); err == nil {
		if a.PreferThumbnails {
			useThumb = client.RowFirst(row, "thumbnail")
		}
		if fn := client.RowFirst(row, "filename"); fn != "" {
			name = fn
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
	tmpName := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpName)
	if err := a.Client.DownloadContext(ctx, "", downloadHash, tmpName); err != nil {
		// A stale thumbnail must not hide the original.
		if useThumb != "" {
			if err := a.Client.DownloadContext(ctx, "", hash, tmpName); err == nil {
				downloadHash = hash
			} else {
				return nil, fmt.Errorf("tieassets: download %s: %w", hash, err)
			}
		} else {
			return nil, fmt.Errorf("tieassets: download %s: %w", hash, err)
		}
	}
	data, err := os.ReadFile(tmpName)
	if err != nil {
		return nil, err
	}
	res := fyne.NewStaticResource(name, data)
	a.mu.Lock()
	if a.cache == nil {
		a.cache = make(map[string]fyne.Resource)
	}
	if _, ok := a.cache[hash]; !ok {
		a.order = append(a.order, hash)
	}
	a.cache[hash] = res
	if a.max <= 0 {
		a.max = 256
	}
	for len(a.order) > a.max {
		old := a.order[0]
		a.order = a.order[1:]
		delete(a.cache, old)
	}
	a.mu.Unlock()
	return res, nil
}

// Prefetch warms the cache for refs (a scene gallery after a listing). It
// downloads at most four assets at a time and never blocks the caller.
func (a *Assets) Prefetch(ctx context.Context, refs []string) {
	if a.Client == nil || len(refs) == 0 {
		return
	}
	sem := make(chan struct{}, 4)
	for _, ref := range refs {
		ref := ref
		go func() {
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()
			_, _ = a.Open(ctx, ref)
		}()
	}
}
