// Package tiestore stores scene documents in a tie collection.
//
// A scene is a directory with a stable DirUID:
//
//	tie:/scenes/house/living-room/     (dir-type "scene" in a full app)
//	  scene.json                       content-addressed file, versions in _prev
//	  page.risor                       optional sidecar
//	  thumbnail.png                    optional preview
//
// The current document is addressed by a forward pointer on the directory,
//
//	(D, scene-doc, H)
//
// so loading is one Get (plus one blob download) and saving writes a new
// scene.json (tie versions the old one) and updates the pointer last. The
// DirUID is the stable identity for tags, member-of edges and links-to edges;
// H is the immutable version.
package tiestore

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/uidbz/tie/client"

	"github.com/uidbz/fynerisor/scene"
)

// OnIndexError receives best-effort indexing failures after a successful save.
type Store struct {
	Client       *client.TieClient
	OnIndexError func(error)
}

// SceneDocRelation is the directory → current document hash pointer.
const SceneDocRelation = "scene-doc"

// SceneScriptRelation is the optional sidecar script hash.
const SceneScriptRelation = "scene-script"

// SceneThumbnailRelation is the optional preview image hash.
const SceneThumbnailRelation = "thumbnail"

// New returns a store for tc.
func New(tc *client.TieClient) *Store { return &Store{Client: tc} }

// Resolve resolves a scene reference and requires a tie: directory path.
func (s *Store) Resolve(base, ref string) (string, error) {
	out, err := scene.ResolveRef(base, ref)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(out, "tie:") {
		if strings.HasPrefix(out, "/") {
			out = "tie:" + out
		} else {
			return "", fmt.Errorf("tiestore: %q is not a tie: path", out)
		}
	}
	u, err := url.Parse(out)
	if err != nil {
		return "", err
	}
	if u.Scheme != "tie" || !strings.HasPrefix(u.Path, "/") {
		return "", fmt.Errorf("tiestore: %q is not a tie directory path", out)
	}
	return out, nil
}

// Load reads the current document pointer and downloads its blob.
func (s *Store) Load(ctx context.Context, ref string) (*scene.Snapshot, error) {
	if s.Client == nil {
		return nil, errors.New("tiestore: no tie client")
	}
	path, err := s.Resolve("", ref)
	if err != nil {
		return nil, err
	}
	uid, err := s.Client.DirUIDFromPath(path)
	if err != nil {
		return nil, fmt.Errorf("tiestore: resolve %s: %w", path, err)
	}
	if uid == "" {
		return nil, fmt.Errorf("%w: %s", scene.ErrNotFound, path)
	}
	hash, err := s.documentHash(string(uid))
	if err != nil {
		return nil, err
	}
	if hash == "" {
		return nil, fmt.Errorf("%w: %s has no scene document", scene.ErrNotFound, path)
	}
	tmp, err := os.CreateTemp("", "fynerisor-scene-*.json")
	if err != nil {
		return nil, err
	}
	name := tmp.Name()
	tmp.Close()
	defer os.Remove(name)
	if err := s.Client.DownloadContext(ctx, "", hash, name); err != nil {
		return nil, fmt.Errorf("tiestore: download %s: %w", hash, err)
	}
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return &scene.Snapshot{Ref: path, Version: hash, Data: data, Writable: true}, nil
}

// Save writes a new scene.json and updates the directory pointer.
func (s *Store) Save(ctx context.Context, ref string, data []byte, baseVersion string) (*scene.Snapshot, error) {
	if s.Client == nil {
		return nil, errors.New("tiestore: no tie client")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path, err := s.Resolve("", ref)
	if err != nil {
		return nil, err
	}
	uid, err := s.Client.MkTieDirAll(path)
	if err != nil {
		return nil, fmt.Errorf("tiestore: create %s: %w", path, err)
	}
	if baseVersion != "" {
		cur, err := s.documentHash(string(uid))
		if err != nil {
			return nil, err
		}
		if cur != baseVersion {
			return nil, fmt.Errorf("%w: %s changed", scene.ErrConflict, path)
		}
	}
	host, err := s.Client.ResolveHost("")
	if err != nil {
		return nil, fmt.Errorf("tiestore: resolve filehost: %w", err)
	}
	tmp, err := os.CreateTemp("", "fynerisor-scene-*.json")
	if err != nil {
		return nil, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}
	hash, err := s.Client.WriteFile(host, client.DirUID(uid), "scene.json", tmpName, nil)
	if err != nil {
		return nil, fmt.Errorf("tiestore: write scene.json: %w", err)
	}
	batch := s.Client.NewBatch()
	batch.Set(string(uid), SceneDocRelation, []string{hash})
	if _, err := s.Client.BatchContext(ctx, batch); err != nil {
		return nil, fmt.Errorf("tiestore: update scene-doc: %w", err)
	}
	if err := s.Index(ctx, path, data); err != nil && s.OnIndexError != nil {
		s.OnIndexError(err)
	}
	return &scene.Snapshot{Ref: path, Version: hash, Data: data, Writable: true}, nil
}

// documentHash returns the pointer value, falling back to the scene.json child
// for directories created before the pointer existed.
func (s *Store) documentHash(uid string) (string, error) {
	row, err := s.Client.Get(uid)
	if err == nil {
		if h := client.RowFirst(row, SceneDocRelation); h != "" {
			return h, nil
		}
	}
	dir, err := client.ReadTieDir(s.Client, client.DirUID(uid))
	if err != nil {
		return "", nil // an empty/fresh directory is not an outage
	}
	for _, f := range dir.Files {
		if f.Filename == "scene.json" {
			return f.Uid, nil
		}
	}
	return "", nil
}

// ScenePath converts a tie scene directory URL into its client path form.
func ScenePath(ref string) (string, error) {
	if !strings.HasPrefix(ref, "tie:") {
		if strings.HasPrefix(ref, "/") {
			ref = "tie:" + ref
		} else {
			return "", fmt.Errorf("tiestore: %q is not a tie: path", ref)
		}
	}
	return ref, nil
}

// SaveThumbnail stores a preview PNG for the scene and points the directory's
// thumbnail relation at it. It is used by browsers after a successful save.
func (s *Store) SaveThumbnail(ctx context.Context, ref string, png []byte) error {
	if s.Client == nil {
		return errors.New("tiestore: no tie client")
	}
	path, err := s.Resolve("", ref)
	if err != nil {
		return err
	}
	uid, err := s.Client.DirUIDFromPath(path)
	if err != nil || uid == "" {
		return fmt.Errorf("tiestore: thumbnail: no scene directory %s", path)
	}
	host, err := s.Client.ResolveHost("")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp("", "fynerisor-thumb-*.png")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := tmp.Write(png); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	hash, err := s.Client.WriteFile(host, uid, "thumbnail.png", name, nil)
	if err != nil {
		return err
	}
	batch := s.Client.NewBatch()
	batch.Set(string(uid), SceneThumbnailRelation, []string{hash})
	_, err = s.Client.BatchContext(ctx, batch)
	return err
}
