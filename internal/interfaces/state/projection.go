package state

import (
	ctx "context"
	"fmt"
	"strings"

	"github.com/hntrl/hyper/internal/ast"
	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/interfaces/stream"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/hntrl/hyper/internal/runtime/log"
	"github.com/hntrl/hyper/internal/runtime/resource"
	"github.com/hntrl/hyper/internal/symbols"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Projection struct {
	ParentType symbols.Type
	events     map[*stream.Event]*symbols.Function `hash:"ignore"`
	collection *mongo.Collection                   `hash:"ignore"`
}

var ProjectionSignal = log.Signal("PROJECTION")
var ProjectionEventSignal = log.Signal("PROJECTION_EVENT")

func (proj Projection) ClassName() string {
	return "Projection"
}
func (proj Projection) Fields() map[string]symbols.Class {
	return proj.ParentType.Fields()
}
func (proj Projection) Constructors() symbols.ConstructorMap {
	csMap := symbols.NewConstructorMap()
	csMap.AddConstructor(ProjectionRecord{parentType: proj}, func(obj symbols.ValueObject) (symbols.ValueObject, error) {
		rec := obj.(ProjectionRecord)
		return &ProjectionType{proj, rec.fields}, nil
	})
	csMap.AddGenericConstructor(proj, func(fields map[string]symbols.ValueObject) (symbols.ValueObject, error) {
		return &ProjectionType{proj, fields}, nil
	})
	return csMap
}

func (proj Projection) Class() symbols.Class {
	return proj
}
func (proj Projection) Value() interface{} {
	return nil
}
func (proj Projection) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, proj)
}
func (proj Projection) Get(key string) (symbols.Object, error) {
	methods := map[string]symbols.Object{
		"find": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.NewPartialObject(proj),
				QueryOpts{},
			},
			Returns: symbols.Iterable{
				ParentType: ProjectionRecord{parentType: proj},
				Items:      []symbols.ValueObject{},
			},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				if proj.collection == nil {
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
				err := flattenObject(filterObject, filter, "")
				if err != nil {
					return nil, err
				}

				cursor, err := proj.collection.Find(ctx.TODO(), filter, fo)
				if err != nil {
					return nil, err
				}

				itemIdx := 0
				iter := symbols.NewIterable(proj.ParentType, cursor.RemainingBatchLength())

				for cursor.Next(ctx.TODO()) {
					var record bson.M
					err := cursor.Decode(&record)
					if err != nil {
						return nil, err
					}

					bytes, err := bson.MarshalExtJSON(cursor.Current, false, true)
					if err != nil {
						return nil, err
					}
					genericObject, err := symbols.FromBytes(bytes)
					if err != nil {
						return nil, err
					}
					obj, err := symbols.Construct(proj, genericObject)
					if err != nil {
						return nil, err
					}
					if typeObject, ok := obj.(*ProjectionType); ok {
						if recordId, ok := record["_id"].(primitive.ObjectID); ok {
							iter.Items[itemIdx] = ProjectionRecord{
								parentType: proj,
								recordId:   recordId,
								fields:     typeObject.fields,
							}
							itemIdx++
						} else {
							return nil, fmt.Errorf("projection record has no identifier")
						}
					} else {
						return nil, fmt.Errorf("expected projection type")
					}
				}
				return iter, nil
			},
		}),
		"findOne": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				symbols.NewPartialObject(proj),
				QueryOpts{},
			},
			Returns: ProjectionRecord{parentType: proj},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				if proj.collection == nil {
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
				err := flattenObject(filterObject, filter, "")
				if err != nil {
					return nil, err
				}

				var record bson.M
				res := proj.collection.FindOne(ctx.TODO(), filter, fo)
				err = res.Decode(&record)
				if err != nil {
					if err == mongo.ErrNoDocuments {
						return nil, fmt.Errorf("no matching documents")
					}
					return nil, err
				}

				bytes, err := bson.MarshalExtJSON(record, false, true)
				if err != nil {
					return nil, err
				}
				genericObject, err := symbols.FromBytes(bytes)
				if err != nil {
					return nil, err
				}
				obj, err := symbols.Construct(proj, genericObject)
				if err != nil {
					return nil, err
				}
				if typeObject, ok := obj.(*ProjectionType); ok {
					if recordId, ok := record["_id"].(primitive.ObjectID); ok {
						return ProjectionRecord{
							parentType: proj,
							recordId:   recordId,
							fields:     typeObject.fields,
						}, nil
					} else {
						return nil, fmt.Errorf("projection record has no identifier")
					}
				} else {
					return nil, fmt.Errorf("expected projection type")
				}
			},
		}),
		"insert": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{
				proj,
			},
			Returns: ProjectionRecord{parentType: proj},
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				if proj.collection == nil {
					return nil, fmt.Errorf("db connection not initialized")
				}
				passedType, ok := args[0].(*ProjectionType)
				if !ok {
					return nil, fmt.Errorf("expected entity type")
				}
				res, err := proj.collection.InsertOne(ctx.TODO(), passedType.Value())
				if err != nil {
					return nil, err
				}
				if id, ok := res.InsertedID.(primitive.ObjectID); ok {
					return ProjectionRecord{parentType: proj, recordId: id, fields: passedType.fields}, nil
				} else {
					return nil, fmt.Errorf("returned record did not have identifier")
				}
			},
		}),
	}
	return methods[key], nil
}

