package db

import (
	"context"

	"github.com/inview-team/verikoira.backend/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type Company struct {
	Name     string `json:"name"`
	Contacts string `json:"contacts"`
	Info     string `json:"info"`
	Keywords string `json:"keywords"`
}

func New(conf *config.DatabaseConfig, ctx context.Context) (*Storage, error) {
	clientOptions := options.Client().ApplyURI(conf.Address)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	mongoSt := &Storage{
		client:     client,
		collection: client.Database(conf.Database).Collection(conf.Collection),
	}

	return mongoSt, nil
}

func (s *Storage) SaveData(data Company) error {
	_, err := s.collection.InsertOne(context.TODO(), data)
	return err
}
