package revmongo

import (
	"context"
	"errors"
	"time"

	opts "github.com/qiniu/qmgo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateIndexSpecPartial set partiaFilter by user
func CreateIndexSpecPartial(model interface{}, keys []string, unique bool, partialFilter bson.M) (string, error) {
	colName := getModelName(model)
	coll := DB.Collection(colName)
	err := coll.CreateOneIndex(context.Background(), opts.IndexModel{
		Key: keys,
		IndexOptions: &options.IndexOptions{
			PartialFilterExpression: partialFilter,
			Unique:                  &unique,
		},
	})
	if err != nil {
		return "", err
	}
	return "", nil
}

// All index with PartialFilterExperession IsRemoved = false
// CreateIndex with keys and unique (true or false)
func CreateIndex(model interface{}, keys []string, unique bool) (string, error) {
	colName := getModelName(model)
	coll := DB.Collection(colName)
	err := coll.CreateOneIndex(context.Background(), opts.IndexModel{
		Key: keys,
		IndexOptions: &options.IndexOptions{
			PartialFilterExpression: bson.M{"IsRemoved": false},
			Unique:                  &unique,
		},
	})
	if err != nil {
		return "", err
	}
	return "", nil
}

// All index without PartialFilterExperession IsRemoved = false
// CreateIndex with keys and unique (true or false)
func CreateIndexNoPartial(model interface{}, keys []string, unique bool) (string, error) {
	colName := getModelName(model)
	coll := DB.Collection(colName)
	err := coll.CreateOneIndex(context.Background(), opts.IndexModel{
		Key: keys,
		IndexOptions: &options.IndexOptions{
			Unique: &unique,
		},
	})
	if err != nil {
		return "", err
	}
	return "", nil
}

func CreateTextIndex(model interface{}, keys []string) (string, error) {
	colName := getModelName(model)
	coll := DB.Collection(colName)
	err := coll.CreateOneIndex(context.Background(), opts.IndexModel{
		Key: keys,
	})
	if err != nil {
		return "", err
	}
	return "", nil
}

// CreateExpireIndex create expire index for single field
func CreateExpireIndex(model interface{}, fieldName string, expireAfter int32) (string, error) {
	colName := getModelName(model)
	coll := DB.Collection(colName)
	err := coll.CreateOneIndex(context.Background(), opts.IndexModel{
		Key:          []string{fieldName},
		IndexOptions: options.Index().SetExpireAfterSeconds(expireAfter),
	})
	return "", err
}

// IsDup check whether the err is Duplicate error
func IsDup(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}

// CollectionIndexes listout collection indexes
func CollectionIndexes(model interface{}) ([]string, error) {
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(Dial))
	if err != nil {
		return nil, err
	}

	mongoDB := mongoClient.Database(DBName)
	colName := getModelName(model)
	collection := mongoDB.Collection(colName)
	indexView := collection.Indexes()
	opts := options.ListIndexes().SetMaxTime(2 * time.Second)

	var results []bson.M

	cursor, err := indexView.List(context.TODO(), opts)
	if err != nil {
		return nil, err
	}

	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	var output []string
	for _, item := range results {
		output = append(output, item["name"].(string))
	}
	return output, nil
}

// Drop One Index from collection with indexName
func DropOneIndex(model interface{}, name string) error {
	colName := getModelName(model)
	return DB.Collection(colName).DropIndex(context.Background(), []string{name})
}
