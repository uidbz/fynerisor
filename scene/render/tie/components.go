package tieassets

import (
	"fmt"
	"image/color"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/uidbz/tie/client"

	"github.com/uidbz/fynerisor/gui/guithread"
	"github.com/uidbz/fynerisor/scene"
	"github.com/uidbz/fynerisor/scene/render"
)

// Register adds the tie component family to r. The host must also install a
// tie-aware AssetResolver (New(tc)) in render.Options.
func Register(r *render.Registry, tc *client.TieClient) {
	r.Register(scene.ComponentSpec{
		Type: "tie.image", Title: "tie image", Category: "tie", Icon: "image", Kind: scene.Leaf,
		DefaultPlace: scene.Place{W: 320, H: 240},
		Doc:          "An image from a tie collection by content hash.",
		Props: []scene.PropSpec{
			{Name: "hash", Title: "Hash", Group: "Content", Kind: scene.PropAsset},
			{Name: "caption", Title: "Caption", Group: "Content", Kind: scene.PropString},
			{Name: "fill", Title: "Fill", Group: "Look", Kind: scene.PropEnum, Default: "cover",
				Enum: []string{"stretch", "contain", "cover", "original"}},
		},
		Events: []scene.EventSpec{{Name: "tapped"}, {Name: "doubleTapped"}},
	}, func(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
		return newTieImage(bc, n), nil
	})

	r.Register(scene.ComponentSpec{
		Type: "tie.gallery", Title: "tie gallery", Category: "tie", Icon: "grid", Kind: scene.Leaf,
		DefaultPlace: scene.Place{W: 480, H: 360},
		Doc:          "A thumbnail grid for a tie tag query (include AND, exclude NONE).",
		Props: []scene.PropSpec{
			{Name: "tags", Title: "Tags", Group: "Data", Kind: scene.PropTags},
			{Name: "exclude", Title: "Exclude tags", Group: "Data", Kind: scene.PropTags},
			{Name: "scope", Title: "Scope", Group: "Data", Kind: scene.PropString, Help: "tie-type scope, e.g. image-file"},
			{Name: "tile", Title: "Tile size", Group: "Layout", Kind: scene.PropNumber, Default: 120.0, Min: 48, Max: 512, HasMin: true, HasMax: true},
			{Name: "limit", Title: "Limit", Group: "Data", Kind: scene.PropInt, Default: 60, Min: 1, Max: 1000, HasMin: true, HasMax: true},
		},
		Events: []scene.EventSpec{{Name: "itemTapped"}, {Name: "itemSelected"}},
	}, func(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
		return newGallery(bc, n, tc, false), nil
	})

	r.Register(scene.ComponentSpec{
		Type: "tie.collection", Title: "tie collection", Category: "tie", Icon: "folder", Kind: scene.Leaf,
		DefaultPlace: scene.Place{W: 480, H: 360},
		Doc:          "Members of a tie collection (reverse member-of), as a thumbnail grid.",
		Props: []scene.PropSpec{
			{Name: "uid", Title: "Collection UID", Group: "Content", Kind: scene.PropString},
			{Name: "tile", Title: "Tile size", Group: "Layout", Kind: scene.PropNumber, Default: 120.0, Min: 48, Max: 512, HasMin: true, HasMax: true},
			{Name: "limit", Title: "Limit", Group: "Data", Kind: scene.PropInt, Default: 60, Min: 1, Max: 1000, HasMin: true, HasMax: true},
		},
		Events: []scene.EventSpec{{Name: "itemTapped"}, {Name: "itemSelected"}},
	}, func(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
		return newGallery(bc, n, tc, true), nil
	})

	r.Register(scene.ComponentSpec{
		Type: "tie.tags", Title: "tie tags", Category: "tie", Icon: "tags", Kind: scene.Leaf,
		DefaultPlace: scene.Place{W: 280, H: 60},
		Doc:          "The tags of a tie subject as a wrapped label.",
		Props:        []scene.PropSpec{{Name: "subject", Title: "Subject", Group: "Content", Kind: scene.PropString}},
	}, func(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
		return newSubjectLabel(bc, n, tc, "tags"), nil
	})

	r.Register(scene.ComponentSpec{
		Type: "tie.related", Title: "tie relations", Category: "tie", Icon: "link", Kind: scene.Leaf,
		DefaultPlace: scene.Place{W: 320, H: 120},
		Doc:          "The relations of a tie subject, e.g. album, artist, rating, member-of.",
		Props:        []scene.PropSpec{{Name: "subject", Title: "Subject", Group: "Content", Kind: scene.PropString}},
	}, func(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
		return newSubjectLabel(bc, n, tc, "relations"), nil
	})

	r.Register(scene.ComponentSpec{
		Type: "tie.document", Title: "tie document", Category: "tie", Icon: "document", Kind: scene.Leaf,
		DefaultPlace: scene.Place{W: 320, H: 64},
		Doc:          "A file summary: filename, media type and size.",
		Props:        []scene.PropSpec{{Name: "subject", Title: "Subject", Group: "Content", Kind: scene.PropString}},
		Events:       []scene.EventSpec{{Name: "tapped"}},
	}, func(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
		return newSubjectLabel(bc, n, tc, "document"), nil
	})

	r.Register(scene.ComponentSpec{
		Type: "tie.audio", Title: "tie audio", Category: "tie", Icon: "audio", Kind: scene.Leaf,
		DefaultPlace: scene.Place{W: 300, H: 56},
		Doc:          "A play control for an audio subject; emits tapped with item=hash. The host opener starts playback.",
		Props: []scene.PropSpec{
			{Name: "subject", Title: "Subject", Group: "Content", Kind: scene.PropString},
			{Name: "title", Title: "Title", Group: "Content", Kind: scene.PropString},
			{Name: "artist", Title: "Artist", Group: "Content", Kind: scene.PropString},
		},
		Events: []scene.EventSpec{{Name: "tapped"}},
	}, func(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
		return newMediaButton(bc, n, tc, true), nil
	})

	r.Register(scene.ComponentSpec{
		Type: "tie.video", Title: "tie video", Category: "tie", Icon: "video", Kind: scene.Leaf,
		DefaultPlace: scene.Place{W: 320, H: 180},
		Doc:          "A poster/play card for a video subject; emits tapped with item=hash.",
		Props: []scene.PropSpec{
			{Name: "subject", Title: "Subject", Group: "Content", Kind: scene.PropString},
			{Name: "title", Title: "Title", Group: "Content", Kind: scene.PropString},
		},
		Events: []scene.EventSpec{{Name: "tapped"}},
	}, func(bc *render.BuildContext, n *scene.Node) (render.Live, error) {
		return newMediaButton(bc, n, tc, false), nil
	})
}

