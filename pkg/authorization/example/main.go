package main

import (
	"context"
	"fmt"
	"log"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	"github.com/openfga/go-sdk/credentials"
	// 3. Criar service básico
)

type ConfigOpenFGA struct {
	APIURL               string
	StoreID              string
	AuthorizationModelID string
	APITokenIssuer       string
	APIAudience          string
	ClientID             string
	ClientSecret         string
	Scopes               string
}

var openFGAClient *client.OpenFgaClient

func NewConfigOpenFGA() *ConfigOpenFGA {
	return &ConfigOpenFGA{
		APIURL:               "https://api.us1.fga.dev",
		StoreID:              "01K0DH8V7Q7NDTJ73KTDPERC81", // Será detectado automaticamente se vazio
		AuthorizationModelID: "01K0ERJX9TJM30Q2V3AWM0A56N", // Será detectado automaticamente se vazio
		APITokenIssuer:       "auth.fga.dev",               // Em produção, deve ser o emissor do token JWT
		APIAudience:          "https://api.us1.fga.dev/",
		ClientID:             "ocidGaDVAixCDiFcdtO1tq8uC7Icfy9l",
		ClientSecret:         "Wm4E88PVjnL5EP3gcWrWG8x-ntxXPqUh-59N3jqs4KLDkIeBHoCEyzd5oSjjDv_F",
	}
}

func NewOpenFGAClient(ctx context.Context, config *ConfigOpenFGA, forManagement bool) (*client.OpenFgaClient, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	storeID := config.StoreID
	if forManagement {
		storeID = "" // StoreID deve ser vazio para ListStores e CreateStore
	}

	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl:               config.APIURL,
		StoreId:              storeID,
		AuthorizationModelId: config.AuthorizationModelID,
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodClientCredentials,
			Config: &credentials.Config{
				ClientCredentialsClientId:       config.ClientID,
				ClientCredentialsClientSecret:   config.ClientSecret,
				ClientCredentialsApiAudience:    config.APIAudience,
				ClientCredentialsApiTokenIssuer: config.APITokenIssuer,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create OpenFGA client: %w", err)
	}

	options := client.ClientCheckOptions{}
	body := client.ClientCheckRequest{
		User:     "user:zbMdAdazsqO11O5kYtvblC43iUi2",
		Relation: "can_view",
		Object:   "tenant:01981aab-55d2-7cc0-a033-39e166a3031f",
	}
	data, err := fgaClient.Check(context.Background()).Body(body).Options(options).Execute()
	if err != nil {
		fmt.Errorf("failed to check permissions: %w", err)
	}

	fmt.Println("Check result:", data)

	openFGAClient = fgaClient
	return fgaClient, nil
}

func WriteAuthorizationModel(ctx context.Context, fgaClient *client.OpenFgaClient, model client.ClientWriteAuthorizationModelRequest) (*client.ClientWriteAuthorizationModelResponse, error) {
	if fgaClient == nil {
		return nil, fmt.Errorf("fgaClient cannot be nil")
	}

	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	response, err := fgaClient.WriteAuthorizationModel(ctx).Body(model).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to write authorization model: %w", err)
	}

	fmt.Println("Authorization Model written successfully:", response)
	return nil, nil
}

type TupleOperationAction string

const (
	TupleOperationActionWrite  TupleOperationAction = "write"
	TupleOperationActionDelete TupleOperationAction = "delete"
)

type TupleOperationObject string

const (
	TupleOperationObjectUser   TupleOperationObject = "user"
	TupleOperationObjectGroup  TupleOperationObject = "group"
	TupleOperationObjectTenant TupleOperationObject = "tenant"
	TupleOperationObjectVault  TupleOperationObject = "vault"
	TupleOperationObjectSecret TupleOperationObject = "secret"
)

type TupleOperationRelation string
type TupleVaultOperationRelation TupleOperationRelation

const (
	TupleOperationRelationMember TupleOperationRelation = "member"
	TupleOperationRelationOwner  TupleOperationRelation = "owner"
	TupleOperationRelationViewer TupleOperationRelation = "viewer"
	TupleOperationRelationDelete TupleOperationRelation = "delete"
	TupleOperationRelationWriter TupleOperationRelation = "writer"
	TupleOperationRelationParent TupleOperationRelation = "parent"
)

