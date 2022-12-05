package resource

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConnection struct {
	Client *mongo.Client
}

func (conn MongoConnection) Attach() (Resource, error) {
	url := fmt.Sprintf("mongodb://%s:%d/?directConnection=true", os.Getenv("MONGO_HOST"), 27017)
	opts := options.Client().ApplyURI(url)
	if os.Getenv("MONGO_PASSWORD") != "" {
		// creds enabled
		cred := options.Credential{
			Username: os.Getenv("MONGO_USER"),
			Password: os.Getenv("MONGO_PASSWORD"),
		}
		opts = opts.SetAuth(cred)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}
	conn.Client = client
	return conn, nil
}
func (conn MongoConnection) Detach() error {
	if conn.Client == nil {
		return nil
	}
	err := conn.Client.Disconnect(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

func (conn MongoConnection) EnsureCollection(dbName string, coll string) (*mongo.Collection, error) {
	db := conn.Client.Database(dbName)
	if err := db.CreateCollection(context.TODO(), coll); err != nil {
		if v, ok := err.(mongo.CommandError); ok && v.Name == "NamespaceExists" {
			// do nothing, maybe report incompatability?
		} else {
			return nil, err
		}
	}
	return db.Collection(coll), nil
}