func (proj *Projection) Attach(process *runtime.Process) error {
	return nil
}
func (proj *Projection) AttachResource(process *runtime.Process) error {
	var natsConn resource.NatsConnection
	err := process.Resource("nats", &natsConn)
	if err != nil {
		return err
	}
	var dbConn resource.MongoConnection
	err = process.Resource("mdb", &dbConn)
	if err != nil {
		return err
	}
	for evPtr, fnPtr := range proj.events {
		ev := *evPtr
		fn := *fnPtr
		topic := fmt.Sprintf("%s.%s", process.Context.Name, ev.ParentType.Name)
		natsConn.Client.QueueSubscribe(topic, "projection_group", func(m *nats.Msg) {
			obj, err := symbols.FromBytes(m.Data)
			if err != nil {
				log.Output(log.LoggerMessage{
					LogLevel: log.LevelERROR,
					Signal:   ProjectionEventSignal,
					Message:  "failed to construct object from event",
					Data: log.LogData{
						"err":   err.Error(),
						"event": string(m.Data),
					},
				})
				return
			}
			_, err = fn.Call([]symbols.ValueObject{obj}, stream.Message{})
			if err != nil {
				log.Output(log.LoggerMessage{
					LogLevel: log.LevelERROR,
					Signal:   ProjectionEventSignal,
					Message:  "projection event handler failed invocation",
					Data: log.LogData{
						"err":   err.Error(),
						"event": obj,
					},
				})
				return
			}
			log.Printf(log.LevelINFO, ProjectionEventSignal, "projection \"%s\" received \"%s\"", proj.ParentType.Name, m.Subject)
		})
		log.Printf(log.LevelDEBUG, ProjectionEventSignal, "projection \"%s\" listening for \"%s\"", proj.ParentType.Name, evPtr.Topic)
	}
	dbName := strings.Replace(process.Context.Name, ".", "_", -1)
	collection, err := dbConn.EnsureCollection(dbName, proj.ParentType.Name)
	if err != nil {
		return err
	}
	proj.collection = collection
	log.Printf(log.LevelDEBUG, ProjectionSignal, "projection \"%s:%s\" collection created", dbName, proj.ParentType.Name)
	return nil
}
func (proj *Projection) Detach() error {
	return nil
}

func (proj Projection) ObjectClassFromNode(ctx *context.Context, node ast.ContextObject) (symbols.Class, error) {
	assumedType, err := (context.TypeInterface{}).ObjectClassFromNode(ctx, node)
	if err != nil {
		return nil, err
	}
	if typeClass, ok := assumedType.(symbols.Type); ok {
		return &Projection{typeClass, make(map[*stream.Event]*symbols.Function), nil}, nil
	} else {
		panic("expected type class")
	}
}

