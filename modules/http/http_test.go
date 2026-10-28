package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

func TestRequestHonoursContext(t *testing.T) {
	// Issue #25: the request must carry the builtin's context, so
	// stopping a script aborts an in-flight request.
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
		}
	}))
	defer slow.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := get(ctx, object.NewString(slow.URL))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestBinaryBodyAndJSONInts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/bin" {
			w.Write([]byte{0, 1, 2, 0xff, 'x'})
			return
		}
		w.Write([]byte(`{"id": 3}`))
	}))
	defer srv.Close()

	resp, err := get(context.Background(), object.NewString(srv.URL+"/bin"))
	if err != nil {
		t.Fatal(err)
	}
	m := resp.(*object.Map)
	bb, ok := m.Get("body_bytes").(*object.Bytes)
	if !ok || string(bb.Value()) != string([]byte{0, 1, 2, 0xff, 'x'}) {
		t.Errorf("body_bytes = %v, want exact bytes", bb)
	}

	resp, err = get(context.Background(), object.NewString(srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	jsonFn := resp.(*object.Map).Get("json").(*object.Builtin)
	parsed, err := jsonFn.Value()(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if id, ok := parsed.(*object.Map).Get("id").(*object.Int); !ok || id.Value() != 3 {
		t.Errorf("json id = %v, want int 3", parsed.(*object.Map).Get("id"))
	}
}

func TestAuthAndHeadersOptions(t *testing.T) {
	var gotAuth, gotX string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotX = r.Header.Get("X-Test")
	}))
	defer srv.Close()

	options := object.NewMap(map[string]object.Object{
		"auth": object.NewMap(map[string]object.Object{
			"username": object.NewString("u"),
			"password": object.NewString("p"),
		}),
		"headers": object.NewMap(map[string]object.Object{
			"X-Test": object.NewString("yes"),
		}),
	})
	if _, err := get(context.Background(), object.NewString(srv.URL), options); err != nil {
		t.Fatal(err)
	}
	if gotX != "yes" {
		t.Errorf("X-Test = %q", gotX)
	}
	// base64("u:p")
	if !strings.HasPrefix(gotAuth, "Basic ") {
		t.Errorf("Authorization = %q", gotAuth)
	}
}

func TestFetchMethodOverride(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			got = r.Method
		}
	}))
	defer srv.Close()

	options := object.NewMap(map[string]object.Object{
		"method": object.NewString("post"),
	})
	if _, err := fetch(context.Background(), object.NewString(srv.URL), options); err != nil {
		t.Fatal(err)
	}
	if got != http.MethodPost {
		t.Errorf("fetch did not use POST")
	}

	// get must not honour an options method override.
	got = ""
	options2 := object.NewMap(map[string]object.Object{
		"method": object.NewString("post"),
	})
	if _, err := get(context.Background(), object.NewString(srv.URL), options2); err != nil {
		t.Fatal(err)
	}
	if got == http.MethodPost {
		t.Errorf("http.get allowed a method override")
	}
}

func TestResponseSizeLimit(t *testing.T) {
	orig := MaxResponseBytes
	MaxResponseBytes = 16
	defer func() { MaxResponseBytes = orig }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strings.Repeat("x", 1024)))
	}))
	defer srv.Close()

	_, err := get(context.Background(), object.NewString(srv.URL))
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected size-limit error, got %v", err)
	}
}
