package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type FirebaseDBInterface interface {
	Get(ctx context.Context, collection string) ([]byte, error)
	Create(ctx context.Context, data interface{}, collection string) ([]byte, error)
	Update(ctx context.Context, id string, data interface{}, collection string) error
	Delete(ctx context.Context, id, collection string) error
	GetByQuery(ctx context.Context, collection string) firestore.Query
	GetByConditional(ctx context.Context, conditional []Conditional, collection string) ([]byte, error)
	GetByFilter(ctx context.Context, filters map[string]interface{}, collection string) ([]byte, error)
	StructToData(data interface{}) (map[string]interface{}, error)
	IsConnected() bool
}

type Filter string

const (
	FilterEquals        Filter = "=="
	FilterNotEquals     Filter = "!="
	FilterGreaterThan   Filter = ">"
	FilterLessThan      Filter = "<"
	FilterArrayContains Filter = "array-contains"
)

type Conditional struct {
	Field  string
	Value  interface{}
	Filter Filter
}

type Conditionals []Conditional

func (c Conditional) Validate() error {
	if c.Field == "" {
		return errors.New("field is required")
	}
	if c.Value == nil {
		return errors.New("value is required")
	}
	if c.Filter == "" {
		return errors.New("filter is required")
	}
	if c.Filter != FilterEquals && c.Filter != FilterNotEquals &&
		c.Filter != FilterGreaterThan && c.Filter != FilterLessThan &&
		c.Filter != FilterArrayContains {
		return fmt.Errorf("invalid filter operator: %s", c.Filter)
	}
	return nil
}

// FirebaseDB implements the DatabaseService interface for Firebase Firestore.
type FirebaseDB struct {
	client *firestore.Client
}

// InitializeFirebaseDB creates and initializes a new FirebaseDB instance.
// It doesn't connect immediately; connection happens in the Connect method.
func InitializeFirebaseDB(config FirebaseConfig) (FirebaseDBInterface, error) {

	fdb := &FirebaseDB{}

	err := fdb.connect(config)
	if err != nil {
		return nil, err
	}

	return fdb, nil
}

// Connect establishes a connection to Firebase Firestore.
// config is expected to be a FirebaseConfig struct.
func (db *FirebaseDB) connect(cfg FirebaseConfig) error {

	if cfg.ProjectID == "" {
		return errors.New(errorProjectIDRequired)
	}

	if db.client != nil {
		return errors.New(errorClientNotInitialized)
	}

	var opt option.ClientOption
	if cfg.ServiceAccountKeyPath != "" {
		opt = option.WithCredentialsFile(cfg.ServiceAccountKeyPath)
	}

	databaseName := "default"
	if cfg.DatabaseURL != "" {
		databaseName = cfg.DatabaseURL
	}

	// Firestore doesn't have a concept of a "database name" like traditional RDBMS.
	// Connections are made to the project ID, and then you interact with collections and documents within that project.
	client, err := firestore.NewClientWithDatabase(
		context.Background(),
		cfg.ProjectID,
		databaseName,
		opt,
	)
	if err != nil {
		return fmt.Errorf(errorNotConnected, err.Error())
	}

	db.client = client

	return nil
}

