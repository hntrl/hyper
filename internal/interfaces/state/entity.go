package state

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/hntrl/hyper/internal/ast"
	langCtx "github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/hntrl/hyper/internal/runtime/log"
	"github.com/hntrl/hyper/internal/runtime/resource"
	"github.com/hntrl/hyper/internal/symbols"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type Entity struct {
	ParentType    symbols.Type
	eventLog      *mongo.Collection               `hash:"ignore"`
	workingRecord *mongo.Collection               `hash:"ignore"`
	cancelStream  context.CancelFunc              `hash:"ignore"`
	methods       map[effectType]symbols.Function `hash:"ignore"`
}

var EntityStreamSignal = log.Signal("ENTITY_STREAM")
var EntityMethodSignal = log.Signal("ENTITY_METHOD")
var EntityCollectionSignal = log.Signal("ENTITY_COLLECTION")

func (ent Entity) ClassName() string {
	return ent.ParentType.Name
}
func (ent Entity) Fields() map[string]symbols.Class {
	return ent.ParentType.Fields()
}
func (ent Entity) Constructors() symbols.ConstructorMap {
	csMap := symbols.NewConstructorMap()
	csMap.AddConstructor(EntityInstance{ParentType: ent}, func(obj symbols.ValueObject) (symbols.ValueObject, error) {
		inst := obj.(EntityInstance)
		return &EntityType{ent, inst.fields}, nil
	})
	csMap.AddGenericConstructor(ent, func(fields map[string]symbols.ValueObject) (symbols.ValueObject, error) {
		return &EntityType{ent, fields}, nil
	})
	return csMap
}
func (ent Entity) Get(key string) (symbols.Object, error) {
	methods := map[string]symbols.Object{
		"find": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.NewPartialObject(ent),
				QueryOpts{},
			},
			Returns: symbols.Iterable{
				ParentType: EntityInstance{ParentType: ent},
				Items:      []symbols.ValueObject{},
			},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				if ent.eventLog == nil {
					return nil, fmt.Errorf("db connection not initialized")
				}
				fo := options.Find()
				qo := args[1].(*QueryOpts)
				if qo.Skip != nil {
					val := qo.Skip.Value()
					fo = fo.SetSkip(int64(val.(int)))
				}
				if qo.Limit != nil {
					val := qo.Limit.Value()
					fo = fo.SetLimit(int64(val.(int)))
				}

				filterObject := args[0]
				filter := make(bson.M)
				err := flattenObject(filterObject, filter, "state.")
				if err != nil {
					return nil, err
				}

				cursor, err := ent.workingRecord.Find(context.TODO(), filter, fo)
				if err != nil {
					return nil, err
				}

				var results []entityState
				if err = cursor.All(context.TODO(), &results); err != nil {
					return nil, err
				}

				iter := symbols.NewIterable(EntityInstance{ParentType: ent}, len(results))
				for idx, result := range results {
					bytes, err := bson.MarshalExtJSON(result.State, false, true)
					if err != nil {
						return nil, err
					}
					genericObject, err := symbols.FromBytes(bytes)
					if err != nil {
						return nil, err
					}
					obj, err := symbols.Construct(ent, genericObject)
					if err != nil {
						return nil, err
					}
					if entType, ok := obj.(*EntityType); ok {
						iter.Items[idx] = EntityInstance{
							ParentType: ent,
							entId:      result.EntityID,
							fields:     entType.Fields,
						}
					} else {
						return nil, fmt.Errorf("expected entity type")
					}
				}
				return iter, nil
			},
		}),
		"findOne": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.NewPartialObject(ent),
				QueryOpts{},
			},
			Returns: EntityInstance{ParentType: ent},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				if ent.eventLog == nil {
					return nil, fmt.Errorf("db connection not initialized")
				}
				fo := options.FindOne()
				qo := args[1].(*QueryOpts)
				if qo.Skip != nil {
					val := qo.Skip.Value()
					fo = fo.SetSkip(val.(int64))
				}

				filterObject := args[0]
				filter := make(bson.M)
				err := flattenObject(filterObject, filter, "state.")
				if err != nil {
					return nil, err
				}

				var result entityState
				err = ent.workingRecord.FindOne(context.TODO(), filter, fo).Decode(&result)
				if err != nil {
					if err == mongo.ErrNoDocuments {
						return nil, symbols.Error{
							Name:    "NotFound",
							Message: "no matching entities",
						}
					}
					return nil, err
				}

				bytes, err := bson.MarshalExtJSON(result.State, false, true)
				if err != nil {
					return nil, err
				}
				genericObject, err := symbols.FromBytes(bytes)
				if err != nil {
					return nil, err
				}
				obj, err := symbols.Construct(ent, genericObject)
				if err != nil {
					return nil, fmt.Errorf("cannot construct entity from db record: %s", err.Error())
				}

				if insObj, ok := obj.(*EntityType); ok {
					return EntityInstance{
						ParentType: ent,
						entId:      result.EntityID,
						fields:     insObj.Fields,
					}, nil
				} else {
					return nil, fmt.Errorf("expected entity type")
				}
			},
		}),
		"insert": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				ent,
			},
			Returns: EntityInstance{ParentType: ent},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				if ent.eventLog == nil {
					return nil, fmt.Errorf("db connection not initialized")
				}
				passedEnt, ok := args[0].(*EntityType)
				if !ok {
					return nil, fmt.Errorf("expected entity type")
				}
				state := stateEvent{
					EntityID:  strconv.Itoa(seededRand.Int())[0:12],
					Timestamp: time.Now(),
					Effect:    effectTypeCreate,
					State:     passedEnt.Value(),
				}
				if _, err := ent.eventLog.InsertOne(context.TODO(), state); err != nil {
					return nil, err
				}
				return EntityInstance{ParentType: ent, entId: state.EntityID, fields: passedEnt.Fields}, nil
			},
		}),
	}
	return methods[key], nil
}

