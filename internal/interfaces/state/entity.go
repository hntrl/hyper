package state

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/domain"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/hntrl/hyper/internal/runtime/log"
	"github.com/hntrl/hyper/internal/runtime/resource"
	"github.com/hntrl/hyper/internal/symbols"
	"github.com/hntrl/hyper/internal/symbols/errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

var EntityStreamSignal = log.Signal("ENTITY_STREAM")
var EntityMethodSignal = log.Signal("ENTITY_METHOD")
var EntityCollectionSignal = log.Signal("ENTITY_COLLECTION")

type EntityInterface struct{}

func (EntityInterface) FromNode(ctx *domain.Context, node ast.ContextObject) (*domain.ContextItem, error) {
	table := ctx.Symbols()
	ent := Entity{
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
			ent.Properties[k] = v.PropertyClass
		}
	}
	for _, item := range node.Fields {
		switch field := item.Init.(type) {
		case ast.FieldExpression:
			class, err := table.EvaluateTypeExpression(field.Init)
			if err != nil {
				return nil, err
			}
			ent.Properties[field.Name] = class
		default:
			return nil, errors.NodeError(field, 0, "%T not allowed in entity", item)
		}
	}
	store := EntityStore{
		entityType: ent,
		methods:    make(map[EffectType]symbols.Function),
	}
	if !node.Private {
		return &domain.ContextItem{
			HostItem:   store,
			RemoteItem: ent,
		}, nil
	} else {
		return &domain.ContextItem{
			HostItem:   store,
			RemoteItem: nil,
		}, nil
	}
}

// EntityStore is synonymous to Entity (it uses the same value type
// EntityValue). Only difference is this acts as the connection to the database
// for the host context.
type EntityStore struct {
	entityType   Entity
	eventLog     *mongo.Collection               `hash:"ignore"`
	projection   *mongo.Collection               `hash:"ignore"`
	cancelStream context.CancelFunc              `hash:"ignore"`
	methods      map[EffectType]symbols.Function `hash:"ignore"`
}

func (es EntityStore) Descriptors() *symbols.ClassDescriptors {
	descriptors := *es.entityType.Descriptors()
	descriptors.ClassProperties = symbols.ClassObjectPropertyMap{
		"find": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				es.entityType,
				QueryOptions,
			},
			Returns: symbols.NewArrayClass(EntityInstance{entityFactory: es}),
			Handler: func(filterValue *EntityValue, options QueryOptionsValue) (*symbols.ArrayValue, error) {
				dbFindOptions := mongoOptions.Find()
				if options.Skip != -1 {
					dbFindOptions = dbFindOptions.SetSkip(options.Skip)
				}
				if options.Limit != -1 {
					dbFindOptions = dbFindOptions.SetLimit(options.Limit)
				}
				filter := make(bson.M)
				err := flattenObject(filterValue, filter, "state.")
				if err != nil {
					return nil, err
				}
				cursor, err := es.projection.Find(context.TODO(), filter, dbFindOptions)
				if err != nil {
					return nil, err
				}
				var results []EntityState
				if err = cursor.All(context.TODO(), &results); err != nil {
					return nil, err
				}
				instanceType := EntityInstance{entityFactory: es}
				arr := symbols.NewArray(instanceType, len(results))
				for idx, item := range results {
					instanceValue, err := item.EntityInstance(es)
					if err != nil {
						return nil, err
					}
					arr.Set(idx, instanceValue)
				}
				return arr, nil
			},
		}),
		"findOne": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				es.entityType,
				QueryOptions,
			},
			Returns: EntityInstance{entityFactory: es},
			Handler: func(filterValue *EntityValue, options QueryOptionsValue) (*EntityInstanceValue, error) {
				dbFindOptions := mongoOptions.FindOne()
				if options.Skip != -1 {
					dbFindOptions = dbFindOptions.SetSkip(options.Skip)
				}
				filter := make(bson.M)
				err := flattenObject(filterValue, filter, "state.")
				if err != nil {
					return nil, err
				}
				var result EntityState
				err = es.projection.FindOne(context.TODO(), filter, dbFindOptions).Decode(&result)
				if err != nil {
					if err == mongo.ErrNoDocuments {
						return nil, symbols.ErrorValue{
							Name:    "NotFound",
							Message: "no matching entities",
						}
					}
					return nil, err
				}
				return result.EntityInstance(es)
			},
		}),
		"insert": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				es.entityType,
			},
			Returns: EntityInstance{entityFactory: es},
			Handler: func(stateValue *EntityValue) (*EntityInstanceValue, error) {
				state := EntityStateEvent{
					EntityID:  strconv.Itoa(seededRand.Int())[0:12],
					Timestamp: time.Now(),
					Effect:    EffectTypeCreate,
					State:     stateValue.Value(),
				}
				if _, err := es.eventLog.InsertOne(context.TODO(), state); err != nil {
					return nil, err
				}
				return state.EntityInstance(es)
			},
		}),
	}
	return &descriptors
}