// tieSurface is the transparent event layer used by image/media cards.
type tieSurface struct {
	widget.BaseWidget
	bc *render.BuildContext
	id string
}

func newTieSurface(bc *render.BuildContext, id string) *tieSurface {
	t := &tieSurface{bc: bc, id: id}
	t.ExtendBaseWidget(t)
	return t
}
func (t *tieSurface) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.NRGBA{}))
}
func (t *tieSurface) Tapped(e *fyne.PointEvent) {
	t.bc.Emit(t.id, "tapped", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}
func (t *tieSurface) DoubleTapped(e *fyne.PointEvent) {
	t.bc.Emit(t.id, "doubleTapped", map[string]any{"x": e.Position.X, "y": e.Position.Y})
}
func (t *tieSurface) MouseIn(*desktop.MouseEvent) {}
func (t *tieSurface) MouseOut()                   {}

// ---------------------------------------------------------------- image

type tieImage struct {
	img   *canvas.Image
	stack *fyne.Container
	bc    *render.BuildContext
	gen   atomic.Int64
	ref   string
}

func newTieImage(bc *render.BuildContext, n *scene.Node) *tieImage {
	img := canvas.NewImageFromResource(nil)
	img.FillMode = render.FillMode(bc.Str(n, "fill", "cover"))
	l := &tieImage{img: img, bc: bc}
	l.stack = container.NewStack(img, newTieSurface(bc, n.ID))
	l.apply(n)
	return l
}

func (l *tieImage) apply(n *scene.Node) {
	l.img.FillMode = render.FillMode(l.bc.Str(n, "fill", "cover"))
	hash := strings.TrimPrefix(l.bc.Str(n, "hash", ""), "tie:")
	if hash == l.ref && hash != "" {
		return
	}
	l.ref = hash
	l.gen.Add(1)
	if hash == "" {
		l.img.Resource = nil
		l.img.Refresh()
		return
	}
	gen := l.gen.Load()
	go func() {
		res, err := l.bc.Assets.Open(l.bc.Context, "tie:"+hash)
		guithread.Do(func() {
			if l.gen.Load() != gen {
				return
			}
			if err != nil {
				l.bc.Statusf("tie.image %s: %v", hash, err)
				return
			}
			l.img.Resource = res
			l.img.Refresh()
		})
	}()
}

func (l *tieImage) Object() fyne.CanvasObject { return l.stack }
func (l *tieImage) Update(n *scene.Node, _ scene.Fields) error {
	l.apply(n)
	return nil
}
func (l *tieImage) Destroy() { l.gen.Add(1) }

// --------------------------------------------------------------- gallery

type galleryItem struct {
	hash string
	name string
}

type galleryLive struct {
	box        *fyne.Container
	gw         *widget.GridWrap
	msg        *widget.Label
	bc         *render.BuildContext
	tc         *client.TieClient
	collection bool
	tile       float32

	mu     sync.Mutex
	items  []galleryItem
	loaded map[*canvas.Image]string
	ref    string
	gen    atomic.Int64
}

func newGallery(bc *render.BuildContext, n *scene.Node, tc *client.TieClient, collection bool) *galleryLive {
	tile := float32(bc.Num(n, "tile", 120))
	if tile < 48 {
		tile = 48
	}
	l := &galleryLive{bc: bc, tc: tc, collection: collection, tile: tile, loaded: make(map[*canvas.Image]string)}
	l.gw = widget.NewGridWrap(
		func() int {
			l.mu.Lock()
			defer l.mu.Unlock()
			return len(l.items)
		},
		func() fyne.CanvasObject {
			img := canvas.NewImageFromResource(nil)
			img.FillMode = canvas.ImageFillCover
			return container.NewGridWrap(fyne.NewSize(l.tile, l.tile), img)
		},
		func(id widget.GridWrapItemID, obj fyne.CanvasObject) {
			l.updateItem(int(id), obj)
		},
	)
	l.msg = widget.NewLabel("")
	l.msg.Wrapping = fyne.TextWrapWord
	l.msg.Hide()
	l.gw.OnSelected = func(id widget.GridWrapItemID) {
		l.mu.Lock()
		var item galleryItem
		if int(id) >= 0 && int(id) < len(l.items) {
			item = l.items[id]
		}
		l.mu.Unlock()
		if item.hash != "" {
			bc.Emit(n.ID, "itemSelected", map[string]any{"item": item.hash, "value": item.name})
			bc.Emit(n.ID, "itemTapped", map[string]any{"item": item.hash, "value": item.name})
		}
	}
	l.box = container.NewStack(l.gw, l.msg)
	l.apply(n)
	return l
}

func (l *galleryLive) Object() fyne.CanvasObject { return l.box }

func (l *galleryLive) apply(n *scene.Node) {
	ref := strings.Join(l.bc.List(n, "tags"), "\x00") + "|" +
		strings.Join(l.bc.List(n, "exclude"), "\x00") + "|" +
		l.bc.Str(n, "scope", "") + "|" + l.bc.Str(n, "uid", "") + "|" +
		fmt.Sprint(l.bc.Num(n, "limit", 60))
	if ref == l.ref {
		return
	}
	l.ref = ref
	gen := l.gen.Add(1)
	limit := int(l.bc.Num(n, "limit", 60))
	include := l.bc.List(n, "tags")
	exclude := l.bc.List(n, "exclude")
	scope := l.bc.Str(n, "scope", "")
	uid := l.bc.Str(n, "uid", "")
	collection := l.collection
	go func() {
		var (
			items []galleryItem
			err   error
		)
		if collection {
			if uid != "" {
				var rows []client.Row
				rows, err = l.tc.CollectionMembers(uid)
				if err == nil {
					for i, row := range rows {
						if limit > 0 && i >= limit {
							break
						}
						items = append(items, galleryItem{hash: row.Key, name: client.RowFirst(row, "filename")})
					}
				}
			}
		} else {
			var rows []client.TaggedFile
			rows, _, err = l.tc.FilesWithTags(scope, include, exclude, 0, limit)
			if err == nil {
				for _, f := range rows {
					if f.IsDir || f.IsArchive {
						continue
					}
					items = append(items, galleryItem{hash: f.Hash, name: f.Filename})
				}
			}
		}
		guithread.Do(func() {
			if l.gen.Load() != gen {
				return
			}
			l.mu.Lock()
			l.items = items
			l.mu.Unlock()
			l.gw.Refresh()
			switch {
			case err != nil:
				l.msg.SetText("Gallery query failed: " + err.Error())
				l.msg.Show()
				l.bc.Statusf("tie.gallery: %v", err)
			case len(items) == 0:
				l.msg.SetText("No items match this query.")
				l.msg.Show()
			default:
				l.msg.Hide()
			}
		})
	}()
}

func (l *galleryLive) updateItem(id int, obj fyne.CanvasObject) {
	l.mu.Lock()
	if id < 0 || id >= len(l.items) {
		l.mu.Unlock()
		return
	}
	item := l.items[id]
	l.mu.Unlock()
	box, ok := obj.(*fyne.Container)
	if !ok || len(box.Objects) == 0 {
		return
	}
	img, ok := box.Objects[0].(*canvas.Image)
	if !ok {
		return
	}
	if l.loaded[img] == item.hash {
		return
	}
	l.loaded[img] = item.hash
	img.Resource = nil
	img.Refresh()
	gen := l.gen.Load()
	go func() {
		res, err := l.bc.Assets.Open(l.bc.Context, "tie:"+item.hash)
		guithread.Do(func() {
			// The cell may have been recycled to another item while loading.
			if l.gen.Load() != gen || l.loaded[img] != item.hash {
				return
			}
			if err == nil {
				img.Resource = res
				img.Refresh()
			}
		})
	}()
}

func (l *galleryLive) Update(n *scene.Node, _ scene.Fields) error {
	if tile := float32(l.bc.Num(n, "tile", 120)); tile != l.tile {
		return render.ErrRebuild
	}
	l.apply(n)
	return nil
}
func (l *galleryLive) Destroy() { l.gen.Add(1) }

// ---------------------------------------------------------------- subject

type subjectLive struct {
	label *widget.Label
	bc    *render.BuildContext
	tc    *client.TieClient
	mode  string
	ref   string
	gen   atomic.Int64
}

func newSubjectLabel(bc *render.BuildContext, n *scene.Node, tc *client.TieClient, mode string) *subjectLive {
	l := &subjectLive{label: widget.NewLabel(""), bc: bc, tc: tc, mode: mode}
	l.label.Wrapping = fyne.TextWrapWord
	l.apply(n)
	return l
}

func (l *subjectLive) Object() fyne.CanvasObject { return l.label }

func (l *subjectLive) apply(n *scene.Node) {
	subject := l.bc.Str(n, "subject", "")
	if subject == l.ref {
		return
	}
	l.ref = subject
	gen := l.gen.Add(1)
	if subject == "" {
		l.label.SetText("")
		return
	}
	mode := l.mode
	go func() {
		text := ""
		switch mode {
		case "tags":
			tags, err := client.GetTags(l.tc, subject)
			if err == nil {
				text = strings.Join(tags, " · ")
			}
		case "relations":
			row, err := l.tc.Get(subject)
			if err == nil {
				keys := make([]string, 0, len(row.Attributes))
				for k := range row.Attributes {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				var parts []string
				for _, k := range keys {
					parts = append(parts, fmt.Sprintf("%s: %s", k, strings.Join(row.Attributes[k], ", ")))
				}
				text = strings.Join(parts, "\n")
			}
		default: // document
			row, err := l.tc.Get(subject)
			if err == nil {
				name := client.RowFirst(row, "filename")
				size := client.RowFirst(row, "filesize")
				media := client.RowFirst(row, "media-type")
				text = fmt.Sprintf("%s\n%s · %s bytes", name, media, size)
			}
		}
		guithread.Do(func() {
			if l.gen.Load() != gen {
				return
			}
			l.label.SetText(text)
		})
	}()
}

func (l *subjectLive) Update(n *scene.Node, _ scene.Fields) error {
	l.apply(n)
	return nil
}
func (l *subjectLive) Destroy() { l.gen.Add(1) }

// ------------------------------------------------------------------ media

type mediaButton struct {
	box   *fyne.Container
	btn   *widget.Button
	bc    *render.BuildContext
	tc    *client.TieClient
	audio bool
	ref   string
}

func newMediaButton(bc *render.BuildContext, n *scene.Node, tc *client.TieClient, audio bool) *mediaButton {
	l := &mediaButton{bc: bc, tc: tc, audio: audio}
	l.btn = widget.NewButton("", func() {
		l.bc.Emit(n.ID, "tapped", map[string]any{"item": l.ref})
	})
	if audio {
		l.btn.Icon = theme.MediaPlayIcon()
		l.btn.Importance = widget.LowImportance
	} else {
		l.btn.Icon = theme.MediaVideoIcon()
		l.btn.Importance = widget.LowImportance
	}
	l.box = container.NewStack(l.btn)
	l.apply(n)
	return l
}

func (l *mediaButton) Object() fyne.CanvasObject { return l.box }

func (l *mediaButton) apply(n *scene.Node) {
	subject := l.bc.Str(n, "subject", "")
	l.ref = subject
	title := l.bc.Str(n, "title", "")
	if title == "" && subject != "" {
		title = subject
		if len(title) > 16 {
			title = title[:16] + "…"
		}
	}
	artist := l.bc.Str(n, "artist", "")
	if artist != "" {
		title += " — " + artist
	}
	if title == "" {
		title = "media"
	}
	l.btn.SetText(title)
}

func (l *mediaButton) Update(n *scene.Node, _ scene.Fields) error {
	l.apply(n)
	return nil
}
func (l *mediaButton) Destroy() {}
