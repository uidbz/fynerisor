package scene

import (
	"fmt"
	"time"
)

// OpKind identifies a document operation. Operations are the only way the
// editor mutates a document, which makes undo/redo and incremental rendering
// exact.
type OpKind string

const (
	OpSetProp  OpKind = "setProp"
	OpSetPlace OpKind = "setPlace"
	OpSetField OpKind = "setField"
	OpInsert   OpKind = "insert"
	OpRemove   OpKind = "remove"
	OpMove     OpKind = "move"
)

// Op is one invertible document edit.
type Op struct {
	Kind       OpKind
	Node       string // subject of set/remove/move
	Parent     string // target parent for insert/move
	Index      int    // target child index (after removal, for move)
	FromParent string // move: source parent
	FromIndex  int    // move: source index
	Key        string // setProp: property name; setField: name, hidden, locked, style, on
	Old, New   any
	Subtree    *Node // insert/remove: the full subtree (required for exact undo)
}

// Tx is one undoable edit: a label, an optional merge key and the operations.
type Tx struct {
	Label    string
	MergeKey string
	Ops      []Op
}

// Fields names the parts of a node that changed. Rendering uses it to apply
// only what changed.
type Fields map[string]bool

// Change is emitted by Apply for every operation.
type Change struct {
	Kind       OpKind
	Node       string
	Parent     string
	Index      int
	FromParent string
	FromIndex  int
	Fields     Fields
	OldPlace   Place
	NewPlace   Place
	Subtree    *Node
}

// Apply applies the transaction atomically: it runs the ops on a clone and
// only replaces the document when every op succeeded.
func (d *Document) Apply(tx Tx) ([]Change, error) {
	if d == nil || d.Root == nil {
		return nil, fmt.Errorf("scene: document has no root")
	}
	clone := d.Clone()
	changes, err := clone.applyOps(tx)
	if err != nil {
		return nil, err
	}
	*d = *clone
	return changes, nil
}

func (d *Document) applyOps(tx Tx) ([]Change, error) {
	changes := make([]Change, 0, len(tx.Ops))
	for i, op := range tx.Ops {
		ch, err := d.applyOp(op)
		if err != nil {
			return nil, fmt.Errorf("scene: op %d (%s): %w", i, op.Kind, err)
		}
		changes = append(changes, ch)
	}
	return changes, nil
}

