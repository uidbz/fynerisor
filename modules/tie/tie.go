package tie

import (
	"context"
	"fmt"

	"git.sr.ht/~uid/tie/client"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
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
		cfg.Webservice = url
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
	}

	tc := client.NewTieClient(cfg)
	return New(tc), nil
}

func Module() *object.Module {
	return object.NewBuiltinsModule("tie", map[string]object.Object{
		"connect": object.NewBuiltin("tie.connect", Connect),
	})
}