func (ent Entity) ObjectClassFromNode(ctx *langCtx.Context, node ast.ContextObject) (symbols.Class, error) {
	assumedType, err := (langCtx.TypeInterface{}).ObjectClassFromNode(ctx, node)
	if err != nil {
		return nil, err
	}
	if typeClass, ok := assumedType.(symbols.Type); ok {
		return &Entity{
			ParentType: typeClass,
			methods:    make(map[effectType]symbols.Function),
		}, nil
	} else {
		return nil, fmt.Errorf("expected type class")
	}
}

func (ent *Entity) AddMethod(ctx *langCtx.Context, node ast.ContextObjectMethod) error {
	var targetEffect effectType
	switch node.Name {
	case "onCreate":
		targetEffect = effectTypeCreate
		if len(node.Block.Parameters.Arguments.Items) != 1 {
			return symbols.NodeError(node.Block.Parameters.Arguments, "onCreate method must have one argument")
		}
	case "onUpdate":
		targetEffect = effectTypeUpdate
		if len(node.Block.Parameters.Arguments.Items) != 2 {
			return symbols.NodeError(node.Block.Parameters.Arguments, "onUpdate method must have two arguments")
		}
	case "onDelete":
		targetEffect = effectTypeDelete
		if len(node.Block.Parameters.Arguments.Items) != 1 {
			return symbols.NodeError(node.Block.Parameters.Arguments, "onDelete method must have one argument")
		}
	default:
		return symbols.NodeError(node, "method %s not allowed on %s", node.Name, ent.ParentType.Name)
	}
	if _, ok := ent.methods[targetEffect]; ok {
		return symbols.NodeError(node, "method %s already defined for %s", node.Name, ent.ParentType.Name)
	}
	if node.Block.Parameters.ReturnType != nil {
		return symbols.NodeError(node, "method %s must not have a return type", node.Name)
	}
	table := ctx.Symbols()
	for _, arg := range node.Block.Parameters.Arguments.Items {
		if argExpr, ok := arg.(ast.ArgumentItem); ok {
			class, err := table.ResolveTypeExpression(argExpr.Init)
			if err != nil {
				return err
			}
			if !symbols.ClassEquals(class, ent) {
				return symbols.NodeError(argExpr, "method %s argument must be equal to target type", node.Name)
			}
		} else {
			return symbols.NodeError(argExpr, "method %s argument must be equal to target type", node.Name)
		}
	}
	fn, err := table.ResolveFunctionBlock(node.Block, &symbols.MapObject{})
	if err != nil {
		return err
	}
	ent.methods[targetEffect] = *fn
	return nil
}

