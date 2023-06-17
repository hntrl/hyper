package runtime

import (
	"fmt"
	"reflect"

	"github.com/hntrl/hyper/internal/domain"
	"github.com/hntrl/hyper/internal/runtime/resource"
)

type Process struct {
	Context          *domain.Context
	ctxBuilder       *domain.ContextBuilder
	initializedNodes []RuntimeNode
	resources        map[string]resource.Resource
}

func NewProcess() *Process {
	return &Process{
		Context:          nil,
		ctxBuilder:       nil,
		initializedNodes: nil,
		resources:        make(map[string]resource.Resource),
	}
}

func (p *Process) UseContextBuilder(bd *domain.ContextBuilder) error {
	p.Context = bd.HostContext()
	p.ctxBuilder = bd
	return nil
}

func (p *Process) Attach() error {
	p.initializedNodes = make([]RuntimeNode, 0)
	// Initialize Host Context Resources
	for _, item := range p.Context.Items {
		if hostRuntimeNode, ok := item.HostItem.(RuntimeNode); ok {
			err := hostRuntimeNode.Attach(p)
			if err != nil {
				p.Close()
				return err
			}
			p.initializedNodes = append(p.initializedNodes, hostRuntimeNode)
		}
	}
	// Initialize Imported Context Resources
	for _, ctxPath := range p.Context.ImportedContexts {
		ctx := p.ctxBuilder.GetContextByPath(string(ctxPath))
		for _, item := range ctx.Items {
			if remoteRuntimeNode, ok := item.RemoteItem.(RuntimeNode); ok {
				err := remoteRuntimeNode.Attach(p)
				if err != nil {
					p.Close()
					return err
				}
				p.initializedNodes = append(p.initializedNodes, remoteRuntimeNode)
			}
		}
	}
	return nil
}
func (p *Process) Close() error {
	for _, node := range p.initializedNodes {
		err := node.Detach()
		if err != nil {
			return err
		}
	}
	p.initializedNodes = nil
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