// This acts as the updater between the event log and the projection.
func (es EntityStore) iterateChangeStream(routineCtx context.Context, stream *mongo.ChangeStream) {
	defer stream.Close(routineCtx)
	for stream.Next(routineCtx) {
		var event EntityStateInsertEvent
		if err := stream.Decode(&event); err != nil {
			panic(err)
		}
		switch event.FullDocument.Effect {
		case EffectTypeCreate:
			// Create a new working record
			state := EntityState{
				EntityID:  event.FullDocument.EntityID,
				CreatedAt: event.FullDocument.Timestamp,
				UpdatedAt: event.FullDocument.Timestamp,
				State:     event.FullDocument.State,
			}
			_, err := es.projection.InsertOne(routineCtx, state)
			if err != nil {
				panic(err)
			}
			if fn, ok := es.methods[EffectTypeCreate]; ok {
				value, err := state.EntityInstance(es)
				if err != nil {
					panic(err)
				}
				_, err = fn.Call(value)
				if err != nil {
					panic(err)
				}
			}
		case EffectTypeUpdate:
			// Update the working record
			filter := EntityState{EntityID: event.FullDocument.EntityID}
			updateMethod, hasUpdateMethod := es.methods[EffectTypeUpdate]
			var currentState EntityState
			if hasUpdateMethod {
				err := es.projection.FindOne(routineCtx, filter).Decode(&currentState)
				if err != nil {
					panic(err)
				}
			}
			_, err := es.projection.UpdateOne(routineCtx, filter, bson.M{
				"$set": EntityState{
					UpdatedAt: event.FullDocument.Timestamp,
					State:     event.FullDocument.State,
				},
			})
			if err != nil {
				panic(err)
			}
			if hasUpdateMethod {
				currentInstance, err := currentState.EntityInstance(es)
				if err != nil {
					panic(err)
				}
				newInstance, err := event.FullDocument.EntityInstance(es)
				if err != nil {
					panic(err)
				}
				_, err = updateMethod.Call(currentInstance, newInstance)
				if err != nil {
					panic(err)
				}
			}
		case EffectTypeDelete:
			// Delete the working record
			_, err := es.projection.DeleteOne(routineCtx, EntityState{
				EntityID: event.FullDocument.EntityID,
			})
			if err != nil {
				panic(err)
			}
			if fn, ok := es.methods[EffectTypeDelete]; ok {
				value, err := event.FullDocument.EntityInstance(es)
				if err != nil {
					panic(err)
				}
				_, err = fn.Call(value)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func (es *EntityStore) Attach(process *runtime.Process) error {
	var conn resource.MongoConnection
	err := process.Resource("mdb", &conn)
	if err != nil {
		return err
	}
	dbName := strings.Replace(process.Context.Identifier, ".", "_", -1)
	es.eventLog, err = conn.EnsureCollection(dbName, fmt.Sprintf("%s_events", es.entityType.Name))
	if err != nil {
		return err
	}
	es.projection, err = conn.EnsureCollection(dbName, fmt.Sprintf("%s_projection", es.entityType.Name))
	if err != nil {
		return err
	}
	pipeline := mongo.Pipeline{
		{{
			Key:   "$match",
			Value: bson.D{{Key: "operationType", Value: "insert"}},
		}},
	}
	streamOptions := mongoOptions.ChangeStream().SetFullDocument(mongoOptions.UpdateLookup)
	stream, err := es.eventLog.Watch(context.Background(), pipeline, streamOptions)
	if err != nil {
		return err
	}
	routineCtx, cancel := context.WithCancel(context.Background())
	es.cancelStream = cancel
	go es.iterateChangeStream(routineCtx, stream)
	return nil
}
func (es *EntityStore) Detach() error {
	es.cancelStream()
	es.eventLog = nil
	es.projection = nil
	return nil
}

type Entity struct {
	Name       string
	Private    bool
	Comment    string
	Properties map[string]symbols.Class
}

func (ent Entity) Descriptors() *symbols.ClassDescriptors {
	propertyMap := make(symbols.ClassPropertyMap)
	for name, class := range ent.Properties {
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
	return &symbols.ClassDescriptors{
		Name:       ent.Name,
		Properties: propertyMap,
	}
}

type EntityValue struct {
	entityType Entity
	data       map[string]symbols.ValueObject
}

func (eo EntityValue) Class() symbols.Class {
	return eo.entityType
}
func (eo EntityValue) Value() interface{} {
	out := make(map[string]interface{})
	for k, v := range eo.data {
		out[k] = v.Value()
	}
	return out
}

type EntityInstance struct {
	entityFactory EntityStore
}

func (ei EntityInstance) Descriptors() *symbols.ClassDescriptors {
	propertyMap := make(symbols.ClassPropertyMap)
	for name, class := range ei.entityFactory.entityType.Properties {
		propertyMap[name] = symbols.PropertyAttributes(symbols.PropertyOptions{
			Class: class,
			Getter: func(obj *EntityInstanceValue) (symbols.ValueObject, error) {
				return obj.data[name], nil
			},
		})
	}
	return &symbols.ClassDescriptors{
		Name: fmt.Sprintf("[%s]", ei.entityFactory.entityType.Name),
		Prototype: symbols.ClassPrototypeMap{
			"update": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class: ei,
				Arguments: []symbols.Class{
					ei.entityFactory.entityType,
				},
				Returns: nil,
				Handler: func(instanceValue *EntityInstanceValue, updatedValue EntityValue) error {
					state := EntityStateEvent{
						EntityID:  instanceValue.entityID,
						Timestamp: time.Now(),
						Effect:    EffectTypeUpdate,
						State:     updatedValue,
					}
					if _, err := instanceValue.instanceType.entityFactory.eventLog.InsertOne(context.TODO(), state); err != nil {
						return err
					}
					instanceValue.data = updatedValue.data
					return nil
				},
			}),
			"delete": symbols.NewClassMethod(symbols.ClassMethodOptions{
				Class:     ei,
				Arguments: []symbols.Class{},
				Returns:   nil,
				Handler: func(instanceValue *EntityInstanceValue) error {
					state := EntityStateEvent{
						EntityID:  instanceValue.entityID,
						Timestamp: time.Now(),
						Effect:    EffectTypeDelete,
					}
					if _, err := instanceValue.instanceType.entityFactory.eventLog.InsertOne(context.TODO(), state); err != nil {
						return err
					}
					return nil
				},
			}),
		},
		Properties: propertyMap,
	}
}

type EntityInstanceValue struct {
	instanceType EntityInstance
	entityID     string
	data         map[string]symbols.ValueObject
}

func (eio EntityInstanceValue) Class() symbols.Class {
	return eio.instanceType
}
func (eio EntityInstanceValue) Value() interface{} {
	out := make(map[string]interface{})
	for k, v := range eio.data {
		out[k] = v.Value()
	}
	out["$entity"] = eio.entityID
	return out
}

// Represents the different operations that can be performed on an entity
type EffectType string

const (
	EffectTypeCreate EffectType = "CREATE"
	EffectTypeUpdate            = "UPDATE"
	EffectTypeDelete            = "DELETE"
)

func unmarshalEntityToInstanceValue(entityFactory EntityStore, entityID string, state interface{}) (*EntityInstanceValue, error) {
	instanceType := EntityInstance{entityFactory: entityFactory}
	bytes, err := bson.MarshalExtJSON(state, false, true)
	if err != nil {
		return nil, err
	}
	stateValue, err := symbols.ValueFromBytes(bytes)
	if err != nil {
		return nil, err
	}
	constructedStateValue, err := symbols.Construct(instanceType, stateValue)
	if err != nil {
		return nil, err
	}
	return &EntityInstanceValue{
		instanceType: instanceType,
		entityID:     entityID,
		data:         constructedStateValue.(EntityValue).data,
	}, nil
}

// Represents the internal model of the entity event log
type EntityStateEvent struct {
	RecordID  *primitive.ObjectID `bson:"_id,omitempty"`
	EntityID  string              `bson:"entity_id,omitempty"`
	Timestamp time.Time           `bson:"timestamp,omitempty"`
	Effect    EffectType          `bson:"esfect,omitempty"`
	State     interface{}         `bson:"state,omitempty"`
}

func (es EntityStateEvent) EntityInstance(entityFactory EntityStore) (*EntityInstanceValue, error) {
	return unmarshalEntityToInstanceValue(entityFactory, es.EntityID, es.State)
}

// Represents the internal model of the working record
type EntityState struct {
	RecordID  *primitive.ObjectID `bson:"_id,omitempty"`
	EntityID  string              `bson:"entity_id,omitempty"`
	CreatedAt time.Time           `bson:"created_at,omitempty"`
	UpdatedAt time.Time           `bson:"updated_at,omitempty"`
	State     interface{}         `bson:"state,omitempty"`
}

func (es EntityState) EntityInstance(entityFactory EntityStore) (*EntityInstanceValue, error) {
	return unmarshalEntityToInstanceValue(entityFactory, es.EntityID, es.State)
}

// Represents the event given from the database change stream when an entity state event is inserted
type EntityStateInsertEvent struct {
	FullDocument EntityStateEvent `bson:"fullDocument"`
}