func (proj *Projection) AddMethod(ctx *context.Context, node ast.ContextObjectMethod) error {
	table := ctx.Symbols()
	if node.Name != "onEvent" {
		return symbols.NodeError(node, "unrecognized method %s on projection", node.Name)
	}
	if len(node.Block.Parameters.Arguments.Items) != 1 {
		return symbols.NodeError(node, "onEvent method must have one argument")
	}
	if node.Block.Parameters.ReturnType != nil {
		return symbols.NodeError(node, "onEvent method cannot have a return type")
	}
	arg := node.Block.Parameters.Arguments.Items[0]
	if argExpr, ok := arg.(ast.ArgumentItem); ok {
		class, err := table.ResolveTypeExpression(argExpr.Init)
		if err != nil {
			return err
		}
		if event, ok := class.(*stream.Event); ok {
			fn, err := table.ResolveFunctionBlock(node.Block, &symbols.MapObject{})
			if err != nil {
				return err
			}
			proj.events[event] = fn
			return nil
		} else {
			return symbols.NodeError(arg, fmt.Sprintf("argument to onEvent must be an event, got %T", class))
		}
	} else {
		return symbols.NodeError(arg, "argument to onEvent must be an event, got type")
	}
}

func (proj Projection) Export() (symbols.Object, error) {
	return proj, nil
}

type ProjectionType struct {
	parentType Projection
	fields     map[string]symbols.ValueObject `hash:"ignore"`
}

func (tp ProjectionType) Class() symbols.Class {
	return tp.parentType
}
func (tp ProjectionType) Value() interface{} {
	out := make(map[string]interface{})
	for k, v := range tp.fields {
		out[k] = v.Value()
	}
	return out
}
func (tp *ProjectionType) Set(key string, obj symbols.ValueObject) error {
	tp.fields[key] = obj
	return nil
}
func (tp ProjectionType) Get(key string) (symbols.Object, error) {
	return tp.fields[key], nil
}

type ProjectionRecord struct {
	parentType Projection
	recordId   primitive.ObjectID             `bson:"_id,omitempty" hash:"ignore"`
	fields     map[string]symbols.ValueObject `hash:"ignore"`
}

func (rec ProjectionRecord) ClassName() string {
	return fmt.Sprintf("[%s]", rec.parentType.ClassName())
}
func (rec ProjectionRecord) Fields() map[string]symbols.Class {
	return rec.parentType.Fields()
}
func (rec ProjectionRecord) Constructors() symbols.ConstructorMap {
	return symbols.NewConstructorMap()
}

func (rec ProjectionRecord) Class() symbols.Class {
	return rec
}
func (rec ProjectionRecord) Value() interface{} {
	out := make(map[string]interface{})
	for k, v := range rec.fields {
		out[k] = v.Value()
	}
	return out
}
func (rec ProjectionRecord) Set(key string, obj symbols.ValueObject) error {
	return symbols.CannotSetPropertyError(key, obj)
}
func (rec ProjectionRecord) Get(key string) (symbols.Object, error) {
	methods := map[string]symbols.Object{
		"update": symbols.NewFunction(symbols.FunctionOptions{
			Arguments: []symbols.Class{rec.parentType},
			Returns:   rec,
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				if rec.parentType.collection == nil {
					return nil, fmt.Errorf("db connection not initialized")
				}
				val := args[0].(*ProjectionType)
				res := rec.parentType.collection.FindOneAndUpdate(ctx.TODO(), bson.M{"$record": rec.recordId}, val.Value())
				if err := res.Err(); err != nil {
					return nil, err
				}
				rec.fields = val.fields
				return rec, nil
			},
		}),
		"delete": symbols.NewFunction(symbols.FunctionOptions{
			Handler: func(args []symbols.ValueObject, proto symbols.ValueObject) (symbols.ValueObject, error) {
				if rec.parentType.collection == nil {
					return nil, fmt.Errorf("db connection not initialized")
				}
				res := rec.parentType.collection.FindOneAndDelete(ctx.TODO(), bson.M{"_id": rec.recordId})
				if err := res.Err(); err != nil {
					return nil, err
				}
				return nil, nil
			},
		}),
	}
	return methods[key], nil
}
