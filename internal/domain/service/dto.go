package service

import (
	"time"

	"github.com/google/uuid"
)

type CreateSubscriptionInput struct {
	ServiceName string
	Price       int
	UserId      uuid.UUID
	StartDate   time.Time
	EndDate     *time.Time
}

type UpdateSubscriptionInput struct {
	ServiceName      *string
	Price            *int
	UserId           *uuid.UUID
	StartDate        *time.Time
	SetEndDateToNull bool
	EndDate          *time.Time
}

type TotalCostInput struct {
	StartDate   time.Time
	EndDate     time.Time
	UserId      *uuid.UUID
	ServiceName *string
}