func (ent Entity) Export() (symbols.Object, error) {
	return ent.ParentType, nil
}

func (ent Entity) iterateChangeStream(routineCtx context.Context, stream *mongo.ChangeStream) {
	defer stream.Close(routineCtx)
	for stream.Next(routineCtx) {
		var event updateEvent
		if err := stream.Decode(&event); err != nil {
			log.Output(log.LoggerMessage{
				LogLevel: log.LevelERROR,
				Signal:   EntityStreamSignal,
				Message:  err.Error(),
				Data:     event,
			})
			continue
		}

		reportDatabaseError := func(err error, effect effectType) {
			log.Output(log.LoggerMessage{
				LogLevel: log.LevelERROR,
				Signal:   EntityCollectionSignal,
				Message:  "database operation encountered an error",
				Data: log.LogData{
					"err":        err.Error(),
					"db":         ent.workingRecord.Name(),
					"collection": ent.workingRecord.Database().Name(),
					"effect":     effect,
				},
			})
		}
		reportMethodError := func(err error, method effectType) {
			log.Output(log.LoggerMessage{
				LogLevel: log.LevelERROR,
				Signal:   EntityMethodSignal,
				Message:  "entity method handler failed invocation",
				Data: log.LogData{
					"err":    err.Error(),
					"entity": ent.ParentType.Name,
					"method": method,
				},
			})
		}
		reportStreamError := func(err error, method effectType) {
			log.Output(log.LoggerMessage{
				LogLevel: log.LevelERROR,
				Signal:   EntityMethodSignal,
				Message:  "entity stream failed to marshal object",
				Data: log.LogData{
					"err":    err.Error(),
					"entity": ent.ParentType.Name,
					"method": method,
				},
			})
		}

		switch event.FullDocument.Effect {
		case effectTypeCreate:
			_, err := ent.workingRecord.InsertOne(routineCtx, entityState{
				EntityID:  event.FullDocument.EntityID,
				CreatedAt: event.FullDocument.Timestamp,
				UpdatedAt: event.FullDocument.Timestamp,
				State:     event.FullDocument.State,
			})
			if err != nil {
				reportDatabaseError(err, effectTypeCreate)
				continue
			}
			if fn, ok := ent.methods[effectTypeCreate]; ok {
				obj, err := event.FullDocument.EntityInstance(ent)
				if err != nil {
					reportStreamError(err, effectTypeCreate)
					continue
				}
				_, err = fn.Call([]symbols.ValueObject{obj}, &symbols.MapObject{})
				if err != nil {
					reportMethodError(err, effectTypeCreate)
					continue
				}
			}
		case effectTypeUpdate:
			filter := entityState{EntityID: event.FullDocument.EntityID}

			var oldState entityState
			err := ent.workingRecord.FindOne(routineCtx, filter).Decode(&oldState)
			if err != nil {
				reportDatabaseError(err, effectTypeUpdate)
				continue
			}

			var newState entityState
			err = ent.workingRecord.FindOneAndUpdate(routineCtx, filter, bson.M{
				"$set": entityState{
					UpdatedAt: event.FullDocument.Timestamp,
					State:     event.FullDocument.State,
				},
			}).Decode(&newState)
			if err != nil {
				reportDatabaseError(err, effectTypeUpdate)
				continue
			}

			if fn, ok := ent.methods[effectTypeUpdate]; ok {
				old, err := oldState.EntityInstance(ent)
				if err != nil {
					reportStreamError(err, effectTypeUpdate)
					continue
				}
				new, err := newState.EntityInstance(ent)
				if err != nil {
					reportStreamError(err, effectTypeUpdate)
					continue
				}
				_, err = fn.Call([]symbols.ValueObject{old, new}, &symbols.MapObject{})
				if err != nil {
					reportMethodError(err, effectTypeUpdate)
					continue
				}
			}
		case effectTypeDelete:
			res := ent.workingRecord.FindOneAndDelete(routineCtx, entityState{EntityID: event.FullDocument.EntityID})
			if res.Err() != nil {
				reportDatabaseError(res.Err(), effectTypeDelete)
				continue
			}
			if fn, ok := ent.methods[effectTypeDelete]; ok {
				var entityState entityState
				if err := res.Decode(&entityState); err != nil {
					reportStreamError(err, effectTypeDelete)
					continue
				}
				obj, err := entityState.EntityInstance(ent)
				if err != nil {
					reportStreamError(err, effectTypeDelete)
					continue
				}
				_, err = fn.Call([]symbols.ValueObject{obj}, &symbols.MapObject{})
				if err != nil {
					reportMethodError(err, effectTypeDelete)
					continue
				}
			}
		}
	}
}

