package io

import (
	"context"
	"io"
	"os"

	"github.com/deepnoodle-ai/risor/v2/pkg/object"
	"github.com/deepnoodle-ai/risor/v2/pkg/op"
)

const IOType object.Type = "io"

type IO struct{}

func NewIO() *IO {
	return &IO{}
}

func (i *IO) Type() object.Type {
	return IOType
}

func (i *IO) Inspect() string {
	return "io"
}

func (i *IO) String() string {
	return "io"
}

func (i *IO) Interface() interface{} {
	return i
}

func (i *IO) Equals(other object.Object) bool {
	_, ok := other.(*IO)
	return ok
}

func (i *IO) IsTruthy() bool {
	return true
}

func (i *IO) RunOperation(opType op.BinaryOpType, right object.Object) (object.Object, error) {
	return nil, object.Errorf("eval error: unsupported operation for io: %v", opType)
}

func (i *IO) Cost() int {
	return 0
}

func (i *IO) GetAttr(name string) (object.Object, bool) {
	switch name {
	case "cp":
		return object.NewBuiltin("io.cp", func(ctx context.Context, args ...object.Object) (object.Object, error) {
			if len(args) != 2 {
				return nil, object.NewArgsError("io.cp", 2, len(args))
			}
			src, err := object.AsString(args[0])
			if err != nil {
				return nil, err
			}
			dst, err := object.AsString(args[1])
			if err != nil {
				return nil, err
			}

			// Open source file
			srcFile, err := os.Open(src)
			if err != nil {
				return nil, err
			}
			defer srcFile.Close()

			// Create destination file
			dstFile, err := os.Create(dst)
			if err != nil {
				return nil, err
			}
			defer dstFile.Close()

			// Copy contents
			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return nil, err
			}

			return object.Nil, nil
		}), true
	}
	return nil, false
}

func (i *IO) SetAttr(name string, value object.Object) error {
	return object.Errorf("type error: io object attributes are read-only")
}

func (i *IO) Attrs() []object.AttrSpec {
	return []object.AttrSpec{
		{Name: "cp", Doc: "Copy a file from src to dst"},
	}
}
