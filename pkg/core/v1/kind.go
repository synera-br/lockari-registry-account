package corev1

import (
	"errors"

	"lockari-api-application/pkg/utils"
)

type Kind string

const (
	KindVault    Kind = "vault"
	KindSecret   Kind = "secret"
	KindKeyValue Kind = "key_value"
)

type ObjectKind struct {
	Kind Kind   `json:"kind"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (k *Kind) IsValid() error {
	if k == nil {
		return errors.New("invalid kind: kind cannot be nil")
	}

	switch *k {
	case KindVault, KindSecret, KindKeyValue:
		return nil
	default:
		return errors.New("invalid kind: must be one of vault, secret, or key_value")
	}
}

func (k *Kind) String() string {
	if k == nil {
		return ""
	}
	return string(*k)
}

func (k *ObjectKind) GetKind() Kind {
	if k == nil {
		return ""
	}
	return k.Kind
}

func (k *ObjectKind) GenerateID() error {
	if k == nil {
		return errors.New("invalid object kind: object kind cannot be nil")
	}

	k.ID = utils.GenerateIDv7()

	if k.ID == "" {
		return errors.New("invalid object kind: failed to generate ID")
	}

	if err := k.Kind.IsValid(); err != nil {
		return err
	}
	return nil
}

func (k *ObjectKind) IsValid() error {
	if k == nil {
		return errors.New("invalid object kind: object kind cannot be nil")
	}

	if err := k.Kind.IsValid(); err != nil {
		return err
	}

	if k.ID == "" {
		return errors.New("invalid object kind: id is required")
	}

	if k.Name == "" {
		return errors.New("invalid object kind: name is required")
	}

	return nil
}
