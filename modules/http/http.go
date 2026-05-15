// Package http provides HTTP client functionality for Risor scripts.
// Ported from risor v1 to work with risor v2 object system.
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
package http

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

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

// get performs an HTTP GET request
func get(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return object.Errorf("http.get: expected 1-2 arguments, got %d", len(args)), nil
	}

	urlStr, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	if len(args) == 2 {
		options, err := object.AsMap(args[1])
		if err != nil {
			return nil, err
		}

		// Process auth if present
		if err := processAuth(options, headers); err != nil {
			return nil, err
		}

		// Process headers if present
		if headersObj, ok := options.Value()["headers"]; ok {
			headersMap, err := object.AsMap(headersObj)
			if err != nil {
				return nil, err
			}
			for k, v := range headersMap.Value() {
				val, err := object.AsString(v)
				if err != nil {
					return nil, err
				}
				headers[k] = val
			}
		}
	}

	return doRequest(ctx, "GET", urlStr, headers, nil)
}

// post performs an HTTP POST request
func post(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 3 {
		return object.Errorf("http.post: expected 1-3 arguments, got %d", len(args)), nil
	}

	urlStr, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	var body []byte

	if len(args) >= 2 {
		options, err := object.AsMap(args[1])
		if err != nil {
			return nil, err
		}

		// Process auth if present
		if err := processAuth(options, headers); err != nil {
			return nil, err
		}

		// Process headers if present
		if headersObj, ok := options.Value()["headers"]; ok {
			headersMap, err := object.AsMap(headersObj)
			if err != nil {
				return nil, err
			}
			for k, v := range headersMap.Value() {
				val, err := object.AsString(v)
				if err != nil {
					return nil, err
				}
				headers[k] = val
			}
		}
	}

	if len(args) == 3 {
		body, err = objectToBody(args[2])
		if err != nil {
			return nil, err
		}
	}

	return doRequest(ctx, "POST", urlStr, headers, body)
}

// put performs an HTTP PUT request
func put(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 3 {
		return object.Errorf("http.put: expected 1-3 arguments, got %d", len(args)), nil
	}

	urlStr, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	var body []byte

	if len(args) >= 2 {
		options, err := object.AsMap(args[1])
		if err != nil {
			return nil, err
		}

		// Process auth if present
		if err := processAuth(options, headers); err != nil {
			return nil, err
		}

		// Process headers if present
		if headersObj, ok := options.Value()["headers"]; ok {
			headersMap, err := object.AsMap(headersObj)
			if err != nil {
				return nil, err
			}
			for k, v := range headersMap.Value() {
				val, err := object.AsString(v)
				if err != nil {
					return nil, err
				}
				headers[k] = val
			}
		}
	}

	if len(args) == 3 {
		body, err = objectToBody(args[2])
		if err != nil {
			return nil, err
		}
	}

	return doRequest(ctx, "PUT", urlStr, headers, body)
}

// del performs an HTTP DELETE request
func del(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return object.Errorf("http.delete: expected 1-2 arguments, got %d", len(args)), nil
	}

	urlStr, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	if len(args) == 2 {
		options, err := object.AsMap(args[1])
		if err != nil {
			return nil, err
		}

		// Process auth if present
		if err := processAuth(options, headers); err != nil {
			return nil, err
		}

		// Process headers if present
		if headersObj, ok := options.Value()["headers"]; ok {
			headersMap, err := object.AsMap(headersObj)
			if err != nil {
				return nil, err
			}
			for k, v := range headersMap.Value() {
				val, err := object.AsString(v)
				if err != nil {
					return nil, err
				}
				headers[k] = val
			}
		}
	}

	return doRequest(ctx, "DELETE", urlStr, headers, nil)
}

// fetch performs a customizable HTTP request
func fetch(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) < 1 || len(args) > 2 {
		return object.Errorf("http.fetch: expected 1-2 arguments, got %d", len(args)), nil
	}

	urlStr, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	method := "GET"
	var headers map[string]string
	var body []byte

	if len(args) == 2 {
		options, err := object.AsMap(args[1])
		if err != nil {
			return nil, err
		}

		// Extract method
		if methodObj, ok := options.Value()["method"]; ok {
			method, err = object.AsString(methodObj)
			if err != nil {
				return nil, err
			}
			method = strings.ToUpper(method)
		}

		// Process auth if present
		headers = make(map[string]string)
		if err := processAuth(options, headers); err != nil {
			return nil, err
		}

		// Extract headers
		if headersObj, ok := options.Value()["headers"]; ok {
			headersMap, err := object.AsMap(headersObj)
			if err != nil {
				return nil, err
			}
			for k, v := range headersMap.Value() {
				val, err := object.AsString(v)
				if err != nil {
					return nil, err
				}
				headers[k] = val
			}
		}

		// Extract body
		if bodyObj, ok := options.Value()["body"]; ok {
			body, err = objectToBody(bodyObj)
			if err != nil {
				return nil, err
			}
		}
	}

	return doRequest(ctx, method, urlStr, headers, body)
}

