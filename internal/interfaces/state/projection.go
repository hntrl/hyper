package state

import (
	"context"
	"fmt"
	"strings"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/domain"
	"github.com/hntrl/hyper/internal/interfaces/stream"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/hntrl/hyper/internal/runtime/log"
	"github.com/hntrl/hyper/internal/runtime/resource"
	"github.com/hntrl/hyper/internal/symbols"
	"github.com/hntrl/hyper/internal/symbols/errors"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

var ProjectionSignal = log.Signal("PROJECTION")
var ProjectionEventSignal = log.Signal("PROJECTION_EVENT")

type ProjectionInterface struct{}

func (ProjectionInterface) FromNode(ctx *domain.Context, node ast.ContextObject) (*domain.ContextItem, error) {
	table := ctx.Symbols()
	proj := Projection{
		Name:       node.Name,
		Private:    node.Private,
		Comment:    node.Comment,
		Properties: make(map[string]symbols.Class),
	}
	if node.Extends != nil {
		extendedType, err := table.ResolveSelector(*node.Extends)
		if err != nil {
			return nil, err
		}
		extendedTypeClass, ok := extendedType.(symbols.Class)
		if !ok {
			return nil, errors.NodeError(node.Extends, 0, "cannot extend %T", extendedType)
		}
		properties := extendedTypeClass.Descriptors().Properties
		if properties == nil {
			return nil, errors.NodeError(node.Extends, 0, "cannot extend %s", extendedTypeClass.Descriptors().Name)
		}
		for k, v := range properties {
			proj.Properties[k] = v.PropertyClass
		}
	}
	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.FieldExpression:
			class, err := table.EvaluateTypeExpression(field.Init)
			if err != nil {
				return nil, err
			}
			proj.Properties[field.Name] = class
		default:
			return nil, errors.NodeError(field, 0, "%T not allowed in projection", item)
		}
	}
	store := &ProjectionStore{
		projectionType: proj,
		events:         make(map[*stream.Event]symbols.Callable),
	}
	if !node.Private {
		return &domain.ContextItem{
			HostItem:   store,
			RemoteItem: proj,
		}, nil
	} else {
		return &domain.ContextItem{
			HostItem:   store,
			RemoteItem: nil,
		}, nil
	}
}

type ProjectionStore struct {
	projectionType Projection
	collection     *mongo.Collection                  `hash:"ignore"`
	events         map[*stream.Event]symbols.Callable `hash:"ignore"`
}

