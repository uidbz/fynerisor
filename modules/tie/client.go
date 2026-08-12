package tie

import (
	"context"
	"errors"
	"fmt"

	"git.sr.ht/~uid/tie/client"
	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const ClientType object.Type = "tie.client"

type Client struct {
	tc *client.TieClient
}

func New(tc *client.TieClient) *Client {
	return &Client{tc: tc}
}

func (c *Client) Type() object.Type {
	return ClientType
}

func (c *Client) Inspect() string {
	return "tie.client"
}

func (c *Client) Interface() interface{} {
	return c.tc
}

func (c *Client) IsTruthy() bool {
	return c.tc != nil
}

func (c *Client) Cost() int {
	return 8
}

func (c *Client) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("type error: unable to marshal tie.client")
}

func (c *Client) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return object.Errorf("eval error: unsupported operation for tie.client: %v", opType), nil
}

func (c *Client) Equals(other object.Object) bool {
	if other.Type() != ClientType {
		return false
	}
	return c.tc == other.(*Client).tc
}

func (c *Client) SetAttr(name string, value object.Object) error {
	return fmt.Errorf("attribute error: tie.client object has no attribute %q", name)
}

func (c *Client) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "add":
		return object.NewBuiltin("tie.add", c.Add), true
	case "get":
		return object.NewBuiltin("tie.get", c.Get), true
	case "set":
		return object.NewBuiltin("tie.set", c.Set), true
	case "delete":
		return object.NewBuiltin("tie.delete", c.Delete), true
	case "update":
		return object.NewBuiltin("tie.update", c.Update), true
	case "query":
		return object.NewBuiltin("tie.query", c.Query), true
	case "expand":
		return object.NewBuiltin("tie.expand", c.Expand), true
	case "associated":
		return object.NewBuiltin("tie.associated", c.Associated), true
	case "exists":
		return object.NewBuiltin("tie.exists", c.Exists), true
	case "sync":
		return object.NewBuiltin("tie.sync", c.Sync), true
	case "batch":
		return object.NewBuiltin("tie.batch", c.Batch), true
	case "dump":
		return object.NewBuiltin("tie.dump", c.Dump), true
	case "restore":
		return object.NewBuiltin("tie.restore", c.Restore), true
	case "drop":
		return object.NewBuiltin("tie.drop", c.Drop), true
	}
	return nil, false
}

func (c *Client) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "add"},
		{Name: "get"},
		{Name: "set"},
		{Name: "delete"},
		{Name: "update"},
		{Name: "query"},
		{Name: "expand"},
		{Name: "associated"},
		{Name: "exists"},
		{Name: "sync"},
		{Name: "batch"},
		{Name: "dump"},
		{Name: "restore"},
		{Name: "drop"},
	}
}

func (c *Client) Add(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("tie.add: expected 3 arguments, got %d", len(args))
	}

	key, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	value1, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	value2, err := object.AsString(args[2])
	if err != nil {
		return nil, err
	}

	_, err = c.tc.Add(key, value1, value2)
	if err != nil {
		return nil, err
	}

	return object.Nil, nil
}

func (c *Client) Get(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tie.get: expected 1 argument, got %d", len(args))
	}

	key, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	row, err := c.tc.Get(key)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return object.Nil, nil
		}
		return nil, err
	}

	return rowToObject(row), nil
}

func (c *Client) Set(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("tie.set: expected 3 arguments, got %d", len(args))
	}

	key, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	relation, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	valuesList, ok := args[2].(*object.List)
	if !ok {
		return nil, fmt.Errorf("tie.set: third argument must be a list (got %s)", args[2].Type())
	}

	values := make([]string, 0, len(valuesList.Value()))
	for i, valObj := range valuesList.Value() {
		val, err := object.AsString(valObj)
		if err != nil {
			return nil, fmt.Errorf("tie.set: list element %d must be a string: %w", i, err)
		}
		values = append(values, val)
	}

	err = c.tc.Set(key, relation, values)
	if err != nil {
		return nil, err
	}

	return object.Nil, nil
}

func (c *Client) Delete(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("tie.delete: expected 3 arguments, got %d", len(args))
	}

	key, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	value1, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	value2, err := object.AsString(args[2])
	if err != nil {
		return nil, err
	}

	_, err = c.tc.Delete(key, value1, value2)
	if err != nil {
		return nil, err
	}

	return object.Nil, nil
}