// processAuth extracts auth from options and adds Authorization header
func processAuth(options *object.Map, headers map[string]string) error {
	if authObj, ok := options.Value()["auth"]; ok {
		authMap, err := object.AsMap(authObj)
		if err != nil {
			return fmt.Errorf("auth must be a map with username and password keys")
		}

		usernameObj, hasUsername := authMap.Value()["username"]
		passwordObj, hasPassword := authMap.Value()["password"]

		if !hasUsername || !hasPassword {
			return fmt.Errorf("auth map must contain 'username' and 'password' keys")
		}

		username, err := object.AsString(usernameObj)
		if err != nil {
			return fmt.Errorf("auth.username must be a string")
		}

		password, err := object.AsString(passwordObj)
		if err != nil {
			return fmt.Errorf("auth.password must be a string")
		}

		// Create Basic Auth header
		credentials := username + ":" + password
		encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
		headers["Authorization"] = "Basic " + encoded
	}

	return nil
}

// doRequest performs the actual HTTP request
func doRequest(ctx context.Context, method, urlStr string, headers map[string]string, body []byte) (object.Object, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	// Try without context to see if that's the issue
	req, err := http.NewRequest(method, urlStr, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Set default User-Agent if not provided
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "fynerisor-http/1.0")
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Build response object
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

	responseMap := map[string]object.Object{
		"status":     object.NewInt(int64(resp.StatusCode)),
		"statusText": object.NewString(resp.Status),
		"headers":    object.NewMap(responseHeaders),
		"body":       object.NewString(string(bodyBytes)),
		"ok":         object.NewBool(resp.StatusCode >= 200 && resp.StatusCode < 300),
	}

	response := object.NewMap(responseMap)

	// Add json() method to response object
	jsonMethod := object.NewBuiltin("response.json", func(ctx context.Context, args ...object.Object) (object.Object, error) {
		if len(args) != 0 {
			return object.Errorf("response.json: expected 0 arguments, got %d", len(args)), nil
		}

		var data interface{}
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			return nil, err
		}

		result, err := interfaceToObject(data)
		if err != nil {
			return nil, err
		}

		return result, nil
	})

	responseMap["json"] = jsonMethod

	return response, nil
}

// objectToBody converts a Risor object to request body bytes
func objectToBody(obj object.Object) ([]byte, error) {
	switch v := obj.(type) {
	case *object.String:
		return []byte(v.Value()), nil
	case *object.Map, *object.List:
		// Convert to JSON
		data, err := objectToInterface(obj)
		if err != nil {
			return nil, err
		}
		return json.Marshal(data)
	default:
		return []byte(fmt.Sprintf("%v", obj)), nil
	}
}

// objectToInterface converts a Risor object to a Go interface for JSON marshaling
func objectToInterface(obj object.Object) (interface{}, error) {
	switch v := obj.(type) {
	case *object.String:
		return v.Value(), nil
	case *object.Int:
		return v.Value(), nil
	case *object.Float:
		return v.Value(), nil
	case *object.Bool:
		return v.Value(), nil
	case *object.List:
		result := make([]interface{}, len(v.Value()))
		for i, item := range v.Value() {
			val, err := objectToInterface(item)
			if err != nil {
				return nil, err
			}
			result[i] = val
		}
		return result, nil
	case *object.Map:
		result := make(map[string]interface{})
		for k, val := range v.Value() {
			converted, err := objectToInterface(val)
			if err != nil {
				return nil, err
			}
			result[k] = converted
		}
		return result, nil
	default:
		return fmt.Sprintf("%v", obj), nil
	}
}

// interfaceToObject converts a Go interface (from JSON) to a Risor object
func interfaceToObject(data interface{}) (object.Object, error) {
	switch v := data.(type) {
	case string:
		return object.NewString(v), nil
	case float64:
		return object.NewFloat(v), nil
	case int64:
		return object.NewInt(v), nil
	case int:
		return object.NewInt(int64(v)), nil
	case bool:
		return object.NewBool(v), nil
	case nil:
		return object.Nil, nil
	case []interface{}:
		items := make([]object.Object, len(v))
		for i, item := range v {
			obj, err := interfaceToObject(item)
			if err != nil {
				return nil, err
			}
			items[i] = obj
		}
		return object.NewList(items), nil
	case map[string]interface{}:
		objMap := make(map[string]object.Object)
		for k, val := range v {
			obj, err := interfaceToObject(val)
			if err != nil {
				return nil, err
			}
			objMap[k] = obj
		}
		return object.NewMap(objMap), nil
	default:
		return object.NewString(fmt.Sprintf("%v", v)), nil
	}
}