func (ps ProjectionStore) Descriptors() *symbols.ClassDescriptors {
	descriptors := *ps.projectionType.Descriptors()
	projectionRecordType := ProjectionRecord{projectionStore: ps}
	descriptors.ClassProperties = symbols.ClassObjectPropertyMap{
		"find": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.NewPartialClass(ps.projectionType),
				QueryOptions,
			},
			Returns: symbols.NewArrayClass(projectionRecordType),
			Handler: func(filterValue *ProjectionValue, options QueryOptionsValue) (*symbols.ArrayValue, error) {
				dbFindOptions := mongoOptions.Find()
				if options.Skip != -1 {
					dbFindOptions = dbFindOptions.SetSkip(options.Skip)
				}
				if options.Limit != -1 {
					dbFindOptions = dbFindOptions.SetLimit(options.Limit)
				}

				filter := make(bson.M)
				err := flattenObject(filterValue, filter, "")
				if err != nil {
					return nil, err
				}
				cursor, err := ps.collection.Find(context.TODO(), filter, dbFindOptions)
				if err != nil {
					return nil, err
				}

				cursorIndex := 0
				arr := symbols.NewArray(projectionRecordType, cursor.RemainingBatchLength())
				for cursor.Next(context.TODO()) {
					var record bson.M
					err := cursor.Decode(&record)
					if err != nil {
						return nil, err
					}
					bytes, err := bson.MarshalExtJSON(record, false, true)
					if err != nil {
						return nil, err
					}
					recordValue, err := symbols.ValueFromBytes(bytes)
					if err != nil {
						return nil, err
					}
					constructedRecordValue, err := symbols.Construct(ps.projectionType, recordValue)
					if err != nil {
						return nil, err
					}
					recordID := cursor.Current.Lookup("_id").ObjectID()
					arr.Set(cursorIndex, ProjectionRecordValue{
						projectionRecordType: projectionRecordType,
						recordID:             &recordID,
						data:                 constructedRecordValue.(*ProjectionValue).data,
					})
					cursorIndex++
				}
				return arr, nil
			},
		}),
		"findOne": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.NewPartialClass(ps.projectionType),
				QueryOptions,
			},
			Returns: projectionRecordType,
			Handler: func(filterValue *ProjectionValue, options QueryOptionsValue) (*ProjectionRecordValue, error) {
				dbFindOptions := mongoOptions.FindOne()
				if options.Skip != -1 {
					dbFindOptions = dbFindOptions.SetSkip(options.Skip)
				}

				filter := make(bson.M)
				err := flattenObject(filterValue, filter, "")
				if err != nil {
					return nil, err
				}

				var record bson.M
				err = ps.collection.FindOne(context.TODO(), filter, dbFindOptions).Decode(&record)
				if err != nil {
					return nil, err
				}
				bytes, err := bson.MarshalExtJSON(record, false, true)
				if err != nil {
					return nil, err
				}
				recordValue, err := symbols.ValueFromBytes(bytes)
				if err != nil {
					return nil, err
				}
				constructedRecordValue, err := symbols.Construct(ps.projectionType, recordValue)
				if err != nil {
					return nil, err
				}
				recordID := record["_id"].(primitive.ObjectID)
				return &ProjectionRecordValue{
					projectionRecordType: projectionRecordType,
					recordID:             &recordID,
					data:                 constructedRecordValue.(*ProjectionValue).data,
				}, nil
			},
		}),
		"insert": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.NewPartialClass(ps.projectionType),
			},
			Returns: projectionRecordType,
			Handler: func(projectionValue *ProjectionValue) (*ProjectionRecordValue, error) {
				res, err := ps.collection.InsertOne(context.TODO(), projectionValue.Value())
				if err != nil {
					return nil, err
				}
				recordID := res.InsertedID.(primitive.ObjectID)
				return &ProjectionRecordValue{
					projectionRecordType: projectionRecordType,
					recordID:             &recordID,
					data:                 projectionValue.data,
				}, nil
			},
		}),
	}
	return &descriptors
}

func (ps *ProjectionStore) AddMethod(ctx *domain.Context, node ast.ContextObjectMethod) error {
	arguments := node.Block.Parameters.Arguments.Items
	if node.Name != "onEvent" {
		return errors.NodeError(node, 0, "invalid method %s on projection", node.Name)
	}
	if len(arguments) != 1 {
		return errors.NodeError(node, errors.InvalidArgumentLength, "onEvent must only have one argument")
	}
	if node.Block.Parameters.ReturnType != nil {
		return errors.NodeError(node, 0, "onEvent cannot have a return type")
	}
	table := ctx.Symbols()
	if argExpr, ok := arguments[0].(ast.ArgumentItem); ok {
		class, err := table.EvaluateTypeExpression(argExpr.Init)
		if err != nil {
			return err
		}
		if event, ok := class.(stream.Event); ok {
			for registeredEvent := range ps.events {
				if symbols.ClassEquals(event, *registeredEvent) {
					return errors.NodeError(argExpr.Init, 0, "event %s already registered for projection", event.Topic)
				}
			}
			fn, err := table.ResolveFunctionBlock(node.Block)
			if err != nil {
				return err
			}
			ps.events[&event] = fn
			return nil
		}
		return errors.NodeError(argExpr.Init, 0, "argument to onEvent must be an event, got %T", class)
	} else {
		return errors.NodeError(argExpr.Init, 0, "argument to onEvent must be an event, got ambiguous type")
	}
}

func (ps *ProjectionStore) Attach(process *runtime.Process) error {
	var dbConn resource.MongoConnection
	err := process.Resource("mdb", &dbConn)
	if err != nil {
		return err
	}
	var streamConn resource.NatsConnection
	err = process.Resource("stream", &streamConn)
	if err != nil {
		return err
	}

	dbName := strings.Replace(process.Context.Identifier, ".", "_", -1)
	collection, err := dbConn.EnsureCollection(dbName, ps.projectionType.Name)
	if err != nil {
		return err
	}
	ps.collection = collection

	for evPtr, fn := range ps.events {
		ev := *evPtr
		streamConn.Client.QueueSubscribe(string(ev.Topic), "projection_group", func(m *nats.Msg) {
			value, err := symbols.ValueFromBytes(m.Data)
			if err != nil {
				return
			}
			constructedValue, err := symbols.Construct(ev, value)
			if err != nil {
				return
			}
			_, err = fn.Call(constructedValue)
			if err != nil {
				return
			}
		})
	}
	return nil
}
func (ps *ProjectionStore) Detach(process *runtime.Process) error {
	ps.collection = nil
	return nil
}

