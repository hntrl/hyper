package runtime

import (
	"fmt"
	"reflect"

	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/runtime/resource"
)

type Process struct {
	Context    *context.Context
	ctxBuilder *context.ContextBuilder
	resources  map[string]resource.Resource
}

func NewProcess() *Process {
	return &Process{
		Context:    nil,
		ctxBuilder: nil,
		resources:  make(map[string]resource.Resource),
	}
}

func (p *Process) UseContextBuilder(bd *context.ContextBuilder) error {
	p.Context = bd.GetContext(bd.RootContext)
	p.ctxBuilder = bd
	return nil
}

func (p *Process) Attach() error {
	for _, key := range p.Context.UsedPackages {
		importedCtx := p.ctxBuilder.GetContext(key)
		if importedCtx != nil {
			for _, obj := range importedCtx.Items {
				if node, ok := obj.(RuntimeNode); ok {
					err := node.Attach(p)
					if err != nil {
						p.Close()
						return err
					}
				}
			}
		}
	}
	for _, obj := range p.Context.Items {
		if node, ok := obj.(RuntimeNode); ok {
			err := node.Attach(p)
			if err != nil {
				p.Close()
				return err
			}
		}
		if res, ok := obj.(RuntimeResource); ok {
			err := res.AttachResource(p)
			if err != nil {
				p.Close()
				return err
			}
		}
	}
	return nil
}
func (p *Process) Close() error {
	for _, ctx := range p.ctxBuilder.Contexts {
		for _, obj := range ctx.Items {
			if node, ok := obj.(RuntimeNode); ok {
				err := node.Detach()
				if err != nil {
					return err
				}
			}
		}
	}
	for _, res := range p.resources {
		res.Detach()
	}
	return nil
}
func (p *Process) Resource(key string, vPtr interface{}) error {
	ptr := reflect.ValueOf(vPtr)
	if ptr.Kind() != reflect.Ptr {
		return fmt.Errorf("%s is not a pointer", key)
	}
	if p.resources[key] == nil {
		if blankRes, ok := reflect.New(ptr.Type().Elem()).Interface().(resource.Resource); ok {
			res, err := blankRes.Attach()
			if err != nil {
				return err
			}
			p.resources[key] = res
		} else {
			return fmt.Errorf("%T is not a resource", vPtr)
		}
	}
	if ptr.Type().Elem() != reflect.TypeOf(p.resources[key]) {
		return fmt.Errorf("resource %s is not of type %s", key, ptr.Type().Elem())
	}
	ptr.Elem().Set(reflect.ValueOf(p.resources[key]))
	return nil
}