func (d *Document) applyOp(op Op) (Change, error) {
	switch op.Kind {
	case OpSetProp:
		n, _, _, ok := d.Locate(op.Node)
		if !ok {
			return Change{}, fmt.Errorf("node %q not found", op.Node)
		}
		if n.Props == nil {
			n.Props = make(map[string]any)
		}
		if op.New == nil {
			delete(n.Props, op.Key)
		} else {
			n.Props[op.Key] = cloneValue(op.New)
		}
		return Change{Kind: op.Kind, Node: op.Node, Fields: Fields{op.Key: true}}, nil

	case OpSetPlace:
		n, _, _, ok := d.Locate(op.Node)
		if !ok {
			return Change{}, fmt.Errorf("node %q not found", op.Node)
		}
		old := n.Place
		np, ok := op.New.(Place)
		if !ok {
			return Change{}, fmt.Errorf("place must be a Place")
		}
		n.Place = np
		return Change{
			Kind: op.Kind, Node: op.Node, Fields: Fields{"place": true},
			OldPlace: old, NewPlace: np,
		}, nil

	case OpSetField:
		n, _, _, ok := d.Locate(op.Node)
		if !ok {
			return Change{}, fmt.Errorf("node %q not found", op.Node)
		}
		if err := setNodeField(n, op.Key, op.New); err != nil {
			return Change{}, err
		}
		return Change{Kind: op.Kind, Node: op.Node, Fields: Fields{op.Key: true}}, nil

	case OpInsert:
		if op.Subtree == nil {
			return Change{}, fmt.Errorf("insert needs a subtree")
		}
		parent, _, _, ok := d.Locate(op.Parent)
		if !ok {
			return Change{}, fmt.Errorf("parent %q not found", op.Parent)
		}
		if d.Count()+op.Subtree.Count() > MaxNodes {
			return Change{}, fmt.Errorf("insert would exceed %d nodes", MaxNodes)
		}
		if err := d.checkSubtreeIDs(op.Subtree); err != nil {
			return Change{}, err
		}
		idx := clamp(op.Index, 0, len(parent.Children))
		child := op.Subtree.Clone()
		parent.Children = append(parent.Children, nil)
		copy(parent.Children[idx+1:], parent.Children[idx:])
		parent.Children[idx] = child
		return Change{Kind: op.Kind, Node: child.ID, Parent: parent.ID, Index: idx, Subtree: child}, nil

	case OpRemove:
		node, parent, idx, ok := d.Locate(op.Node)
		if !ok {
			return Change{}, fmt.Errorf("node %q not found", op.Node)
		}
		if parent == nil {
			return Change{}, fmt.Errorf("cannot remove the root node")
		}
		sub := node.Clone()
		parent.Children = append(parent.Children[:idx], parent.Children[idx+1:]...)
		return Change{Kind: op.Kind, Node: node.ID, Parent: parent.ID, Index: idx, Subtree: sub}, nil

	case OpMove:
		node, fromParent, fromIdx, ok := d.Locate(op.Node)
		if !ok {
			return Change{}, fmt.Errorf("node %q not found", op.Node)
		}
		if fromParent == nil {
			return Change{}, fmt.Errorf("cannot move the root node")
		}
		target, _, _, ok := d.Locate(op.Parent)
		if !ok {
			return Change{}, fmt.Errorf("target parent %q not found", op.Parent)
		}
		if node.ID == target.ID || d.IsAncestor(node.ID, target.ID) {
			return Change{}, fmt.Errorf("cannot move %q into its own subtree", op.Node)
		}
		fromParent.Children = append(fromParent.Children[:fromIdx], fromParent.Children[fromIdx+1:]...)
		idx := clamp(op.Index, 0, len(target.Children))
		target.Children = append(target.Children, nil)
		copy(target.Children[idx+1:], target.Children[idx:])
		target.Children[idx] = node
		return Change{
			Kind: op.Kind, Node: node.ID, Parent: target.ID, Index: idx,
			FromParent: fromParent.ID, FromIndex: fromIdx,
		}, nil
	}
	return Change{}, fmt.Errorf("unknown op kind %q", op.Kind)
}

func (d *Document) checkSubtreeIDs(sub *Node) error {
	existing := d.IDs()
	dup := ""
	sub.Walk(func(n *Node) bool {
		if existing[n.ID] != nil {
			dup = n.ID
			return false
		}
		return true
	})
	if dup != "" {
		return fmt.Errorf("node id %q already exists", dup)
	}
	return nil
}

func setNodeField(n *Node, key string, value any) error {
	switch key {
	case "type":
		s, ok := value.(string)
		if !ok || s == "" {
			return fmt.Errorf("type must be a non-empty string")
		}
		n.Type = s
	case "name":
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("name must be a string")
		}
		n.Name = s
	case "hidden":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("hidden must be a bool")
		}
		n.Hidden = b
	case "locked":
		b, ok := value.(bool)
		if !ok {
			return fmt.Errorf("locked must be a bool")
		}
		n.Locked = b
	case "style":
		switch s := value.(type) {
		case nil:
			n.Style = nil
		case *Style:
			cp := *s
			n.Style = &cp
		case Style:
			cp := s
			n.Style = &cp
		default:
			return fmt.Errorf("style must be a Style or nil")
		}
	case "on":
		switch m := value.(type) {
		case nil:
			n.On = nil
		case map[string][]Action:
			n.On = make(map[string][]Action, len(m))
			for ev, acts := range m {
				if ev == "" || len(acts) == 0 {
					continue
				}
				cp := make([]Action, len(acts))
				copy(cp, acts)
				n.On[ev] = cp
			}
		default:
			return fmt.Errorf("on must be map[string][]Action or nil")
		}
	default:
		return fmt.Errorf("unknown field %q", key)
	}
	return nil
}