type Projection struct {
	Name       string
	Private    bool
	Comment    string
	Properties map[string]symbols.Class
}

func (p Projection) Descriptors() *symbols.ClassDescriptors {
	propertyMap := make(symbols.ClassPropertyMap)
	for name, class := range p.Properties {
		propertyMap[name] = symbols.PropertyAttributes(symbols.PropertyOptions{
			Class: class,
			Getter: func(val *EntityValue) (symbols.ValueObject, error) {
				return val.data[name], nil
			},
			Setter: func(val *EntityValue, newPropertyValue symbols.ValueObject) error {
				val.data[name] = newPropertyValue
				return nil
			},
		})
	}
	projectionStore := ProjectionStore{projectionType: p}
	projectionRecordClass := ProjectionRecord{projectionStore: projectionStore}
	return &symbols.ClassDescriptors{
		Name: p.Name,
		Constructors: symbols.ClassConstructorSet{
			symbols.Constructor(symbols.Map, func(val *symbols.MapValue) (*ProjectionValue, error) {
				return &ProjectionValue{
					projectionType: p,
					data:           val.Map(),
				}, nil
			}),
			symbols.Constructor(projectionRecordClass, func(val *ProjectionRecordValue) (*ProjectionValue, error) {
				return &ProjectionValue{
					projectionType: p,
					data:           val.data,
				}, nil
			}),
		},
		Properties: propertyMap,
	}
}

type ProjectionValue struct {
	projectionType Projection
	data           map[string]symbols.ValueObject
}

func (p ProjectionValue) Class() symbols.Class {
	return p.projectionType
}
func (p ProjectionValue) Value() interface{} {
	out := make(map[string]interface{})
	for k, v := range p.data {
		out[k] = v.Value()
	}
	return out
}

type ProjectionRecord struct {
	projectionStore ProjectionStore
}

func (pr ProjectionRecord) Descriptors() *symbols.ClassDescriptors {
	propertyMap := make(symbols.ClassPropertyMap)
	for name, class := range pr.projectionStore.projectionType.Properties {
		propertyMap[name] = symbols.PropertyAttributes(symbols.PropertyOptions{
			Class: class,
			Getter: func(obj *EntityInstanceValue) (symbols.ValueObject, error) {
				return obj.data[name], nil
			},
		})
	}
	return &symbols.ClassDescriptors{
		Name: fmt.Sprintf("[%s]", pr.projectionStore.projectionType.Name),
		Prototype: symbols.ClassPrototypeMap{
			"update": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class:     pr,
				Arguments: []symbols.Class{pr.projectionStore.projectionType},
				Returns:   nil,
				Handler: func(projectionValue *ProjectionRecordValue, updateValue *ProjectionValue) error {
					_, err := pr.projectionStore.collection.UpdateOne(context.TODO(), bson.M{"_id": projectionValue.recordID}, updateValue.Value())
					projectionValue.data = updateValue.data
					return err
				},
			}),
			"delete": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class:     pr,
				Arguments: []symbols.Class{},
				Returns:   nil,
				Handler: func(projectionValue *ProjectionRecordValue) error {
					_, err := pr.projectionStore.collection.DeleteOne(context.TODO(), bson.M{"_id": projectionValue.recordID})
					return err
				},
			}),
			"mutable": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class:     pr,
				Arguments: []symbols.Class{},
				Returns:   pr.projectionStore.projectionType,
				Handler: func(projectionValue *ProjectionRecordValue) (*ProjectionValue, error) {
					return &ProjectionValue{
						projectionType: pr.projectionStore.projectionType,
						data:           projectionValue.data,
					}, nil
				},
			}),
		},
		Properties: propertyMap,
	}
}

type ProjectionRecordValue struct {
	projectionRecordType ProjectionRecord
	recordID             *primitive.ObjectID `bson:"_id"`
	data                 map[string]symbols.ValueObject
}

func (p ProjectionRecordValue) Class() symbols.Class {
	return p.projectionRecordType
}
func (p ProjectionRecordValue) Value() interface{} {
	out := make(map[string]interface{})
	for k, v := range p.data {
		out[k] = v.Value()
	}
	out["$_id"] = p.recordID.String()
	return out
}
