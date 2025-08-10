package database_utils

import (
	"github.com/moroz/uuidv7-go"
	"time"
)

func GenerateID() string {
	return uuidv7.Generate().String()
}

func Now() *time.Time {
	now := time.Now()
	return &now
}