// Invert returns the transaction that undoes tx (operations reversed and
// swapped).
func (tx Tx) Invert() Tx {
	out := Tx{Label: "undo " + tx.Label, Ops: make([]Op, 0, len(tx.Ops))}
	for i := len(tx.Ops) - 1; i >= 0; i-- {
		out.Ops = append(out.Ops, tx.Ops[i].Invert())
	}
	return out
}

// Invert returns the operation that undoes op.
func (op Op) Invert() Op {
	switch op.Kind {
	case OpSetProp, OpSetPlace, OpSetField:
		return Op{Kind: op.Kind, Node: op.Node, Key: op.Key, Old: op.New, New: op.Old}
	case OpInsert:
		return Op{Kind: OpRemove, Node: op.Subtree.ID, Parent: op.Parent, Index: op.Index, Subtree: op.Subtree}
	case OpRemove:
		return Op{Kind: OpInsert, Parent: op.Parent, Index: op.Index, Subtree: op.Subtree}
	case OpMove:
		return Op{
			Kind: OpMove, Node: op.Node,
			Parent: op.FromParent, Index: op.FromIndex,
			FromParent: op.Parent, FromIndex: op.Index,
		}
	}
	return op
}

// SetPropOp builds a property operation.
func SetPropOp(id, key string, old, new any) Op {
	return Op{Kind: OpSetProp, Node: id, Key: key, Old: old, New: new}
}

// SetPlaceOp builds a placement operation.
func SetPlaceOp(id string, old, new Place) Op {
	return Op{Kind: OpSetPlace, Node: id, Old: old, New: new}
}

// SetFieldOp builds a node field operation (name, hidden, locked, style, on).
func SetFieldOp(id, key string, old, new any) Op {
	return Op{Kind: OpSetField, Node: id, Key: key, Old: old, New: new}
}

// InsertOp builds an insert operation carrying the subtree.
func InsertOp(parent string, index int, child *Node) Op {
	return Op{Kind: OpInsert, Parent: parent, Index: index, Subtree: child}
}

// RemoveOp builds a remove operation with the subtree captured for undo.
func RemoveOp(d *Document, id string) (Op, error) {
	n, p, i, ok := d.Locate(id)
	if !ok {
		return Op{}, fmt.Errorf("scene: node %q not found", id)
	}
	if p == nil {
		return Op{}, fmt.Errorf("scene: cannot remove the root node")
	}
	return Op{Kind: OpRemove, Node: id, Parent: p.ID, Index: i, Subtree: n.Clone()}, nil
}

