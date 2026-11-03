package render

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/uidbz/fynerisor/scene"
)

// StoreResolver loads embedded scenes from a store. With a nil Store it picks
// FileStore or HTTPStore from the reference scheme, which covers scene files
// and server-hosted pages. tie-browser supplies a tie store.
type StoreResolver struct {
	Store scene.Store
}

// LoadEmbed implements EmbedResolver.
func (r *StoreResolver) LoadEmbed(ctx context.Context, base, ref string) (*scene.Document, string, error) {
	resolved, err := scene.ResolveRef(base, ref)
	if err != nil {
		return nil, "", err
	}
	if bu, err := url.Parse(base); err == nil && (bu.Scheme == "http" || bu.Scheme == "https") {
		ru, err := url.Parse(resolved)
		if err != nil || (ru.Scheme != "http" && ru.Scheme != "https") {
			return nil, "", fmt.Errorf("embed: remote page cannot load %q", resolved)
		}
		if !strings.EqualFold(bu.Host, ru.Host) {
			return nil, "", fmt.Errorf("embed: cross-origin reference %q", resolved)
		}
	}
	store := r.Store
	if store == nil {
		u, err := url.Parse(resolved)
		if err != nil {
			return nil, "", err
		}
		switch u.Scheme {
		case "http", "https":
			store = &scene.HTTPStore{}
		default:
			store = scene.FileStore{}
		}
	}
	snap, err := store.Load(ctx, resolved)
	if err != nil {
		return nil, "", err
	}
	doc, err := scene.Decode(snap.Data)
	if err != nil {
		return nil, "", fmt.Errorf("embed %s: %w", resolved, err)
	}
	return doc, resolved, nil
}
