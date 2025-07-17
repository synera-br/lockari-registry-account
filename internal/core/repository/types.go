package repository

import (
	"context"
	"errors"
	"fmt"
)

func SetCollection(ctx context.Context, collection string) (*string, error) {

	if collection == "" {
		return nil, errors.New("collection is empty")
	}

	if ctx == nil {
		return nil, errors.New("context is nil")
	}

	fmt.Println("\nSetting collection:", collection)
	fmt.Println("\nContext UserID:", ctx.Value("token"))
	var col string

	if ctx.Value("token") == nil {
		return nil, errors.New("user id is nil")
	} else {
		userID := ctx.Value("token").(string)
		if userID == "" {
			return nil, errors.New("user id is empty")
		}
		col = fmt.Sprintf("tenant/%s/%s", userID, collection)
	}

	return &col, nil
}

func GetCollection(tenantid, collection *string) (*string, error) {
	if tenantid == nil || collection == nil {
		return nil, errors.New("tenantid or collection is nil")
	}
	if *tenantid == "" || *collection == "" {
		return nil, errors.New("tenantid or collection is empty")
	}
	col := fmt.Sprintf("tenants/%s/%s", *tenantid, *collection)
	return &col, nil
}