func (c *Client) Update(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 4 {
		return nil, fmt.Errorf("tie.update: expected 4 arguments, got %d", len(args))
	}

	key, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	value1, err := object.AsString(args[1])
	if err != nil {
		return nil, err
	}

	value2, err := object.AsString(args[2])
	if err != nil {
		return nil, err
	}

	newValue2, err := object.AsString(args[3])
	if err != nil {
		return nil, err
	}

	update := c.tc.NewUpdate(key, value1, value2, newValue2)
	_, err = c.tc.Update(update)
	if err != nil {
		return nil, err
	}

	return object.Nil, nil
}

func (c *Client) Query(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tie.query: expected 1 argument, got %d", len(args))
	}

	specMap, ok := args[0].(*object.Map)
	if !ok {
		return nil, fmt.Errorf("tie.query: argument must be a map (got %s)", args[0].Type())
	}

	spec := client.QuerySpec{}

	if termsObj := specMap.GetWithDefault("terms", nil); termsObj != nil {
		termsList, ok := termsObj.(*object.List)
		if !ok {
			return nil, fmt.Errorf("tie.query: 'terms' must be a list (got %s)", termsObj.Type())
		}
		spec.Terms = make([]string, 0, len(termsList.Value()))
		for i, termObj := range termsList.Value() {
			term, err := object.AsString(termObj)
			if err != nil {
				return nil, fmt.Errorf("tie.query: terms[%d] must be a string: %w", i, err)
			}
			spec.Terms = append(spec.Terms, term)
		}
	}

	if excludeObj := specMap.GetWithDefault("exclude", nil); excludeObj != nil {
		excludeList, ok := excludeObj.(*object.List)
		if !ok {
			return nil, fmt.Errorf("tie.query: 'exclude' must be a list (got %s)", excludeObj.Type())
		}
		spec.Exclude = make([]string, 0, len(excludeList.Value()))
		for i, exObj := range excludeList.Value() {
			ex, err := object.AsString(exObj)
			if err != nil {
				return nil, fmt.Errorf("tie.query: exclude[%d] must be a string: %w", i, err)
			}
			spec.Exclude = append(spec.Exclude, ex)
		}
	}

	if scopeObj := specMap.GetWithDefault("scope", nil); scopeObj != nil {
		scope, err := object.AsString(scopeObj)
		if err != nil {
			return nil, fmt.Errorf("tie.query: 'scope' must be a string: %w", err)
		}
		spec.Scope = scope
	}

	if filterObj := specMap.GetWithDefault("filter", nil); filterObj != nil {
		filter, err := object.AsString(filterObj)
		if err != nil {
			return nil, fmt.Errorf("tie.query: 'filter' must be a string: %w", err)
		}
		spec.Filter = filter
	}

	if sortObj := specMap.GetWithDefault("sort_by", nil); sortObj != nil {
		sortBy, err := object.AsString(sortObj)
		if err != nil {
			return nil, fmt.Errorf("tie.query: 'sort_by' must be a string: %w", err)
		}
		spec.SortBy = sortBy
	}

	if reverseObj := specMap.GetWithDefault("reverse", nil); reverseObj != nil {
		reverse, err := object.AsBool(reverseObj)
		if err != nil {
			return nil, fmt.Errorf("tie.query: 'reverse' must be a boolean: %w", err)
		}
		spec.Reverse = reverse
	}

	if expandObj := specMap.GetWithDefault("expand", nil); expandObj != nil {
		expand, err := object.AsBool(expandObj)
		if err != nil {
			return nil, fmt.Errorf("tie.query: 'expand' must be a boolean: %w", err)
		}
		spec.Expand = expand
	}

	if offsetObj := specMap.GetWithDefault("offset", nil); offsetObj != nil {
		offset, err := object.AsInt(offsetObj)
		if err != nil {
			return nil, fmt.Errorf("tie.query: 'offset' must be an integer: %w", err)
		}
		spec.Offset = int(offset)
	}

	if limitObj := specMap.GetWithDefault("limit", nil); limitObj != nil {
		limit, err := object.AsInt(limitObj)
		if err != nil {
			return nil, fmt.Errorf("tie.query: 'limit' must be an integer: %w", err)
		}
		spec.Limit = int(limit)
	}

	rows, _, err := c.tc.Query(spec)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return object.NewList([]object.Object{}), nil
		}
		return nil, err
	}

	rowList := object.NewList(make([]object.Object, 0, len(rows)))
	for _, row := range rows {
		rowList.Append(rowToObject(row))
	}

	return rowList, nil
}

