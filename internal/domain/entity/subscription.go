package entity

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	ID          int
	ServiceName string
	Price       int
	UserId      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
