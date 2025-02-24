package uuidutil

import "github.com/google/uuid"

type UUIDUtil interface {
	Generate() string
}

type uuidUtil struct{}

func (u uuidUtil) Generate() string {
	return uuid.NewString()
}

func NewUUIDUtil() UUIDUtil {
	return &uuidUtil{}
}
