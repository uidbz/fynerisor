package tieassets

import (
	"testing"

	renderbuiltin "github.com/uidbz/fynerisor/scene/render/builtin"
)

func TestRegisterTieComponents(t *testing.T) {
	r := renderbuiltin.NewRegistry()
	Register(r, nil)
	for _, typ := range []string{
		"tie.image", "tie.gallery", "tie.collection", "tie.tags",
		"tie.related", "tie.document", "tie.audio", "tie.video",
	} {
		if _, ok := r.Spec(typ); !ok {
			t.Errorf("missing %s", typ)
		}
	}
}