func (ent *Entity) Attach(process *runtime.Process) error {
	return nil
}
func (ent *Entity) AttachResource(process *runtime.Process) error {
	var conn resource.MongoConnection
	err := process.Resource("mdb", &conn)
	if err != nil {
		return err
	}
	dbName := strings.Replace(process.Context.Name, ".", "_", -1)
	workingRecord, err := conn.EnsureCollection(dbName, ent.ParentType.Name)
	if err != nil {
		return err
	}
	eventLog, err := conn.EnsureCollection(dbName, fmt.Sprintf("%s_EventLog", ent.ParentType.Name))
	if err != nil {
		return err
	}

	log.Printf(
		log.LevelDEBUG,
		EntityCollectionSignal,
		"entity collections \"%s:%s\" created",
		dbName,
		ent.ParentType.Name,
	)

	ent.workingRecord = workingRecord
	ent.eventLog = eventLog

	pipeline := bson.D{{
		Key: "$match",
		Value: bson.D{{
			Key:   "operationType",
			Value: "insert",
		}},
	}}

	csOpts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
	stream, err := ent.eventLog.Watch(context.TODO(), mongo.Pipeline{pipeline}, csOpts)
	if err != nil {
		return err
	}

	routineCtx, cancel := context.WithCancel(context.Background())
	go ent.iterateChangeStream(routineCtx, stream)
	ent.cancelStream = cancel

	log.Printf(
		log.LevelDEBUG,
		EntityStreamSignal,
		"entity stream \"%s:%s\" started",
		dbName,
		eventLog.Name(),
	)

	return nil
}
func (ent *Entity) Detach() error {
	if ent.cancelStream != nil {
		ent.cancelStream()
	}
	return nil
}

type EntityType struct {
	ParentType Entity
	Fields     map[string]symbols.ValueObject `hash:"ignore"`
}

func (tp EntityType) Class() symbols.Class {
	return tp.ParentType
}
func (tp EntityType) Value() interface{} {
	out := make(map[string]interface{})
	for k, v := range tp.Fields {
		out[k] = v.Value()
	}
	return out
}

func (tp *EntityType) Set(key string, obj symbols.ValueObject) error {
	tp.Fields[key] = obj
	return nil
}
func (tp EntityType) Get(key string) (symbols.Object, error) {
	return tp.Fields[key], nil
}

type EntityInstance struct {
	ParentType Entity
	entId      string                         `hash:"ignore"`
	fields     map[string]symbols.ValueObject `hash:"ignore"`
}

func (ins EntityInstance) ClassName() string {
	return fmt.Sprintf("[%s]", ins.ParentType.ClassName())
}
func (ins EntityInstance) Fields() map[string]symbols.Class {
	return ins.ParentType.Fields()
}
func (ins EntityInstance) Constructors() symbols.ConstructorMap {
	return symbols.NewConstructorMap()
}

