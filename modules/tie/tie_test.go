package tie

import (
	"context"
	"testing"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/uidbz/tie/client"
)

// Connect starts from the user's tie config file, so the connect options must
// win over whatever that file names. Getting this wrong is silent: writes land
// in the config's collection and still read back fine within the session.
func TestConnectHonorsNamespaceAndCollection(t *testing.T) {
	opts := object.NewMap(map[string]object.Object{
		"namespace":  object.NewString("atlas"),
		"collection": object.NewString("measurements"),
	})
	res, err := Connect(context.Background(), object.NewString("http://localhost:2161"), opts)
	if err != nil {
		t.Fatal(err)
	}
	info := res.(*Client).Interface().(*client.TieClient).CollectionInfo()
	if info.Namespace != "atlas" || info.CollectionId != "measurements" {
		t.Fatalf("connect options ignored: resolved %q/%q", info.Namespace, info.CollectionId)
	}
}

func TestConnectFilehostOption(t *testing.T) {
	opts := object.NewMap(map[string]object.Object{
		"filehost": object.NewMap(map[string]object.Object{
			"url":   object.NewString("http://localhost:2162"),
			"store": object.NewString("ssd"),
		}),
	})
	res, err := Connect(context.Background(), object.NewString("http://localhost:2161"), opts)
	if err != nil {
		t.Fatal(err)
	}
	cfg := res.(*Client).Interface().(*client.TieClient).Config
	host, ok := cfg.FileHosts["default"]
	if !ok {
		t.Fatalf("filehost option did not replace the default host: %+v", cfg.FileHosts)
	}
	if host.URL != "http://localhost:2162" || host.Store != "ssd" {
		t.Errorf("filehost decoded wrong: %+v", host)
	}
}

func TestConnectFilehostRequiresURL(t *testing.T) {
	opts := object.NewMap(map[string]object.Object{
		"filehost": object.NewMap(map[string]object.Object{
			"store": object.NewString("ssd"),
		}),
	})
	if _, err := Connect(context.Background(), object.NewString("http://localhost:2161"), opts); err == nil {
		t.Fatal("a filehost map without a url must be rejected")
	}
}
