package server

import (
	"effectiveMobile_test/internal/domain/entity"
	"effectiveMobile_test/internal/server/api"
	"effectiveMobile_test/pkg/utils"
	"github.com/google/uuid"
)

func toAPISub(s *entity.Subscription) api.Subscription {
	var endPtr *string
	if s.EndDate != nil {
		ed := utils.FormatMonthYear(*s.EndDate)
		endPtr = &ed
	}

	return api.Subscription{
		Id:          s.ID,
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserId:      (uuid.UUID)(s.UserId),
		StartDate:   utils.FormatMonthYear(s.StartDate),
		EndDate:     endPtr,
		CreatedAt:   &s.CreatedAt,
		UpdatedAt:   &s.UpdatedAt,
	}
}
