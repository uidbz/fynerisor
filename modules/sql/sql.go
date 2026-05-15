package sql

import (
	"context"
	"fmt"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
)

func Connect(ctx context.Context, args ...object.Object) (object.Object, error) {
	numArgs := len(args)

	if numArgs < 1 || numArgs > 2 {
		return nil, fmt.Errorf("sql.connect: expected 1-2 arguments, got %d", numArgs)
	}

	connStr, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	// Default options
	stream := false

	// Check for options map as second argument
	if numArgs == 2 {
		optMap, ok := args[1].(*object.Map)
		if !ok {
			return nil, fmt.Errorf("sql.connect() second argument must be a map (got %s)", args[1].Type())
		}

		// Process stream option if provided
		if streamVal := optMap.GetWithDefault("stream", nil); streamVal != nil {
			streamBool, err := object.AsBool(streamVal)
			if err != nil {
				return nil, fmt.Errorf("sql.connect() 'stream' option must be a boolean (got %s)", streamVal.Type())
			}
			stream = streamBool
		}
	}

	db, connErr := New(ctx, connStr, stream)
	if connErr != nil {
		return nil, connErr
	}

	return db, nil
}

func Module() *object.Module {
	return object.NewBuiltinsModule("sql", map[string]object.Object{
		"connect": object.NewBuiltin("sql.connect", Connect),
	})
}
