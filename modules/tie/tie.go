package tie

import (
	"context"
	"errors"
	"fmt"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/uidbz/tie/client"
)

func Connect(ctx context.Context, args ...object.Object) (object.Object, error) {
	numArgs := len(args)

	if numArgs > 2 {
		return nil, fmt.Errorf("tie.connect: expected 0-2 arguments, got %d", numArgs)
	}

	// Start with defaults (includes InitKey())
	cfg := client.DefaultConfig()

	// Optional first arg: webservice URL
	if numArgs >= 1 && args[0] != object.Nil {
		url, err := object.AsString(args[0])
		if err != nil {
			return nil, err
		}
		if len(url) < 5 {
			return nil, fmt.Errorf("tie.connect: webservice URL too short (got %q)", url)
		}
		cfg.TripleStoreURL = url
	}

	// Optional second arg: options map
	if numArgs == 2 {
		optMap, ok := args[1].(*object.Map)
		if !ok {
			return nil, fmt.Errorf("tie.connect: second argument must be a map (got %s)", args[1].Type())
		}

		if username := optMap.GetWithDefault("username", nil); username != nil {
			u, err := object.AsString(username)
			if err != nil {
				return nil, fmt.Errorf("tie.connect: 'username' must be a string (got %s)", username.Type())
			}
			cfg.Username = u
		}

		if password := optMap.GetWithDefault("password", nil); password != nil {
			p, err := object.AsString(password)
			if err != nil {
				return nil, fmt.Errorf("tie.connect: 'password' must be a string (got %s)", password.Type())
			}
			cfg.Password = p
		}

		if namespace := optMap.GetWithDefault("namespace", nil); namespace != nil {
			n, err := object.AsString(namespace)
			if err != nil {
				return nil, fmt.Errorf("tie.connect: 'namespace' must be a string (got %s)", namespace.Type())
			}
			cfg.Namespace = n
		}

		if collection := optMap.GetWithDefault("collection", nil); collection != nil {
			c, err := object.AsString(collection)
			if err != nil {
				return nil, fmt.Errorf("tie.connect: 'collection' must be a string (got %s)", collection.Type())
			}
			cfg.Collection = c
		}

		if insecure := optMap.GetWithDefault("insecure", nil); insecure != nil {
			ins, err := object.AsBool(insecure)
			if err != nil {
				return nil, fmt.Errorf("tie.connect: 'insecure' must be a boolean (got %s)", insecure.Type())
			}
			cfg.WebserviceInsecure = ins
		}

		// The filehost is a separate service on its own port, so without this the
		// config keeps DefaultConfig's hardcoded localhost address and upload/
		// download can only ever reach a local daemon.
		if fh := optMap.GetWithDefault("filehost", nil); fh != nil {
			host, err := fileHostFromOption(fh)
			if err != nil {
				return nil, err
			}
			// Replace the "default" entry so an empty host name still resolves.
			cfg.FileHosts = map[string]client.FileHost{"default": host}
			cfg.DefaultFileHosts = []string{"default"}
		}
	}

	tc := client.NewTieClient(cfg)
	return New(tc), nil
}

// fileHostFromOption decodes the 'filehost' connect option, which is either a
// URL string or a map of {url, insecure, username, password, store}.
func fileHostFromOption(obj object.Object) (client.FileHost, error) {
	if s, ok := obj.(*object.String); ok {
		return client.FileHost{URL: s.Value()}, nil
	}

	optMap, ok := obj.(*object.Map)
	if !ok {
		return client.FileHost{}, fmt.Errorf("tie.connect: 'filehost' must be a string or a map (got %s)", obj.Type())
	}

	var host client.FileHost
	for key, field := range map[string]*string{
		"url":      &host.URL,
		"username": &host.Username,
		"password": &host.Password,
		"store":    &host.Store,
	} {
		if v := optMap.GetWithDefault(key, nil); v != nil {
			s, err := object.AsString(v)
			if err != nil {
				return client.FileHost{}, fmt.Errorf("tie.connect: filehost '%s' must be a string (got %s)", key, v.Type())
			}
			*field = s
		}
	}

	if v := optMap.GetWithDefault("insecure", nil); v != nil {
		ins, err := object.AsBool(v)
		if err != nil {
			return client.FileHost{}, fmt.Errorf("tie.connect: filehost 'insecure' must be a boolean (got %s)", v.Type())
		}
		host.Insecure = ins
	}

	if host.URL == "" {
		return client.FileHost{}, errors.New("tie.connect: filehost 'url' is required")
	}
	return host, nil
}

func Module() *object.Module {
	return object.NewBuiltinsModule("tie", map[string]object.Object{
		"connect": object.NewBuiltin("tie.connect", Connect),
	})
}
