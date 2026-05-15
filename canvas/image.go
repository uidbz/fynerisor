package risorcanvas

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/storage"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

var _ object.Object = &Image{}

const ImageType object.Type = "canvas.image"

type Image struct {
	instance  *canvas.Image
	container *fyne.Container
}

func (obj *Image) CanvasObject() fyne.CanvasObject {
	return obj.container
}

func (r *Image) Type() object.Type {
	return ImageType
}

func (r *Image) Inspect() string {
	return "canvas.image"
}

func (r *Image) Interface() interface{} {
	return r.container
}

func (r *Image) IsTruthy() bool {
	return true
}

func (r *Image) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal 'canvas.image'")
}

func (r *Image) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	err := errors.New("eval error: unsupported operation for " + string(ImageType))
	errObj := object.Errorf("eval error: unsupported operation for %s: %v", ImageType, opType)
	return errObj, err
}

func (r *Image) Equals(other object.Object) bool {
	return r == other
}

func (r *Image) Attrs() []object.AttrSpec {
	return nil
}

func (r *Image) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: %s object has no attribute %q", ImageType, name)
}

func (g *Image) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "SetImageFromURI":
		return object.NewBuiltin("canvas.Image.SetImageFromURI", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 1 {
				return object.Errorf("wrong number of arguments. got=%d, want=1", len(args)), nil
			}
			path, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}

			uri, err := storage.ParseURI(path)
			if err != nil {
				return object.NewError(err), nil
			}

			fyne.Do(func() {
				g.instance = canvas.NewImageFromURI(uri)
				g.instance.FillMode = canvas.ImageFillOriginal
				g.container.Objects = []fyne.CanvasObject{g.instance}
				g.container.Refresh()
			})

			return object.Nil, nil
		}), true
	}

	return nil, false
}

func NewImage(instance *canvas.Image) *Image {
	return &Image{instance: instance, container: container.NewStack(instance)}
}

// NewImageFromURI creates a new Image from a URI string.
// For HTTP/HTTPS URLs, it downloads the image with proper User-Agent header.
func NewImageFromURI(uri string) (*Image, error) {
	// For HTTP/HTTPS URIs, download and create from resource
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		img, err := loadImageFromHTTP(uri)
		if err != nil {
			return nil, err
		}
		return NewImage(img), nil
	}

	// For file:// and other URIs, use Fyne's built-in support
	fyneURI, err := storage.ParseURI(uri)
	if err != nil {
		return nil, err
	}

	img := canvas.NewImageFromURI(fyneURI)
	img.FillMode = canvas.ImageFillOriginal

	return NewImage(img), nil
}

// loadImageFromHTTP downloads an image from HTTP/HTTPS and creates a canvas.Image.
// This workaround is needed because Fyne's HTTPRepository doesn't set a User-Agent header,
// causing HTTP 403 errors on sites like Wikipedia that require it.
func loadImageFromHTTP(url string) (*canvas.Image, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent to avoid 403 errors from sites that block default Go client
	req.Header.Set("User-Agent", "fynerisor/0.4 (Go-http-client)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	img := canvas.NewImageFromResource(&staticResource{
		name:    "downloaded",
		content: data,
	})
	img.FillMode = canvas.ImageFillOriginal
	return img, nil
}

// staticResource implements fyne.Resource for in-memory data
type staticResource struct {
	name    string
	content []byte
}

func (r *staticResource) Name() string {
	return r.name
}

func (r *staticResource) Content() []byte {
	return r.content
}
