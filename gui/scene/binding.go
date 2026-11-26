package scene

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	risor "github.com/deepnoodle-ai/risor/v2"
	"github.com/deepnoodle-ai/risor/v2/pkg/bytecode"

	scenedoc "github.com/uidbz/fynerisor/scene"
)

// segment is one piece of a ${...} property template: literal text, or an
// expression evaluated against the BindValues map.
type segment struct {
	text string
	expr bool
}

// parseTemplate splits a property string into literal text and ${...}
// expressions. Braces inside an expression nest (`${ ({n}).n }`) and braces
// inside string literals do not count (`${"}" + x}`). An unterminated or
// empty ${ stays literal text.
func parseTemplate(s string) []segment {
	var out []segment
	lit := 0
	for i := 0; i+1 < len(s); i++ {
		if s[i] != '$' || s[i+1] != '{' {
			continue
		}
		end := matchBrace(s, i+2)
		if end < 0 {
			break
		}
		if expr := s[i+2 : end]; strings.TrimSpace(expr) != "" {
			if i > lit {
				out = append(out, segment{text: s[lit:i]})
			}
			out = append(out, segment{text: expr, expr: true})
			lit = end + 1
		}
		i = end
	}
	if lit < len(s) {
		out = append(out, segment{text: s[lit:]})
	}
	return out
}

// matchBrace returns the index of the '}' closing an expression that starts
// at from (just after "${"), or -1.
func matchBrace(s string, from int) int {
	depth := 1
	var quote byte
	for j := from; j < len(s); j++ {
		c := s[j]
		if quote != 0 {
			switch {
			case c == '\\' && quote != '`':
				j++
			case c == quote:
				quote = 0
			}
			continue
		}
		switch c {
		case '"', '\'', '`':
			quote = c
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return j
			}
		}
	}
	return -1
}

// builtinEnv is Risor's standard globals (len, str, keys, math, strings, …).
// risor.WithEnv replaces the default environment, so binding expressions
// get these explicitly; BindValues keys shadow them.
var builtinEnv = sync.OnceValue(risor.Builtins)

// bindEnv is the evaluation environment of one binding refresh.
type bindEnv struct {
	env map[string]any
	sig string // sorted global names: compiled code is bound to them
}

func newBindEnv(vals map[string]any) bindEnv {
	env := maps.Clone(builtinEnv())
	maps.Copy(env, vals)
	return bindEnv{env: env, sig: strings.Join(slices.Sorted(maps.Keys(env)), "\x00")}
}

// exprCache keeps compiled binding expressions. Risor bytecode resolves
// globals by index, so an entry is keyed by the expression and the exact
// set of global names it was compiled with; values may change freely.
type exprCache struct {
	mu    sync.Mutex
	codes map[string]*bytecode.Code
}

// maxCachedExprs bounds the cache; a values closure whose key set keeps
// changing would otherwise grow it without limit.
const maxCachedExprs = 1024

func (c *exprCache) eval(ctx context.Context, expr string, e bindEnv) (any, error) {
	key := e.sig + "\x01" + expr
	c.mu.Lock()
	code := c.codes[key]
	c.mu.Unlock()
	if code == nil {
		var err error
		code, err = risor.Compile(ctx, expr, risor.WithEnv(e.env))
		if err != nil {
			return nil, err
		}
		c.mu.Lock()
		if c.codes == nil || len(c.codes) >= maxCachedExprs {
			c.codes = make(map[string]*bytecode.Code)
		}
		c.codes[key] = code
		c.mu.Unlock()
	}
	return risor.Run(ctx, code, risor.WithEnv(e.env))
}

// resolve evaluates a parsed template. A template that is exactly one
// expression keeps the value's type for number/int/bool properties; any
// other template is rendered as a string.
func (c *exprCache) resolve(ctx context.Context, segs []segment, kind scenedoc.PropKind, e bindEnv) (any, error) {
	if len(segs) == 1 && segs[0].expr {
		v, err := c.eval(ctx, segs[0].text, e)
		if err != nil {
			return nil, err
		}
		switch kind {
		case scenedoc.PropNumber, scenedoc.PropInt, scenedoc.PropBool:
			return v, nil
		default:
			if v == nil {
				return "", nil
			}
			return fmt.Sprint(v), nil
		}
	}
	var b strings.Builder
	for _, s := range segs {
		if !s.expr {
			b.WriteString(s.text)
			continue
		}
		v, err := c.eval(ctx, s.text, e)
		if err != nil {
			return nil, err
		}
		if v != nil {
			b.WriteString(fmt.Sprint(v))
		}
	}
	return b.String(), nil
}