func (c *Client) Expand(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tie.expand: expected 1 argument, got %d", len(args))
	}

	keysList, ok := args[0].(*object.List)
	if !ok {
		return nil, fmt.Errorf("tie.expand: argument must be a list (got %s)", args[0].Type())
	}

	keys := make([]string, 0, len(keysList.Value()))
	for i, keyObj := range keysList.Value() {
		key, err := object.AsString(keyObj)
		if err != nil {
			return nil, fmt.Errorf("tie.expand: keys[%d] must be a string: %w", i, err)
		}
		keys = append(keys, key)
	}

	rows, err := c.tc.Expand(keys)
	if err != nil {
		return nil, err
	}

	rowList := object.NewList(make([]object.Object, 0, len(rows)))
	for _, row := range rows {
		rowList.Append(rowToObject(row))
	}

	return rowList, nil
}

func (c *Client) Associated(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tie.associated: expected 1 argument, got %d", len(args))
	}

	key, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	reply, err := c.tc.Associated(key)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return object.NewList([]object.Object{}), nil
		}
		return nil, err
	}

	tripleList := object.NewList(make([]object.Object, 0))
	reply.Result.ForEachValue2(func(key, value1, value2 string) {
		tripleMap := object.NewMap(map[string]object.Object{
			"key":    object.NewString(key),
			"value1": object.NewString(value1),
			"value2": object.NewString(value2),
		})
		tripleList.Append(tripleMap)
	})

	return tripleList, nil
}

func (c *Client) Exists(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tie.exists: expected 1 argument, got %d", len(args))
	}

	key, err := object.AsString(args[0])
	if err != nil {
		return nil, err
	}

	exists := c.tc.Exists(key)
	return object.NewBool(exists), nil
}

func (c *Client) Sync(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("tie.sync: expected 0 arguments, got %d", len(args))
	}

	err := c.tc.Sync()
	if err != nil {
		return nil, err
	}

	return object.Nil, nil
}

func (c *Client) Batch(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("tie.batch: expected 0 arguments, got %d", len(args))
	}

	batch := c.tc.NewBatch()
	return NewBatch(c.tc, batch), nil
}

func (c *Client) Dump(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("tie.dump: expected 0 arguments, got %d", len(args))
	}

	reply, err := c.tc.Dump()
	if err != nil {
		return nil, err
	}

	tripleList := object.NewList(make([]object.Object, 0, len(reply.Triples)))
	for _, triple := range reply.Triples {
		tripleArray := object.NewList([]object.Object{
			object.NewString(triple.Key),
			object.NewString(triple.Value1),
			object.NewString(triple.Value2),
		})
		tripleList.Append(tripleArray)
	}

	return tripleList, nil
}

func (c *Client) Restore(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("tie.restore: expected 1 argument, got %d", len(args))
	}

	triplesList, ok := args[0].(*object.List)
	if !ok {
		return nil, fmt.Errorf("tie.restore: argument must be a list (got %s)", args[0].Type())
	}

	triples := make([][3]string, 0, len(triplesList.Value()))
	for i, tripleObj := range triplesList.Value() {
		tripleList, ok := tripleObj.(*object.List)
		if !ok {
			return nil, fmt.Errorf("tie.restore: triples[%d] must be a list (got %s)", i, tripleObj.Type())
		}
		tripleVals := tripleList.Value()
		if len(tripleVals) != 3 {
			return nil, fmt.Errorf("tie.restore: triples[%d] must have 3 elements (got %d)", i, len(tripleVals))
		}

		key, err := object.AsString(tripleVals[0])
		if err != nil {
			return nil, fmt.Errorf("tie.restore: triples[%d][0] must be a string: %w", i, err)
		}
		value1, err := object.AsString(tripleVals[1])
		if err != nil {
			return nil, fmt.Errorf("tie.restore: triples[%d][1] must be a string: %w", i, err)
		}
		value2, err := object.AsString(tripleVals[2])
		if err != nil {
			return nil, fmt.Errorf("tie.restore: triples[%d][2] must be a string: %w", i, err)
		}

		triples = append(triples, [3]string{key, value1, value2})
	}

	err := c.tc.Restore(triples)
	if err != nil {
		return nil, err
	}

	return object.Nil, nil
}

func (c *Client) Drop(ctx context.Context, args ...object.Object) (object.Object, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("tie.drop: expected 0 arguments, got %d", len(args))
	}

	err := c.tc.DropCollection()
	if err != nil {
		return nil, err
	}

	return object.Nil, nil
}

// rowToObject converts a tiedb.Row to a Risor map object
func rowToObject(r client.Row) object.Object {
	attrs := object.NewMap(make(map[string]object.Object))
	for relation, values := range r.Attributes {
		valueList := object.NewList(make([]object.Object, 0, len(values)))
		for _, v := range values {
			valueList.Append(object.NewString(v))
		}
		attrs.Set(relation, valueList)
	}

	rowMap := object.NewMap(map[string]object.Object{
		"key":        object.NewString(r.Key),
		"attributes": attrs,
	})
	return rowMap
}
