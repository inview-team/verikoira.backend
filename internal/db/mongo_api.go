package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type Storage struct {
	client *mongo.Client
}

const (
	addr     = "mongodb://mongodb:27017"
	database = "storage"
)

func connect() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(addr)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Connect(context.TODO())
	if err != nil {
		return nil, err
	}
	return client, client.Ping(context.TODO(), nil)
}

func New() (*Storage, error) {
	be := backoff.NewExponentialBackOff()
	be.MaxElapsedTime = time.Minute
	be.InitialInterval = 1 * time.Second
	be.Multiplier = 2
	be.MaxInterval = 15 * time.Second

	b := backoff.WithContext(be, context.Background())
	for {
		d := b.NextBackOff()
		if d == backoff.Stop {
			return nil, fmt.Errorf("stop reconnecting")
		}
		<-time.After(d)
		cli, err := connect()
		if err != nil {
			zap.L().Debug("could not reconnect", zap.Error(err))
			continue
		}
		mongoSt := &Storage{
			client: cli,
		}
		return mongoSt, nil
	}
}

func (s *Storage) GetResult(id string) ([]byte, error) {
	var data []struct {
	}

	coll := s.client.Database(database).Collection(id)
	cur, err := coll.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}

	err = cur.All(context.Background(), &data)
	if err != nil {
		return nil, err
	}

	return json.Marshal(data)
}
