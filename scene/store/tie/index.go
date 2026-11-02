package tiestore

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/uidbz/fynerisor/scene"
)

var hashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Index records the scene's current document, media references and links on
// the directory subject. It is best-effort: Save has already committed the
// document when it runs, so indexing failures are reported through
// Store.OnIndexError instead of failing the save.
func (s *Store) Index(ctx context.Context, ref string, data []byte) error {
	if s.Client == nil {
		return nil
	}
	path, err := s.Resolve("", ref)
	if err != nil {
		return err
	}
	uid, err := s.Client.DirUIDFromPath(path)
	if err != nil || uid == "" {
		return fmt.Errorf("tiestore: index %s: directory not found", path)
	}
	doc, err := scene.Decode(data)
	if err != nil {
		return err
	}
	desiredMembers := mediaRefs(doc)
	currentRows, err := s.Client.CollectionMembers(string(uid))
	if err != nil {
		currentRows = nil
	}
	current := make(map[string]bool, len(currentRows))
	for _, row := range currentRows {
		current[row.Key] = true
	}
	batch := s.Client.NewBatch()
	ops := 0
	for hash := range desiredMembers {
		if !current[hash] {
			batch.Add(hash, "member-of", string(uid))
			ops++
		}
	}
	for hash := range current {
		if !desiredMembers[hash] {
			batch.Delete(hash, "member-of", string(uid))
			ops++
		}
	}
	links := sceneLinks(doc, path)
	targets := make([]string, 0, len(links))
	for target := range links {
		if targetUID, err := s.Client.DirUIDFromPath(target); err == nil && targetUID != "" && targetUID != uid {
			targets = append(targets, string(targetUID))
		}
	}
	batch.Set(string(uid), "links-to", targets)
	ops++
	if _, err := s.Client.BatchContext(ctx, batch); err != nil {
		return err
	}
	// Mark the directory as a scene for consumers that filter by dir-type.
	_ = s.Client.SetDirType(uid, "scene")
	return nil
}

// mediaRefs collects content hashes referenced by props or action values.
func mediaRefs(doc *scene.Document) map[string]bool {
	out := make(map[string]bool)
	add := func(s string) {
		s = strings.TrimPrefix(strings.TrimSpace(s), "tie:")
		if hashPattern.MatchString(s) {
			out[s] = true
		}
	}
	var walkValue func(v any)
	walkValue = func(v any) {
		switch x := v.(type) {
		case string:
			add(x)
		case []any:
			for _, e := range x {
				walkValue(e)
			}
		case map[string]any:
			for _, e := range x {
				walkValue(e)
			}
		}
	}
	doc.Walk(func(n *scene.Node) bool {
		for _, v := range n.Props {
			walkValue(v)
		}
		for _, acts := range n.On {
			for _, a := range acts {
				if a.Args != nil {
					walkValue(a.Args)
				}
				if a.Value != nil {
					walkValue(a.Value)
				}
			}
		}
		return true
	})
	return out
}

// sceneLinks resolves navigate targets to tie scene paths.
func sceneLinks(doc *scene.Document, base string) map[string]bool {
	out := make(map[string]bool)
	doc.Walk(func(n *scene.Node) bool {
		for _, acts := range n.On {
			for _, a := range acts {
				if a.Do != "navigate" || a.To == "" {
					continue
				}
				target, err := scene.ResolveRef(base, a.To)
				if err != nil {
					continue
				}
				if strings.HasPrefix(target, "tie:") {
					out[target] = true
				}
			}
		}
		return true
	})
	return out
}
