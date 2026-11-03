package tieassets

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/uidbz/tie/client"
)

// Integration test: skipped unless the tie test environment is running.
func TestOpenBlob(t *testing.T) {
	cfg := client.TestingConfig()
	cfg.Username = "defaultuser"
	cfg.Password = "defaultpassword"
	tc := client.NewTieClient(cfg)
	tc.SetTimeout(2 * time.Second)
	if _, err := tc.ServerVersion(); err != nil {
		t.Skipf("tie test environment not running: %v", err)
	}
	// A tiny valid PNG.
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
	dir := t.TempDir()
	path := filepath.Join(dir, "pixel.png")
	if err := os.WriteFile(path, png, 0o644); err != nil {
		t.Fatal(err)
	}
	up, err := tc.UploadContext(context.Background(), "", path)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if len(up.Items) == 0 {
		t.Fatal("no uploaded item")
	}
	a := New(tc)
	a.PreferThumbnails = false
	res, err := a.Open(context.Background(), "tie:"+up.Items[0].Hash)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if !bytes.Equal(res.Content(), png) {
		t.Fatalf("content differs (%d bytes)", len(res.Content()))
	}
}