const (
	TupleOperationRelationVaultDownload TupleVaultOperationRelation = "vault_download"
	TupleOperationRelationVaultCopy     TupleVaultOperationRelation = "vault_copy"
)

func (t *TupleOperationRelation) string() string {
	switch *t {
	case TupleOperationRelationMember:
		return "member"
	case TupleOperationRelationOwner:
		return "owner"
	case TupleOperationRelationViewer:
		return "viewer"
	case TupleOperationRelationDelete:
		return "delete"
	case TupleOperationRelationWriter:
		return "writer"
	case TupleOperationRelationParent:
		return "parent"
	default:
		return "unknown"
	}
}

type TupleOperation struct {
	UserType   TupleOperationObject
	UserID     string                 // O ID específico do User (ex: Firebase UID, ID do grupo, ID do tenant)
	Relation   TupleOperationRelation // A relação (ex: "member", "owner", "tenant_owner_group", "parent")
	ObjectType TupleOperationObject
	ObjectID   string               // O ID específico do Object (ex: tenantID, vaultID, secretID)
	Action     TupleOperationAction // "write" ou "delete"

}

func ManageFgaTuples(ctx context.Context, fgaClient *client.OpenFgaClient, tuples []TupleOperation) error {
	if fgaClient == nil {
		return fmt.Errorf("fgaClient cannot be nil")
	}

	if ctx.Err() != nil {
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	var writeTuples []client.ClientTupleKey
	var deleteTuples []client.ClientTupleKeyWithoutCondition
	var writeItens client.ClientWriteRequest

	for _, tuple := range tuples {
		if tuple.Action == "delete" {
			deleteTuples = append(deleteTuples, client.ClientTupleKeyWithoutCondition{
				User:     fmt.Sprintf("%s:%s", tuple.UserType, tuple.UserID),
				Relation: tuple.Relation.string(),
				Object:   fmt.Sprintf("%s:%s", tuple.ObjectType, tuple.ObjectID),
			})
		} else {
			writeTuples = append(writeTuples, client.ClientTupleKey{
				User:     fmt.Sprintf("%s:%s", tuple.UserType, tuple.UserID),
				Relation: tuple.Relation.string(),
				Object:   fmt.Sprintf("%s:%s", tuple.ObjectType, tuple.ObjectID),
			})
		}

	}

	if len(writeTuples) > 0 {
		writeItens.Writes = writeTuples
	}
	if len(deleteTuples) > 0 {
		writeItens.Deletes = deleteTuples
	}

	_, err := fgaClient.Write(ctx).Body(writeItens).Execute()
	if err != nil {
		return fmt.Errorf("failed to write tuple: %w", err)
	}

	fmt.Printf("OpenFGA: %d tuplas escritas e %d tuplas deletadas com sucesso.\n", len(writeTuples), len(deleteTuples))
	return nil
}

func ManageFgaReadTuples(ctx context.Context, fgaClient *client.OpenFgaClient, tuples []TupleOperation) (*client.ClientReadResponse, error) {
	if fgaClient == nil {
		return nil, fmt.Errorf("fgaClient cannot be nil")
	}

	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	var itens *client.ClientReadResponse
	var err error

	for _, tuple := range tuples {
		user := fmt.Sprintf("%s:%s", tuple.UserType, tuple.UserID)
		relation := tuple.Relation.string()
		object := fmt.Sprintf("%s:%s", tuple.ObjectType, tuple.ObjectID)
		itens, err = fgaClient.Read(ctx).Body(client.ClientReadRequest{
			User:     &user,
			Relation: &relation,
			Object:   &object,
		},
		).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to read tuple: %w", err)
		}
		if itens == nil {
			return nil, fmt.Errorf("load is nil")
		}
	}

	fmt.Printf("Found %d tuples\n", len(itens.Tuples))
	fmt.Printf("Getting %d tuples\n", len(itens.GetTuples()))
	return itens, nil
}

