// Package http provides HTTP client functionality for Risor scripts.
//
// # Authentication
//
// All HTTP methods (get, post, put, delete, fetch) support basic authentication
// via the `auth` parameter in the options map:
//
//	http.get("https://api.example.com/data", {
//		auth: {username: "user", password: "pass"}
//	})
//
//	http.post("https://api.example.com/submit", {
//		auth: {username: "user", password: "pass"},
//		headers: {"Content-Type": "application/json"}
//	}, body)
//
// The auth credentials are automatically base64-encoded and added as an
// Authorization: Basic header. This approach is more intuitive than manually
// constructing the header and doesn't require a separate base64 module.
//
// # Requests and responses
//
// Every request carries the builtin's context (issue #25): stopping the
// script aborts an in-flight request. All requests share one client (and
// one connection pool); the 30s timeout is per request. Responses are
// maps: status, statusText, ok, headers, body (text) and body_bytes
// (exact bytes, safe for binary bodies); `json` is a method that parses
// the body (integral numbers decode to ints, via internal/convert).
// Responses larger than MaxResponseBytes are refused.
package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"

	"github.com/uidbz/fynerisor/internal/convert"
)

// MaxResponseBytes bounds a single response body: scripts fetch APIs,
// they do not mirror servers with this module (issue #25). It is a var
// so tests can lower it.
var MaxResponseBytes = 32 << 20 // 32 MiB

// requestTimeout applies per request, as the old per-request client's did.
const requestTimeout = 30 * time.Second

// client is shared by every request (issue #25): one connection pool and
// one keep-alive state instead of a fresh http.Client per call.
var client = &http.Client{
	Timeout:   requestTimeout,
	Transport: &http.Transport{Proxy: http.ProxyFromEnvironment},
}

// Module returns the http module for risor
func Module() *object.Module {
	return object.NewBuiltinsModule("http", map[string]object.Object{
		"get":    object.NewBuiltin("http.get", get),
		"post":   object.NewBuiltin("http.post", post),
		"put":    object.NewBuiltin("http.put", put),
		"delete": object.NewBuiltin("http.delete", del),
		"fetch":  object.NewBuiltin("http.fetch", fetch),
	})
}

func get(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, object.Errorf("http.get: expected 1-2 arguments, got %d", len(args))
	}
	method, urlStr, headers, body, err := parseArgs("http.get", "GET", false, false, args)
	if err != nil {
		return nil, err
	}
	return doRequest(ctx, method, urlStr, headers, body)
}

func post(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 3 {
		return nil, object.Errorf("http.post: expected 1-3 arguments, got %d", len(args))
	}
	method, urlStr, headers, body, err := parseArgs("http.post", "POST", true, false, args)
	if err != nil {
		return nil, err
	}
	return doRequest(ctx, method, urlStr, headers, body)
}

func put(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 3 {
		return nil, object.Errorf("http.put: expected 1-3 arguments, got %d", len(args))
	}
	method, urlStr, headers, body, err := parseArgs("http.put", "PUT", true, false, args)
	if err != nil {
		return nil, err
	}
	return doRequest(ctx, method, urlStr, headers, body)
}

func del(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, object.Errorf("http.delete: expected 1-2 arguments, got %d", len(args))
	}
	method, urlStr, headers, body, err := parseArgs("http.delete", "DELETE", false, false, args)
	if err != nil {
		return nil, err
	}
	return doRequest(ctx, method, urlStr, headers, body)
}

func fetch(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, object.Errorf("http.fetch: expected 1-2 arguments, got %d", len(args))
	}
	method, urlStr, headers, body, err := parseArgs("http.fetch", "GET", true, true, args)
	if err != nil {
		return nil, err
	}
	return doRequest(ctx, method, urlStr, headers, body)
}

