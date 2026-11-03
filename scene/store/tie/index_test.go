package tiestore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/uidbz/tie/client"

	"github.com/uidbz/fynerisor/scene"
)

func TestIndexAndThumbnail(t *testing.T) {
	cfg := client.TestingConfig()
	cfg.Username = "defaultuser"
	cfg.Password = "defaultpassword"
	tc := client.NewTieClient(cfg)
	tc.SetTimeout(2 * time.Second)
	if _, err := tc.ServerVersion(); err != nil {
		t.Skipf("tie test environment not running: %v", err)
	}
	ctx := context.Background()
	store := New(tc)
	base := fmt.Sprintf("tie:/fynerisor-index-tests/%d", time.Now().UnixNano())
	target := base + "/target"
	if _, err := tc.MkTieDirAll(target); err != nil {
		t.Fatal(err)
	}
	media := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	doc := &scene.Document{
		Format: scene.Format, Title: "Index test",
		Root: &scene.Node{ID: "root", Type: "free", Children: []*scene.Node{
			{ID: "img", Type: "image", Props: map[string]any{"src": "tie:" + media}},
			{ID: "go", Type: "link", On: map[string][]scene.Action{
				"tapped": {{Do: "navigate", To: target}},
			}},
		}},
	}
	data, err := scene.Encode(doc)
	if err != nil {
		t.Fatal(err)
	}
	sceneRef := base + "/page"
	snap, err := store.Save(ctx, sceneRef, data, "")
	if err != nil {
		t.Fatal(err)
	}
	store.WaitIndex()
	uid, err := tc.DirUIDFromPath(sceneRef)
	if err != nil || uid == "" {
		t.Fatalf("uid = %q err=%v", uid, err)
	}
	row, err := tc.Get(string(uid))
	if err != nil {
		t.Fatal(err)
	}
	if got := client.RowFirst(row, "scene-doc"); got != snap.Version {
		t.Errorf("scene-doc = %q, want %q", got, snap.Version)
	}
	links := client.RowValues(row, "links-to")
	if len(links) != 1 {
		t.Fatalf("links-to = %v", links)
	}
	members, err := tc.CollectionMembers(string(uid))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range members {
		if m.Key == media {
			found = true
		}
	}
	if !found {
		t.Fatalf("media %s not indexed: %#v", media, members)
	}
	// A minimal PNG (same bytes as the tieassets test).
	png := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
	if err := store.SaveThumbnail(ctx, sceneRef, png); err != nil {
		t.Fatal(err)
	}
	row, err = tc.Get(string(uid))
	if err != nil {
		t.Fatal(err)
	}
	if client.RowFirst(row, SceneThumbnailRelation) == "" {
		t.Fatal("thumbnail relation not written")
	}
}
