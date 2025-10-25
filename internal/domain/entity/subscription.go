package entity

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	serviceName string
	price       int
	userId      uuid.UUID
	startDate   time.Time
}
