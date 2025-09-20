package utils

import "github.com/google/uuid"

func NewID() string {
	id, _ := uuid.NewV7()
	return id.String()
}

func ParseID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

func IsID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
