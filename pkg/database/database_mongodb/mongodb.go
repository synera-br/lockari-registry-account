package database_mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBConfig struct {
	Kind    string         `json:"kind" yaml:"kind"`
	MongoDB MongoDBConnect `json:"config" yaml:"config"`
}

func (m *MongoDBConfig) Validate() error {
	if m == nil {
		return fmt.Errorf("config is nil")
	}

	if err := m.MongoDB.Validate(); err != nil {
		return fmt.Errorf("error to validate config %v", err)
	}
	return nil
}

type MongoDBConnect struct {
	Host     string `json:"host" yaml:"host"`
	User     string `json:"user" yaml:"user"`
	Database string `json:"database" yaml:"database"`
	Password string `json:"password" yaml:"password"`
	Port     string `json:"port" yaml:"port"`
}

func (m *MongoDBConnect) GetURI() string {
	if m == nil {
		panic("MongoDBConnect is nil")
	}
	return fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority&appName=Cluster0", m.User, m.Password, m.Host)
}

func (m *MongoDBConnect) Validate() error {

	if m == nil {
		return fmt.Errorf("config is nil")
	}

	if m.Host == "" {
		return fmt.Errorf("host is required")
	}

	if m.User == "" {
		return fmt.Errorf("user is required")
	}

	if m.Password == "" {
		return fmt.Errorf("password is required")
	}

	if m.Database == "" {
		return fmt.Errorf("database is required")
	}

	if m.Port == "" {
		return fmt.Errorf("port is required")
	}

	return nil
}

// MongoDB implementation
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
	uri      string
	dbName   string
}

func NewMongoDB(uri, dbName string) (*MongoDB, error) {

	if uri == "" {
		return nil, fmt.Errorf("uri is required")
	}

	if dbName == "" {
		return nil, fmt.Errorf("dbName is required")
	}

	clientOptions := options.Client().ApplyURI(uri).SetMaxPoolSize(100).SetMinPoolSize(10)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, err
	}

	return &MongoDB{
		uri:      uri,
		dbName:   dbName,
		client:   client,
		database: client.Database(dbName),
	}, nil
}

func (m *MongoDB) Connect(ctx context.Context) error {

	if m == nil {
		return fmt.Errorf("MongoDB instance is nil")
	}

	clientOptions := options.Client().ApplyURI(m.uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if client == nil {
		return fmt.Errorf("error to connect at database")
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.database = client.Database(m.dbName)
	return nil
}

func (m *MongoDB) Close(ctx context.Context) error {
	if m == nil {
		return fmt.Errorf("MongoDB instance is nil")
	}

	if m.client != nil {
		return m.client.Disconnect(ctx)
	}
	return nil
}

func (m *MongoDB) GetType() string {
	if m == nil {
		panic("MongoDB instance is nil")
	}
	return "mongodb"
}

func (m *MongoDB) GetConnection() (*mongo.Database, error) {
	if m == nil {
		return nil, fmt.Errorf("MongoDB instance is nil")
	}
	if m.database == nil {
		return nil, fmt.Errorf("MongoDB database connection is nil - connection not established")
	}
	return m.database, nil
}
