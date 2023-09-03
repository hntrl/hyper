package mongodb

import (
	"context"
	"fmt"

	"github.com/hntrl/hyper/src/hyper/symbols"
	"github.com/hntrl/hyper/src/resources/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoConnection struct {
	DatabaseName string
	client       *mongo.Client
	collections  []*mongo.Collection
}

func buildFilter(filter database.Filter) bson.M {
	switch filter := filter.(type) {
	case database.ScalarFilter:
		var value bson.M
		switch filter.Condition {
		case database.ScalarConditionEquals:
			value = bson.M{"$eq": filter.Value}
		case database.ScalarConditionContains:
			value = bson.M{"$regex": filter.Value}
		case database.ScalarConditionStartsWith:
			value = bson.M{"$regex": fmt.Sprintf("^%s", filter.Value)}
		case database.ScalarConditionEndsWith:
			value = bson.M{"$regex": fmt.Sprintf("%s$", filter.Value)}
		case database.ScalarConditionLessThan:
			value = bson.M{"$lt": filter.Value}
		case database.ScalarConditionLessThanEquals:
			value = bson.M{"$lte": filter.Value}
		case database.ScalarConditionGreaterThan:
			value = bson.M{"$gt": filter.Value}
		case database.ScalarConditionGreaterThanEquals:
			value = bson.M{"$gte": filter.Value}
		case database.ScalarConditionIn:
			value = bson.M{"$in": filter.Value}
		case database.ScalarConditionIsSet:
			value = bson.M{"$or": bson.M{"$not": nil, "$exists": true}}
		}
		return bson.M{string(filter.Key): value}
	case database.ListFilter:
		var value bson.M
		switch filter.Condition {
		case database.ListConditionEquals:
			value = bson.M{"$eq": filter.Value}
		case database.ListConditionHas:
			value = bson.M{"$all": bson.A{filter.Value}}
		case database.ListConditionHasEvery:
			value = bson.M{"$all": filter.Value}
		case database.ListConditionHasSome:
			value = bson.M{"$in": filter.Value}
		case database.ListConditionIsEmpty:
			value = bson.M{"$size": 0}
		case database.ListConditionIsSet:
			value = bson.M{"$or": bson.M{"$not": nil, "$exists": true}}
		}
		return bson.M{string(filter.Key): value}
	case database.CompositeFilter:
		subFilters := make([]bson.M, len(filter.Filters))
		for i, subFilter := range filter.Filters {
			subFilters[i] = buildFilter(subFilter)
		}
		switch filter.Condition {
		case database.CompositeConditionAND:
			return bson.M{"$and": subFilters}
		case database.CompositeConditionOR:
			return bson.M{"$or": subFilters}
		case database.CompositeConditionNOT:
			return bson.M{"$not": subFilters}
		}
	}
	return nil
}

func (mc *MongoConnection) getCollection(m database.Model) (*mongo.Collection, error) {
	for _, collection := range mc.collections {
		if collection.Name() == m.TableName() {
			return collection, nil
		}
	}
	db := mc.client.Database(mc.DatabaseName)
	collectionName := m.TableName()
	if err := db.CreateCollection(context.Background(), collectionName); err != nil {
		if mongoError, ok := err.(mongo.CommandError); ok && mongoError.Name != "NamespaceExists" {
			return nil, err
		}
	}
	mc.collections = append(mc.collections, db.Collection(collectionName))
	return db.Collection(collectionName), nil
}

func (mc MongoConnection) recordFromInterface(m database.Model, val bson.M) database.Record {
	id, ok := val["_id"].(primitive.ObjectID)
	if !ok {
		// FIXME: gracefully handle this error (it shouldn't happen, but hey here we are)
		panic("expected _id to be primitive.ObjectID")
	}
	return database.NewRecordValue(m, id.String(), val)
}

func (mc MongoConnection) GetRecord(m database.Model, filter database.Filter) (*database.Record, error) {
	if err := m.ValidateFilter(filter); err != nil {
		return nil, err
	}
	coll, err := mc.getCollection(m)
	if err != nil {
		return nil, err
	}
	var result bson.M
	if err := coll.FindOne(context.Background(), buildFilter(filter)).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, symbols.ErrorValue{
				Name:    "NotFound",
				Message: "no matching documents",
			}
		}
		return nil, err
	}
	return nil, nil
}
func (mc MongoConnection) GetManyRecords(m database.Model, queryArguments database.QueryArguments) (database.Cursor, error) {
	if err := m.ValidateFilter(queryArguments.Filter); err != nil {
		return nil, err
	}
	coll, err := mc.getCollection(m)
	if err != nil {
		return nil, err
	}
	cursor, err := coll.Find(context.Background(), buildFilter(queryArguments.Filter))
	if err != nil {
		return nil, err
	}
	return &MongoCursor{
		model:       m,
		conn:        mc,
		mongoCursor: cursor,
	}, nil
}