func ManageFgaCanAccess(ctx context.Context, fgaClient *client.OpenFgaClient, tuples []TupleOperation) (bool, error) {
	if fgaClient == nil {
		return false, fmt.Errorf("fgaClient cannot be nil")
	}

	if ctx.Err() != nil {
		return false, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	var itens *client.ClientCheckResponse
	var err error

	for _, tuple := range tuples {
		user := fmt.Sprintf("%s:%s", tuple.UserType, tuple.UserID)
		relation := tuple.Relation.string()
		object := fmt.Sprintf("%s:%s", tuple.ObjectType, tuple.ObjectID)
		itens, err = fgaClient.Check(ctx).Body(client.ClientCheckRequest{
			User:     user,
			Relation: relation,
			Object:   object,
		}).Execute()

		if err != nil {
			return false, fmt.Errorf("failed to read tuple: %w", err)
		}
		if itens == nil {
			return false, fmt.Errorf("load is nil")
		}
	}

	fmt.Println("Checking access for tuples:", itens.GetResolution())
	fmt.Println(itens.Resolution)
	fmt.Printf("Allowed: %v\n", *itens.Allowed)
	if itens.GetAllowed() {
		fmt.Println("Access granted")
		return true, nil
	} else {
		fmt.Println("Access denied")
		return false, nil
	}
}

func main() {

	config := NewConfigOpenFGA()
	ctx := context.Background()

	fgaClient, err := NewOpenFGAClient(ctx, config, false) // O 'true' indica que é para gerenciamento
	if err != nil {
		fmt.Println("Error creating OpenFGA management client:", err)

	}
	fmt.Println("OpenFGA management client created successfully.")

	fmt.Println("Listing Stores")
	stores1, err := fgaClient.ListStores(ctx).Execute() // Use o cliente de gerenciamento
	if err != nil {
		fmt.Println("Error listing stores:", err)
	} else { // Somente se não houver erro
		fmt.Printf("Stores Count: %d\n", len(stores1.GetStores()))
	}
	fmt.Println("OpenFGA client created successfully:", fgaClient)

	// CreateStore
	fmt.Println("Creating Test Store")
	store, err := fgaClient.CreateStore(ctx).Body(client.ClientCreateStoreRequest{Name: "lockari-develop"}).Execute() // Use o cliente de gerenciamento
	if err != nil {
		fmt.Println("Error creating test store:", err)
	}
	if store == nil {
		fmt.Println("Failed to create test store: store is nil")
	} else {
		fmt.Printf("Test Store ID: %v\n", store.Id)
	}
	// Exemplo de uso do cliente de Store (se a Store '01JZXBQMPXQB4XBDCVCJ7EMMM2' existir e estiver acessível)
	modelResponse, err := fgaClient.ReadAuthorizationModels(ctx).Execute()
	if err != nil {
		fmt.Println("Error reading models:", err)
	} else {
		fmt.Printf("Models count: %d\n", len(modelResponse.GetAuthorizationModels()))
	}

	// ListStores after Create
	fmt.Println("Listing Stores")
	stores, err := fgaClient.ListStores(ctx).Execute()
	if err != nil {
		fmt.Println("Error listing stores:", err)
	}
	if stores == nil {
		fmt.Println("Failed to list stores: stores is nil")
	} else {
		fmt.Printf("Stores Count: %d\n", len(stores.Stores))
	}
	fmt.Println("Getting Current Store")
	currentStore, err := fgaClient.GetStore(ctx).Execute()
	if err != nil {
		fmt.Println("Error getting current store:", err)
	}
	if currentStore == nil {
		fmt.Println("Failed to get current store: currentStore is nil")
	} else {
		fmt.Printf("Current Store Name: %v\n", currentStore.Name)
	}

	// ReadAuthorizationModels
	fmt.Println("Reading Authorization Models")
	models, err := fgaClient.ReadAuthorizationModels(ctx).Execute()
	if err != nil {
		fmt.Println("Error reading authorization models:", err)
	}
	if models == nil {
		fmt.Println("Failed to read authorization models: models is nil")
	} else {
		fmt.Printf("Authorization Models Count: %d\n", len(models.AuthorizationModels))
	}

	// WriteAuthorizationModel
	fmt.Println("Writing an Authorization Model")
	body := client.ClientWriteAuthorizationModelRequest{
		SchemaVersion: "1.1",
		TypeDefinitions: []openfga.TypeDefinition{
			{
				Type:      "user",
				Relations: &map[string]openfga.Userset{},
			},
			{
				Type: "tenant",
				Relations: &map[string]openfga.Userset{
					"writer": {This: &map[string]interface{}{}},
					"viewer": {Union: &openfga.Usersets{
						Child: []openfga.Userset{
							{This: &map[string]interface{}{}},
							{ComputedUserset: &openfga.ObjectRelation{
								Object:   openfga.PtrString(""),
								Relation: openfga.PtrString("writer"),
							}},
						},
					}},
				},
				Metadata: &openfga.Metadata{
					Relations: &map[string]openfga.RelationMetadata{
						"writer": {
							DirectlyRelatedUserTypes: &[]openfga.RelationReference{
								{Type: "user"},
								{Type: "user", Condition: openfga.PtrString("ViewCountLessThan2")},
							},
						},
						"viewer": {
							DirectlyRelatedUserTypes: &[]openfga.RelationReference{
								{Type: "user"},
							},
						},
					},
				},
			},
		},
	}

	authorizationModel, err := fgaClient.WriteAuthorizationModel(ctx).Body(body).Execute()
	if err != nil {
		fmt.Printf("error writing authorization model: %w", err)
	}
	fmt.Printf("Authorization Model ID: %v\n", authorizationModel.AuthorizationModelId)

	// ReadAuthorizationModels - after Write
	fmt.Println("Reading Authorization Models")
	models, err = fgaClient.ReadAuthorizationModels(ctx).Execute()
	if err != nil {
		fmt.Printf("error reading authorization models: %w", err)
	}
	fmt.Printf("Models Count: %d\n", len(models.AuthorizationModels))

	// ReadLatestAuthorizationModel - after Write
	latestAuthorizationModel, err := fgaClient.ReadLatestAuthorizationModel(ctx).Execute()
	if err != nil {
		fmt.Printf("Error reading latest authorization model: %w", err)
	}
	fmt.Printf("Latest Authorization Model ID: %v\n", (*latestAuthorizationModel.AuthorizationModel).Id)

	// Write
	fmt.Println("Writing Tuples")
	myWriteTuple := []TupleOperation{
		{
			UserType:   TupleOperationObjectUser,
			UserID:     "anne",
			Relation:   TupleOperationRelationViewer,
			ObjectType: TupleOperationObjectTenant,
			ObjectID:   "2952ab2d-d83f-756d-9397-c5ed9f3cb77a",
			Action:     TupleOperationActionWrite, // Exemplo de escrita
		}}
	myDeleteTuple := []TupleOperation{
		{
			UserType:   TupleOperationObjectUser,
			UserID:     "anne",
			Relation:   TupleOperationRelationViewer,
			ObjectType: TupleOperationObjectTenant,
			ObjectID:   "2952ab2d-d83f-756d-9397-c5ed9f3cb77a",
			Action:     TupleOperationActionDelete, // Exemplo de deleção
		},
	}

	fmt.Println("\n >>>> Writing Tuples to OpenFGA <<<<<")
	err = ManageFgaTuples(ctx, fgaClient, myWriteTuple)
	if err != nil {
		fmt.Printf("error writing tuples: %w", err)
	} else {
		fmt.Println("Tuples written successfully")
	}

	fmt.Println("\n >>>> Reading Tuples to OpenFGA <<<<<")
	tuples, err := ManageFgaReadTuples(ctx, fgaClient, myWriteTuple)
	if err != nil {
		fmt.Printf("error reading tuples: %w", err)
	} else {
		fmt.Printf("Read Tuples: %v\n", tuples)
	}

	fmt.Println("\n >>>> Checking access Tuples to OpenFGA <<<<<")
	fmt.Println(" >>>> Access granted <<<<<")
	can, err := ManageFgaCanAccess(ctx, fgaClient, myWriteTuple)
	if err != nil {
		fmt.Printf("error checking access: %w", err)
	} else {
		fmt.Printf("Can access: %v\n", can)
	}

	fmt.Println(" >>>> Access denied <<<<<")
	copyMyWriteTuple := myWriteTuple
	copyMyWriteTuple[0].Relation = TupleOperationRelationWriter
	can, err = ManageFgaCanAccess(ctx, fgaClient, copyMyWriteTuple)
	if err != nil {
		fmt.Printf("error checking access: %w", err)
	} else {
		fmt.Printf("Can access: %v\n", can)
	}

	fmt.Println("\n >>>> Deleting Tuples to OpenFGA <<<<<")
	err = ManageFgaTuples(ctx, fgaClient, myDeleteTuple)
	if err != nil {
		fmt.Printf("error deleting tuples: %w", err)
	} else {
		fmt.Println("Tuples deleted successfully")
	}

	fmt.Println(">>>> [END] Writing Tuples to OpenFGA <<<<<")
	log.Fatalln("Done Writing Tuples")
	// _, err = fgaClient.Write(ctx).Body(client.ClientWriteRequest{
	// 	Writes: []client.ClientTupleKey{
	// 		{
	// 			User:     "user:anne",
	// 			Relation: "writer",
	// 			Object:   "tenant:2192ab2a-d83f-756d-9397-c5ed9f3cb69a",
	// 			Condition: &openfga.RelationshipCondition{
	// 				Name:    "ViewCountLessThan2",
	// 				Context: &map[string]interface{}{"Name": "Roadmap", "Type": "tenant"},
	// 			},
	// 		},
	// 	},
	// }).Options(client.ClientWriteOptions{
	// 	AuthorizationModelId: &authorizationModel.AuthorizationModelId,
	// }).Execute()
	// if err != nil {
	// 	fmt.Printf("error writing tuples: %w", err)
	// }
	fmt.Println("Done Writing Tuples")

	// Set the model ID
	err = fgaClient.SetAuthorizationModelId("01K0ERJX9TJM30Q2V3AWM0A56N")
	if err != nil {
		fmt.Printf("error setting authorization model ID: %w", err)
	}

	// Read
	fmt.Println("Reading Tuples")
	readTuples, err := fgaClient.Read(ctx).Execute()
	if err != nil {
		fmt.Printf("error reading tuples: %w", err)
	}
	fmt.Printf("Read Tuples: %v\n", readTuples)

	// ReadChanges
	fmt.Println("Reading Tuple Changes")
	readChangesTuples, err := fgaClient.ReadChanges(ctx).Execute()
	if err != nil {
		fmt.Printf("error reading changes tuples: %w", err)
	}
	fmt.Printf("Read Changes Tuples: %v\n", readChangesTuples)

	// Check
	fmt.Println("Checking for access")
	failingCheckResponse, err := fgaClient.Check(ctx).Body(client.ClientCheckRequest{
		User:     "user:anne",
		Relation: "viewer",
		Object:   "tenant:2192ab2a-d83f-756d-9397-c5ed9f3cb69a",
	}).Execute()
	if err != nil {
		fmt.Printf("Failed due to: %w\n", err.Error())
	} else {
		fmt.Printf("Allowed: %v\n", failingCheckResponse.Allowed)
	}

	// Checking for access with context
	fmt.Println("Checking for access with context")
	checkResponse, err := fgaClient.Check(ctx).Body(client.ClientCheckRequest{
		User:     "user:anne",
		Relation: "viewer",
		Object:   "tenant:2192ab2a-d83f-756d-9397-c5ed9f3cb69a",
		Context:  &map[string]interface{}{"ViewCount": 2},
	}).Execute()
	if err != nil {
		fmt.Printf("Failed due to: %w\n", err.Error())
	}
	fmt.Printf("Allowed: %v\n", checkResponse.Allowed)

	// ListObjects
	fmt.Println("Listing objects user has access to")
	listObjectsResponse, err := fgaClient.ListObjects(ctx).Body(client.ClientListObjectsRequest{
		User:     "user:anne",
		Relation: "writer",
		Type:     "tenant",
		Context:  &map[string]interface{}{"ViewCount": 100},
	}).Execute()
	fmt.Printf("Response: Objects = %v\n", listObjectsResponse.Objects)
}
