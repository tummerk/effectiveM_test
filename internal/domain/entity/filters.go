package entity

import (
	"github.com/google/uuid"
	"time"
)

// понятно что хранить в entity фильтры глупо, сделал для упрощения
type ListFilter struct {
	UserId      *uuid.UUID
	ServiceName *string
}

type CostFilter struct {
	DateStart   time.Time
	DateEnd     time.Time
	UserId      *uuid.UUID
	ServiceName *string
}
