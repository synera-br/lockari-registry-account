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
	fmt.Println("\nContext UserID:", ctx.Value("UserID"))
	var col string

	if ctx.Value("UserID") == nil {
		return nil, errors.New("user id is nil")
	} else {
		userID := ctx.Value("UserID").(string)
		if userID == "" {
			return nil, errors.New("user id is empty")
		}
		col = fmt.Sprintf("tenant/%s/%s", userID, collection)
	}

	return &col, nil
}