// MoveOp builds a move operation; index is the target position after removal.
func MoveOp(d *Document, id, parent string, index int) (Op, error) {
	_, p, i, ok := d.Locate(id)
	if !ok {
		return Op{}, fmt.Errorf("scene: node %q not found", id)
	}
	if p == nil {
		return Op{}, fmt.Errorf("scene: cannot move the root node")
	}
	return Op{
		Kind: OpMove, Node: id, Parent: parent, Index: index,
		FromParent: p.ID, FromIndex: i,
	}, nil
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// History is a bounded undo/redo stack with coalescing. Push applies nothing;
// callers first apply the transaction and then push it.
type History struct {
	past, future []Tx
	limit        int
	savedPos     int
	rev          int // bumped by every push/undo/redo; see Checkpoint
	mergeKey     string
	mergeAt      time.Time
}

// Checkpoint records the current history position so an asynchronous save can
// mark exactly the state it wrote, not whatever the user typed meanwhile.
type Checkpoint struct {
	pos int
	rev int
}

// NewHistory returns a history with the given limit (default 200).
func NewHistory(limit int) *History {
	if limit <= 0 {
		limit = 200
	}
	return &History{limit: limit, savedPos: 0}
}

// Push records an applied transaction. Consecutive transactions with the same
// non-empty MergeKey within one second are merged into one undo step.
func (h *History) Push(tx Tx) {
	if len(tx.Ops) == 0 {
		return
	}
	h.rev++
	if len(h.future) > 0 {
		h.future = nil
		h.savedPos = -1 // the saved state is no longer on this branch
	}
	now := time.Now()
	if tx.MergeKey != "" && len(h.past) > 0 && h.mergeKey == tx.MergeKey && now.Sub(h.mergeAt) < time.Second {
		last := &h.past[len(h.past)-1]
		last.Ops = append(last.Ops, tx.Ops...)
		last.Label = tx.Label
		h.mergeAt = now
		return
	}
	h.past = append(h.past, tx)
	h.mergeKey, h.mergeAt = tx.MergeKey, now
	if len(h.past) > h.limit {
		drop := len(h.past) - h.limit
		h.past = h.past[drop:]
		if h.savedPos >= 0 {
			h.savedPos -= drop
		}
	}
}

// CanUndo reports whether there is a transaction to undo.
func (h *History) CanUndo() bool { return len(h.past) > 0 }

// CanRedo reports whether there is a transaction to redo.
func (h *History) CanRedo() bool { return len(h.future) > 0 }

// Undo applies the inverse of the last transaction.
func (h *History) Undo(d *Document) ([]Change, bool, error) {
	if len(h.past) == 0 {
		return nil, false, nil
	}
	tx := h.past[len(h.past)-1]
	changes, err := d.Apply(tx.Invert())
	if err != nil {
		return nil, false, err
	}
	h.rev++
	h.past = h.past[:len(h.past)-1]
	h.future = append(h.future, tx)
	h.mergeKey = ""
	return changes, true, nil
}

// Redo reapplies the last undone transaction.
func (h *History) Redo(d *Document) ([]Change, bool, error) {
	if len(h.future) == 0 {
		return nil, false, nil
	}
	tx := h.future[len(h.future)-1]
	changes, err := d.Apply(tx)
	if err != nil {
		return nil, false, err
	}
	h.rev++
	h.future = h.future[:len(h.future)-1]
	h.past = append(h.past, tx)
	h.mergeKey = ""
	return changes, true, nil
}

// Dirty reports whether the document differs from the last saved state.
func (h *History) Dirty() bool { return h.savedPos < 0 || len(h.past) != h.savedPos }

// Checkpoint records the position and generation of the current history. Take
// it when a save starts, and pass it to MarkSavedAt when the save completes:
// edits made while the save is in flight then keep the document dirty instead
// of being silently reported as saved.
func (h *History) Checkpoint() Checkpoint {
	// A save boundary must also break coalescing, so keystrokes after the
	// save cannot merge into the transaction that was written.
	h.mergeKey = ""
	return Checkpoint{pos: len(h.past), rev: h.rev}
}

// MarkSavedAt marks the checkpointed state saved. It reports false when the
// history changed since the checkpoint (undo/redo, or a merge into an earlier
// transaction); the caller may then compare the encoded bytes to decide.
func (h *History) MarkSavedAt(cp Checkpoint) bool {
	if cp.rev != h.rev {
		return false
	}
	h.savedPos = cp.pos
	h.mergeKey = ""
	return true
}

// MarkSaved records the current position as saved. It also clears the merge
// key so typing after a save cannot merge into the saved transaction and make
// the document look clean while it changed.
func (h *History) MarkSaved() {
	h.savedPos = len(h.past)
	h.mergeKey = ""
}

// UndoLabel returns the label of the next undo step.
func (h *History) UndoLabel() string {
	if len(h.past) == 0 {
		return ""
	}
	return h.past[len(h.past)-1].Label
}

// RedoLabel returns the label of the next redo step.
func (h *History) RedoLabel() string {
	if len(h.future) == 0 {
		return ""
	}
	return h.future[len(h.future)-1].Label
}