// Get retrieves a single document by its ID from a default collection.
// Note: Firestore is schemaless, but often a default collection is used.
// This implementation assumes a collection name might be needed or configured elsewhere.
// For this example, let's assume Get needs a collection name.
// This highlights a potential mismatch with the generic interface if not handled carefully.
// We might need to adjust the interface or how collection names are passed.
// For now, this is a placeholder.
func (db *FirebaseDB) Get(ctx context.Context, collection string) ([]byte, error) {
	// Placeholder: A real implementation would specify the collection.
	if db.client == nil {
		return nil, errors.New(errorClientNotInitialized)
	}

	if err := db.validateWithoutData(ctx, collection); err != nil {
		return nil, err
	}

	doc, err := db.client.Collection(collection).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var results []interface{}
	for _, d := range doc {
		data := d.Data()
		data["id"] = d.Ref.ID
		results = append(results, data)
	}

	b, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Create adds a new document to a default collection.
// Placeholder: Collection name needed.
func (db *FirebaseDB) Create(ctx context.Context, data interface{}, collection string) ([]byte, error) {

	if err := db.validateWithData(ctx, data, collection); err != nil {
		return nil, err
	}

	var docData map[string]interface{}
	var err error

	if m, ok := data.(map[string]interface{}); ok {
		docData = m
	} else {

		bytes, _ := json.Marshal(data)
		json.Unmarshal(bytes, &docData)
		if docData == nil {
			return nil, fmt.Errorf("invalid data type: cannot convert to map for processing")
		}
	}

	// Gerar ID ou usar o existente. Priorize IDs gerados externamente (UUIDv7).
	docID := ""
	if idVal, ok := docData["id"]; ok && idVal != nil {
		if idStr, isStr := idVal.(string); isStr && idStr != "" {
			docID = idStr
		}
	}

	if docID == "" {
		// Gerar um UUIDv7 se nenhum ID for fornecido no 'data'
		newUUID, uuidErr := uuid.NewV7()
		if uuidErr != nil {
			return nil, fmt.Errorf("falha ao gerar UUIDv7 para o documento: %w", uuidErr)
		}
		docID = newUUID.String()
		docData["id"] = docID // Atualiza o ID no map de dados
	}

	// Adicionar timestamps de criação e atualização.
	// firestore.ServerTimestamp é um placeholder que o Firestore preenche no servidor.
	docData["createdAt"] = firestore.ServerTimestamp
	docData["updatedAt"] = firestore.ServerTimestamp

	// 2. Criar o documento no Firestore
	colRef := db.client.Collection(collection)
	docRef := colRef.Doc(docID)

	_, err = docRef.Set(ctx, docData) // Sem SetOptions significa sobrescrever
	if err != nil {
		return nil, fmt.Errorf("falha ao criar documento no Firestore com ID %s: %w", docID, err)
	}

	fmt.Println("Documento criado com sucesso:", docID)
	fmt.Println("Dados do documento:", docData)
	b, err := json.Marshal(docData)
	if err != nil {
		return nil, fmt.Errorf("falha ao serializar dados para JSON: %w", err)
	}

	return b, nil
}

// Update modifies an existing document in a default collection.
// Placeholder: Collection name needed.
func (db *FirebaseDB) Update(ctx context.Context, id string, data interface{}, collection string) error {
	if db.client == nil {
		return fmt.Errorf(errorClientNotInitialized)
	}

	// Check if data is a map and if "updatedAt" is present.
	// If data is not a map or "updatedAt" is not present, add it.
	if mapData, ok := data.(map[string]interface{}); ok {
		if _, exists := mapData["updatedAt"]; !exists {
			mapData["updatedAt"] = firestore.ServerTimestamp
		}
	}

	if id == "" {
		return fmt.Errorf("id is empty")
	}

	if err := db.validateWithData(ctx, data, collection); err != nil {
		return err
	}

	_, err := db.client.Collection(collection).Doc(id).Set(ctx, data, firestore.MergeAll)
	if err != nil {
		return err
	}

	return nil
}

// Delete removes a document from a default collection by its ID.
// Placeholder: Collection name needed.
func (db *FirebaseDB) Delete(ctx context.Context, id, collection string) error {
	if db.client == nil {
		return errors.New(errorClientNotInitialized)
	}
	if id == "" {
		return fmt.Errorf(errorGenericError, "id is empty")
	}

	if err := db.validateWithoutData(ctx, collection); err != nil {
		return err
	}

	_, err := db.client.Collection(collection).Doc(id).Delete(ctx)

	return err
}

func (db *FirebaseDB) GetByQuery(ctx context.Context, collection string) firestore.Query {

	query := db.client.Collection(collection).Query

	return query

}

// GetByConditional retrieves multiple documents based on a set of conditionals from a default collection.
// Placeholder: Collection name needed.
// This function allows for more complex queries using conditionals.
func (db *FirebaseDB) GetByConditional(ctx context.Context, conditional []Conditional, collection string) ([]byte, error) {

	if err := db.validateWithData(ctx, conditional, collection); err != nil {
		return nil, err
	}

	if db.client == nil {
		return nil, errors.New(errorClientNotInitialized)
	}

	if len(conditional) == 0 {
		return nil, errors.New(errorConditionalRequired)
	}

	query := db.client.Collection(collection).Query
	for _, cond := range conditional {
		if cond.Field == "" {
			return nil, errors.New(errorConditionalFieldRequired)
		}
		if cond.Value == nil {
			return nil, errors.New(errorConditionalValueRequired)
		}
		if cond.Filter == "" {
			return nil, errors.New(errorConditionalFilterRequired)
		}
		if cond.Filter != FilterEquals && cond.Filter != FilterNotEquals &&
			cond.Filter != FilterGreaterThan && cond.Filter != FilterLessThan &&
			cond.Filter != FilterArrayContains {
			return nil, fmt.Errorf(errorGenericError, "invalid filter operator")
		}

		query = query.Where(cond.Field, string(cond.Filter), cond.Value)
	}

	iter := query.Documents(ctx)

	var results []interface{}
	defer iter.Stop()
	for {
		doc, err := iter.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		data := doc.Data()
		data["id"] = doc.Ref.ID // Importante para updates
		results = append(results, data)
	}

	b, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GetByFilter retrieves multiple documents based on a set of filters from a default collection.
// Placeholder: Collection name needed.
func (db *FirebaseDB) GetByFilter(ctx context.Context, filters map[string]interface{}, collection string) ([]byte, error) {

	if err := db.validateWithData(ctx, filters, collection); err != nil {
		return nil, err
	}

	query := db.client.Collection(collection).Query
	for key, value := range filters {
		query = query.Where(key, "==", value)
	}

	iter := query.Documents(ctx)
	var results []interface{}
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		data := doc.Data()
		data["id"] = doc.Ref.ID // Importante para updates
		results = append(results, data)
	}

	b, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Close terminates the Firebase connection.
func (db *FirebaseDB) Close() error {
	if db.client != nil {
		err := db.client.Close()
		if err != nil {
			return fmt.Errorf(errorToCloseConnection, err.Error())
		}
		return nil
	}
	return errors.New(errorNotInitialized)
}

func (db *FirebaseDB) IsConnected() bool {
	return db.client != nil
}

// validateWithData
func (db *FirebaseDB) validateWithData(ctx context.Context, data interface{}, collection string) error {
	if db.client == nil {
		return errors.New(errorNotInitialized)
	}

	if data == nil {
		return errors.New("data is nil")
	}

	if collection == "" {
		return errors.New(errorCollectionRequired)
	}

	if ctx.Value("Authorization") == "" {
		return errors.New(errorAuthorizationTokenRequired)
	}
	return nil
}

// validateWithoutData
func (db *FirebaseDB) validateWithoutData(ctx context.Context, collection string) error {
	if db.client == nil {
		return errors.New(errorNotInitialized)
	}

	if collection == "" {
		return errors.New(errorCollectionRequired)
	}

	if ctx.Value("Authorization") == "" {
		return errors.New("authorization token is nil")
	}
	return nil
}

func (db *FirebaseDB) StructToData(data interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}

	if data == nil {
		return nil, fmt.Errorf(errorGenericError, "data is nil")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, err
	}

	if result == nil {
		result = make(map[string]interface{})
	}
	if len(result) == 0 {
		return nil, fmt.Errorf(errorGenericError, "resulting map is empty after unmarshalling")
	}

	return result, nil
}
