package canned

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestInsertDocument(t *testing.T) {
	ctx := context.Background()

	c, err := MongoDbContainer(ctx, MongoDbContainerRequest{})
	if err != nil {
		t.Fatal(err.Error())
	}

	defer c.Container.Terminate(ctx)

	mongoClient, err := c.GetDriver()
	if err != nil {
		t.Fatal(err.Error())
	}

	collection := mongoClient.Database("testdatabase").Collection("persons")

	result, err := collection.InsertOne(context.TODO(), bson.D{primitive.E{Key: "name", Value: "John Doe"}})
	if err != nil {
		t.Fatal(err.Error())
	}

	if result.InsertedID == nil {
		t.Fatal("Insert failed")
	}

	mongoClient.Disconnect(context.TODO())

	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestInsertDocumentWithMongoDbContainerRequestParameters(t *testing.T) {
	ctx := context.Background()

	testDbName := "testdb"

	c, err := MongoDbContainer(ctx, MongoDbContainerRequest{
		Database: testDbName,
		User:     "top",
		Password: "secret",
	})

	defer c.Container.Terminate(ctx)

	mongoClient, err := c.GetDriver()
	if err != nil {
		t.Fatal(err.Error())
	}

	collection := mongoClient.Database(testDbName).Collection("persons")

	result, err := collection.InsertOne(context.TODO(), bson.D{primitive.E{Key: "name", Value: "John Doe"}})
	if err != nil {
		t.Fatal(err.Error())
	}

	if result.InsertedID == nil {
		t.Fatal("Insert failed")
	}

	mongoClient.Disconnect(context.TODO())

	if err != nil {
		t.Fatal(err.Error())
	}
}