// parseArgs extracts url, headers, auth and body from the builtin
// arguments. All five builtins previously copy-pasted this (issue #25).
// For get/post/put/delete args[1] is the options map; for fetch (allowMethod)
// the options map may also carry "method" and "body". A trailing positional
// body argument (post/put) takes precedence over options["body"].
func parseArgs(name, method string, allowBody, allowMethod bool, args []object.Object) (m, urlStr string, headers map[string]string, body []byte, err error) {
	m = method
	urlStr, err = object.AsString(args[0])
	if err != nil {
		return
	}
	headers = make(map[string]string)

	if len(args) >= 2 {
		options, merr := object.AsMap(args[1])
		if merr != nil {
			err = fmt.Errorf("%s: options must be a map, got %s", name, args[1].Type())
			return
		}
		if merr = processAuth(name, options, headers); merr != nil {
			err = merr
			return
		}
		if headersObj, ok := options.Value()["headers"]; ok {
			headersMap, merr := object.AsMap(headersObj)
			if merr != nil {
				err = fmt.Errorf("%s: headers must be a map, got %s", name, headersObj.Type())
				return
			}
			for k, v := range headersMap.Value() {
				val, merr := object.AsString(v)
				if merr != nil {
					err = fmt.Errorf("%s: header %q must be a string", name, k)
					return
				}
				headers[k] = val
			}
		}
		if allowMethod {
			if methodObj, ok := options.Value()["method"]; ok {
				m, merr = object.AsString(methodObj)
				if merr != nil {
					err = fmt.Errorf("%s: method must be a string", name)
					return
				}
				m = strings.ToUpper(m)
			}
		}
		if allowBody {
			if bodyObj, ok := options.Value()["body"]; ok {
				body, merr = objectToBody(bodyObj)
				if merr != nil {
					err = merr
					return
				}
			}
		}
	}

	if allowBody && len(args) == 3 {
		body, err = objectToBody(args[2])
	}
	return
}

// processAuth extracts auth from options and adds an Authorization header.
func processAuth(name string, options *object.Map, headers map[string]string) error {
	authObj, ok := options.Value()["auth"]
	if !ok {
		return nil
	}
	authMap, err := object.AsMap(authObj)
	if err != nil {
		return fmt.Errorf("%s: auth must be a map with username and password keys", name)
	}
	usernameObj, hasUsername := authMap.Value()["username"]
	passwordObj, hasPassword := authMap.Value()["password"]
	if !hasUsername || !hasPassword {
		return fmt.Errorf("%s: auth map must contain 'username' and 'password' keys", name)
	}
	username, err := object.AsString(usernameObj)
	if err != nil {
		return fmt.Errorf("%s: auth.username must be a string", name)
	}
	password, err := object.AsString(passwordObj)
	if err != nil {
		return fmt.Errorf("%s: auth.password must be a string", name)
	}
	credentials := username + ":" + password
	headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(credentials))
	return nil
}

// doRequest performs the request bound to ctx (issue #25): cancelling
// the script's run aborts it in flight.
func doRequest(ctx context.Context, method, urlStr string, headers map[string]string, body []byte) (object.Object, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "fynerisor-http/1.0")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, int64(MaxResponseBytes)+1))
	if err != nil {
		return nil, err
	}
	if len(bodyBytes) > MaxResponseBytes {
		return nil, fmt.Errorf("http: response from %s exceeds %d bytes", urlStr, MaxResponseBytes)
	}

	responseHeaders := make(map[string]object.Object)
	for k, v := range resp.Header {
		if len(v) == 1 {
			responseHeaders[k] = object.NewString(v[0])
		} else {
			vals := make([]object.Object, len(v))
			for i, val := range v {
				vals[i] = object.NewString(val)
			}
			responseHeaders[k] = object.NewList(vals)
		}
	}

	// json parses the body on demand. It is complete before NewMap —
	// the old code mutated the map after wrapping it (issue #25).
	jsonMethod := object.NewBuiltin("response.json", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 0 {
			return nil, object.Errorf("response.json: expected 0 arguments, got %d", len(args))
		}
		return convert.DecodeJSON(bodyBytes)
	})

	return object.NewMap(map[string]object.Object{
		"status":     object.NewInt(int64(resp.StatusCode)),
		"statusText": object.NewString(resp.Status),
		"headers":    object.NewMap(responseHeaders),
		"body":       object.NewString(string(bodyBytes)),
		"body_bytes": object.NewBytes(bodyBytes),
		"ok":         object.NewBool(resp.StatusCode >= 200 && resp.StatusCode < 300),
		"json":       jsonMethod,
	}), nil
}

// objectToBody converts a Risor object to request body bytes.
func objectToBody(obj object.Object) ([]byte, error) {
	switch v := obj.(type) {
	case *object.String:
		return []byte(v.Value()), nil
	case *object.Bytes:
		return v.Value(), nil
	case *object.Map, *object.List:
		data, err := convert.MarshalJSON(obj, "")
		if err != nil {
			return nil, err
		}
		return data, nil
	default:
		return nil, fmt.Errorf("http: request body must be a string, bytes, map or list, got %s", obj.Type())
	}
}