func (ins EntityInstance) Class() symbols.Class {
	return ins
}
func (ins EntityInstance) Value() interface{} {
	out := make(map[string]interface{})
	for k, v := range ins.fields {
		out[k] = v.Value()
	}
	out["$entity"] = ins.entId
	return out
}
func (ins EntityInstance) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, obj)
}
func (ins EntityInstance) Get(key string) (symbols.Object, error) {
	methods := map[string]symbols.Object{
		"update": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{ins.ParentType},
			Returns:   ins,
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				if ins.ParentType.eventLog == nil {
					return nil, fmt.Errorf("db connection not initialized")
				}
				ent := args[0].(*EntityType)
				state := stateEvent{
					EntityID:  ins.entId,
					Timestamp: time.Now(),
					Effect:    effectTypeUpdate,
					State:     ent.Value(),
				}
				if _, err := ins.ParentType.eventLog.InsertOne(context.TODO(), state); err != nil {
					return nil, err
				}
				ins.fields = ent.Fields
				return ins, nil
			},
		}),
		"delete": symbols.NewFunction(symbols.FunctionOptions{
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				if ins.ParentType.eventLog == nil {
					return nil, fmt.Errorf("db connection not initialized")
				}
				state := stateEvent{
					EntityID:  ins.entId,
					Timestamp: time.Now(),
					Effect:    effectTypeDelete,
					State:     nil,
				}
				if _, err := ins.ParentType.eventLog.InsertOne(context.TODO(), state); err != nil {
					return nil, err
				}
				return nil, nil
			},
		}),
	}
	if fn := methods[key]; fn != nil {
		return fn, nil
	}
	return ins.fields[key], nil
}

type effectType string

const (
	effectTypeCreate effectType = "CREATE"
	effectTypeUpdate effectType = "UPDATE"
	effectTypeDelete effectType = "DELETE"
)

func constructEntityFromInterface(ent Entity, state interface{}) (*EntityType, error) {
	bytes, err := bson.MarshalExtJSON(state, false, true)
	if err != nil {
		return nil, err
	}
	generic, err := symbols.FromBytes(bytes)
	if err != nil {
		return nil, err
	}
	obj, err := symbols.Construct(ent, generic)
	if err != nil {
		return nil, err
	}
	return obj.(*EntityType), nil
}

// Represents the internal type of the event change in the database
type stateEvent struct {
	NodeID    string      `bson:"node_id,omitempty"`
	EventID   string      `bson:"event_id,omitempty"`
	EntityID  string      `bson:"entity_id,omitempty"`
	Timestamp time.Time   `bson:"timestamp,omitempty"`
	Effect    effectType  `bson:"effect,omitempty"`
	State     interface{} `bson:"state,omitempty"`
}

func (ev stateEvent) EntityInstance(ent Entity) (EntityInstance, error) {
	entType, err := constructEntityFromInterface(ent, ev.State)
	if err != nil {
		return EntityInstance{}, err
	}
	return EntityInstance{
		ParentType: ent,
		entId:      ev.EntityID,
		fields:     entType.Fields,
	}, nil
}

// Represents the type exposed by a change stream event in mongo
type updateEvent struct {
	FullDocument stateEvent `bson:"fullDocument"`
}

// Represents the internal type of the event state in the databsae
type entityState struct {
	ObjectID  *primitive.ObjectID `bson:"_id,omitempty"`
	CreatedAt time.Time           `bson:"created_at,omitempty"`
	UpdatedAt time.Time           `bson:"updated_at,omitempty"`
	EntityID  string              `bson:"entity_id,omitempty"`
	State     interface{}         `bson:"state,omitempty"`
}

func (st entityState) EntityInstance(ent Entity) (EntityInstance, error) {
	entType, err := constructEntityFromInterface(ent, st.State)
	if err != nil {
		return EntityInstance{}, err
	}
	return EntityInstance{
		ParentType: ent,
		entId:      st.EntityID,
		fields:     entType.Fields,
	}, nil
}
