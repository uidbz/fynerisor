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
	"sync"

	"github.com/uidbz/tie/api"
	"github.com/uidbz/tie/client"

	"github.com/uidbz/fynerisor/scene"
)

// OnIndexError receives best-effort indexing failures after a successful save.
type Store struct {
	Client Client
	// Schema, when set, tells the indexer which props are asset references.
	// Without it every 64-hex string in a prop is treated as media (which also
	// sweeps in collection and directory UIDs).
	Schema interface {
		Spec(string) (scene.ComponentSpec, bool)
	}
	OnIndexError func(error)

	thumbMu sync.Mutex
}

// Client is the subset of *client.TieClient the store uses. Keeping it narrow
// lets tests exercise the error paths (a transient triplestore failure must
// never look like a missing document) without a live server.
type Client interface {
	DirUIDFromPath(path string) (client.DirUID, error)
	MkTieDirAll(path string) (client.DirUID, error)
	GetContext(ctx context.Context, key string) (client.Row, error)
	QueryContext(ctx context.Context, spec client.QuerySpec) ([]client.Row, int, error)
	DownloadContext(ctx context.Context, hostName, hash, dest string) error
	ResolveHost(name string) (client.FileHost, error)
	WriteFile(host client.FileHost, parent client.DirUID, name, srcPath string, extraTags []string) (string, error)
	NewBatch() *api.Batch
	BatchContext(ctx context.Context, batch *api.Batch) (client.BatchReply, error)
	CollectionMembers(uid string) ([]client.Row, error)
	SetDirType(uid client.DirUID, dirType string) error
	Sync() error
}

// SceneDocRelation is the directory → current document hash pointer.
const SceneDocRelation = "scene-doc"

// SceneScriptRelation is the optional sidecar script hash.
const SceneScriptRelation = "scene-script"

// SceneThumbnailRelation is the optional preview image hash.
const SceneThumbnailRelation = "thumbnail"

// DefaultScriptName is the sidecar name used when a scene has no explicit one.
const DefaultScriptName = "page.risor"

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
	hash, err := s.documentHash(ctx, string(uid))
	if err != nil {
		return nil, err
	}
	if hash == "" {
		return nil, fmt.Errorf("%w: %s has no scene document", scene.ErrNotFound, path)
	}
	data, err := s.download(ctx, hash)
	if err != nil {
		return nil, err
	}
	return &scene.Snapshot{Ref: path, Version: hash, Data: data, Writable: true}, nil
}

// LoadScript loads the optional sidecar script for a scene. A nil slice means
// the scene has no script. The reference is resolved against the scene
// directory, so "page.risor" lives inside it (a sibling resolution put it in
// the parent directory, where Load then treated it as another scene).
func (s *Store) LoadScript(ctx context.Context, ref, name string) ([]byte, string, error) {
	if s.Client == nil {
		return nil, "", errors.New("tiestore: no tie client")
	}
	path, err := s.Resolve("", ref)
	if err != nil {
		return nil, "", err
	}
	uid, err := s.Client.DirUIDFromPath(path)
	if err != nil {
		return nil, "", fmt.Errorf("tiestore: resolve %s: %w", path, err)
	}
	if uid == "" {
		return nil, "", nil
	}
	row, err := s.Client.GetContext(ctx, string(uid))
	if errors.Is(err, client.ErrNotFound) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("tiestore: read %s: %w", path, err)
	}
	hash := client.RowFirst(row, SceneScriptRelation)
	if hash == "" {
		if name == "" {
			name = DefaultScriptName
		}
		hash, err = s.childFileHash(ctx, string(uid), name)
		if err != nil {
			return nil, "", err
		}
	}
	if hash == "" {
		return nil, "", nil
	}
	data, err := s.download(ctx, hash)
	if err != nil {
		return nil, "", err
	}
	return data, hash, nil
}

// SaveScript writes the sidecar script into the scene directory and updates the
// scene-script pointer.
func (s *Store) SaveScript(ctx context.Context, ref, name string, data []byte) error {
	if s.Client == nil {
		return errors.New("tiestore: no tie client")
	}
	path, err := s.Resolve("", ref)
	if err != nil {
		return err
	}
	uid, err := s.Client.MkTieDirAll(path)
	if err != nil {
		return fmt.Errorf("tiestore: create %s: %w", path, err)
	}
	if name == "" {
		name = DefaultScriptName
	}
	host, err := s.Client.ResolveHost("")
	if err != nil {
		return fmt.Errorf("tiestore: resolve filehost: %w", err)
	}
	tmp, err := os.CreateTemp("", "fynerisor-script-*.risor")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	hash, err := s.Client.WriteFile(host, uid, name, tmpName, nil)
	if err != nil {
		return fmt.Errorf("tiestore: write %s: %w", name, err)
	}
	batch := s.Client.NewBatch()
	batch.Set(string(uid), SceneScriptRelation, []string{hash})
	if _, err := s.Client.BatchContext(ctx, batch); err != nil {
		return fmt.Errorf("tiestore: update scene-script: %w", err)
	}
	return nil
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
		cur, err := s.documentHash(ctx, string(uid))
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
	hash, err := s.Client.WriteFile(host, uid, "scene.json", tmpName, nil)
	if err != nil {
		return nil, fmt.Errorf("tiestore: write scene.json: %w", err)
	}
	batch := s.Client.NewBatch()
	batch.Set(string(uid), SceneDocRelation, []string{hash})
	if _, err := s.Client.BatchContext(ctx, batch); err != nil {
		return nil, fmt.Errorf("tiestore: update scene-doc: %w", err)
	}
	// Reads are read-your-writes; Sync only bounds what a crash can lose.
	_ = s.Client.Sync()
	if err := s.Index(ctx, path, data); err != nil && s.OnIndexError != nil {
		s.OnIndexError(err)
	}
	return &scene.Snapshot{Ref: path, Version: hash, Data: data, Writable: true}, nil
}

// download fetches one blob into memory.
func (s *Store) download(ctx context.Context, hash string) ([]byte, error) {
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
	return os.ReadFile(name)
}

// documentHash returns the pointer value. A missing pointer means "this
// directory is not a scene yet"; a transport error is returned as such.
//
// The old fallback looked for a scene.json child, but tie classifies JSON as
// unknown-file, which ReadTieDir does not list — and, worse, swallowing the
// Get error made a network blip look like a missing document (a spurious save
// conflict, or a fresh starter page overwriting the home scene).
func (s *Store) documentHash(ctx context.Context, uid string) (string, error) {
	row, err := s.Client.GetContext(ctx, uid)
	if errors.Is(err, client.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("tiestore: read %s: %w", uid, err)
	}
	return client.RowFirst(row, SceneDocRelation), nil
}

// childFileHash finds a child of uid by filename. Used only for the sidecar
// fallback, so a scene authored without a scene-script pointer still runs.
func (s *Store) childFileHash(ctx context.Context, uid, name string) (string, error) {
	rows, _, err := s.Client.QueryContext(ctx, client.QuerySpec{
		Terms: []string{uid}, Reverse: true, Expand: true,
	})
	if errors.Is(err, client.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("tiestore: list %s: %w", uid, err)
	}
	for _, row := range rows {
		if client.RowFirst(row, "filename") == name {
			return row.Key, nil
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
// Thumbnails are serialized: two saves finishing together must not race on the
// same thumbnail.png (and its version chain).
func (s *Store) SaveThumbnail(ctx context.Context, ref string, png []byte) error {
	if s.Client == nil {
		return errors.New("tiestore: no tie client")
	}
	s.thumbMu.Lock()
	defer s.thumbMu.Unlock()
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