func (mc MongoConnection) CreateRecord(m database.Model, writeArguments database.WriteArguments) error {
	if err := m.ValidateWriteArguments(writeArguments); err != nil {
		return err
	}
	coll, err := mc.getCollection(m)
	if err != nil {
		return err
	}
	if _, err := coll.InsertOne(context.Background(), writeArguments); err != nil {
		return err
	}
	return nil
}
func (mc MongoConnection) CreateManyRecords(m database.Model, writeArguments []database.WriteArguments) (uint64, error) {
	coll, err := mc.getCollection(m)
	if err != nil {
		return 0, err
	}
	records := make([]interface{}, len(writeArguments))
	for i, item := range writeArguments {
		if err := m.ValidateWriteArguments(item); err != nil {
			return 0, err
		}
		records[i] = writeArguments
	}
	if _, err := coll.InsertMany(context.Background(), records); err != nil {
		return 0, err
	}
	return uint64(len(records)), nil
}

func (mc MongoConnection) UpdateOne(m database.Model, filter database.Filter, writeArguments database.WriteArguments) error {
	if err := m.ValidateFilter(filter); err != nil {
		return err
	}
	if err := m.ValidateWriteArguments(writeArguments); err != nil {
		return err
	}
	coll, err := mc.getCollection(m)
	if err != nil {
		return err
	}
	if _, err := coll.UpdateOne(context.Background(), buildFilter(filter), writeArguments); err != nil {
		return err
	}
	return nil
}
func (mc MongoConnection) UpdateMany(m database.Model, filter database.Filter, writeArguments database.WriteArguments) (uint64, error) {
	if err := m.ValidateFilter(filter); err != nil {
		return 0, err
	}
	if err := m.ValidateWriteArguments(writeArguments); err != nil {
		return 0, err
	}
	coll, err := mc.getCollection(m)
	if err != nil {
		return 0, err
	}
	updateResult, err := coll.UpdateMany(context.Background(), buildFilter(filter), writeArguments)
	if err != nil {
		return 0, err
	}
	return uint64(updateResult.ModifiedCount), nil
}
func (mc MongoConnection) UpdateRecord(m database.Model, record database.Record, writeArguments database.WriteArguments) (*database.Record, error) {
	if err := m.ValidateWriteArguments(writeArguments); err != nil {
		return nil, err
	}
	coll, err := mc.getCollection(m)
	if err != nil {
		return nil, err
	}
	var result bson.M
	if err := coll.FindOneAndUpdate(context.Background(), bson.M{"_id": record.Identifier()}, writeArguments).Decode(&result); err != nil {
		return nil, err
	}
	newRecord := mc.recordFromInterface(m, result)
	return &newRecord, nil
}

func (mc MongoConnection) DeleteOne(m database.Model, filter database.Filter) error {
	if err := m.ValidateFilter(filter); err != nil {
		return err
	}
	coll, err := mc.getCollection(m)
	if err != nil {
		return err
	}
	_, err = coll.DeleteOne(context.Background(), buildFilter(filter))
	if err != nil {
		return err
	}
	return nil
}
func (mc MongoConnection) DeleteMany(m database.Model, filter database.Filter) (uint64, error) {
	if err := m.ValidateFilter(filter); err != nil {
		return 0, err
	}
	coll, err := mc.getCollection(m)
	if err != nil {
		return 0, err
	}
	deleteResult, err := coll.DeleteMany(context.Background(), buildFilter(filter))
	if err != nil {
		return 0, err
	}
	return uint64(deleteResult.DeletedCount), nil
}
func (mc MongoConnection) DeleteRecord(m database.Model, record database.Record) error {
	coll, err := mc.getCollection(m)
	if err != nil {
		return nil
	}
	_, err = coll.DeleteOne(context.Background(), bson.M{"_id": record.Identifier()})
	if err != nil {
		return err
	}
	return nil
}

type MongoCursor struct {
	model       database.Model
	conn        MongoConnection
	mongoCursor *mongo.Cursor
}

func (c *MongoCursor) All() ([]database.Record, error) {
	items := make([]bson.M, c.mongoCursor.RemainingBatchLength())
	err := c.mongoCursor.All(context.Background(), &items)
	if err != nil {
		return nil, err
	}
	records := make([]database.Record, len(items))
	for idx, item := range items {
		records[idx] = c.conn.recordFromInterface(c.model, item)
	}
	return records, nil
}
func (c *MongoCursor) Next() bool {
	return c.mongoCursor.Next(context.Background())
}
func (c *MongoCursor) Current() (*database.Record, error) {
	var current bson.M
	err := bson.Unmarshal(c.mongoCursor.Current, &current)
	if err != nil {
		return nil, err
	}
	record := c.conn.recordFromInterface(c.model, current)
	return &record, nil
}
func (c *MongoCursor) RemainingLength() uint64 {
	return uint64(c.mongoCursor.RemainingBatchLength())
}
