package interfaces

import (
	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/domain"
	"github.com/hntrl/hyper/internal/stdlib"
	"github.com/hntrl/hyper/internal/symbols"
	"github.com/hntrl/hyper/internal/symbols/errors"
)

type FileInterface struct{}

func (FileInterface) FromNode(ctx *domain.Context, node ast.ContextObject) (*domain.ContextItem, error) {
	table := ctx.Symbols()
	if node.Private {
		return nil, errors.NodeError(node, 0, "file cannot be private: files aren't exported")
	}
	fileClass := FileClass{
		Name:    node.Name,
		Private: node.Private,
		Comment: node.Comment,
	}
	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.FieldAssignmentExpression:
			switch field.Name {
			case "allowed":
				value, err := table.ResolveExpression(field.Init)
				if err != nil {
					return nil, err
				}
				arrayValue, ok := value.(*symbols.ArrayValue)
				if !ok {
					return nil, errors.NodeError(field.Init, 0, "expected array of type mime.MimeType, got %s", value.Class().Descriptors().Name)
				}
				if arrayValue.ItemClass() != stdlib.MimeType {
					return nil, errors.NodeError(field.Init, 0, "expected array of type mime.MimeType, got %s", arrayValue.ItemClass().Descriptors().Name)
				}
				arrayValueItems := arrayValue.Slice()
				allowed := make([]stdlib.MimeTypeValue, len(arrayValueItems))
				for idx, valueObj := range arrayValueItems {
					allowed[idx] = valueObj.(stdlib.MimeTypeValue)
				}
				fileClass.Allowed = allowed
			default:
				return nil, errors.NodeError(field, 0, "unrecognized assignment %s in file", field.Name)
			}
		default:
			return nil, errors.NodeError(field, 0, "%T not allowed in file", item)
		}
	}
	return &domain.ContextItem{
		HostItem:   fileClass,
		RemoteItem: nil,
	}, nil
}

type FileClass struct {
	Name    string
	Private bool
	Comment string
	Allowed []stdlib.MimeTypeValue
}

func (fc FileClass) Descriptors() *symbols.ClassDescriptors {
	return &symbols.ClassDescriptors{
		Name: fc.Name,
		Constructors: symbols.ClassConstructorSet{
			symbols.Constructor(symbols.String, func(val symbols.StringValue) (FileValue, error) {
				return FileValue{
					fileClass: fc,
					fileID:    string(val),
				}, nil
			}),
		},
		Prototype: symbols.ClassPrototypeMap{
			"identifier": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class:     fc,
				Arguments: []symbols.Class{},
				Returns:   symbols.String,
				Handler: func(fileValue *FileValue) (symbols.StringValue, error) {
					return symbols.StringValue(fileValue.fileID), nil
				},
			}),
		},
	}
}

type FileValue struct {
	fileClass FileClass
	fileID    string `hash:"ignore"`
}

func (fv FileValue) Class() symbols.Class {
	return fv.fileClass
}
func (fv FileValue) Value() interface{} {
	return fv.fileID
}
