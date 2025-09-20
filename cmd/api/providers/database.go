package providers

import (
	"encoding/json"
	"fmt"
	"registry-account/cmd/api/types"
	"registry-account/pkg/database/database_mongodb"

	"go.mongodb.org/mongo-driver/mongo"
)

func ProvideDatabase(params types.DatabaseParams) (*database_mongodb.MongoDB, *mongo.Database, error) {
	fields := params.Config.Settings["database"]
	if fields == nil {
		return nil, nil, fmt.Errorf("database config is nil")
	}

	b, err := json.Marshal(fields)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal database config: %w", err)
	}

	var fdata database_mongodb.MongoDBConfig
	err = json.Unmarshal(b, &fdata)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal database config: %w", err)
	}

	if fdata.MongoDB.Host == "" || fdata.MongoDB.Database == "" {
		return nil, nil, fmt.Errorf("incomplete database configuration")
	}

	dbResult, err := database_mongodb.NewMongoDB(fdata.MongoDB.GetURI(), fdata.MongoDB.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	conn, err := dbResult.GetConnection()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	return dbResult, conn, nil
}
