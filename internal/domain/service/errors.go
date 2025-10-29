package service

import "fmt"

var ErrInvalidPrice = fmt.Errorf("price must be a non-negative value")
var ErrInvalidServiceName = fmt.Errorf("service name cannot be empty")
var ErrSubscriptionNotFound = fmt.Errorf("subscription not found")
var ErrInvalidDateRange = fmt.Errorf("end date of the period cannot be earlier than the start date")
var ErrInvalidUserId = fmt.Errorf("invalid user id")

func GetFail(originalError error) error {
	return fmt.Errorf("failed to get subscription: %w", originalError)
}

func ListFail(originalError error) error {
	return fmt.Errorf("failed to list subscriptions: %w", originalError)
}

func UpdateFail(originalError error) error {
	return fmt.Errorf("failed to update subscription: %w", originalError)
}

func DeletionFail(originalError error) error {
	return fmt.Errorf("failed to delete subscription: %w", originalError)
}

func TotalCostFail(originalError error) error {
	return fmt.Errorf("failed to calculate total cost: %w", originalError)
}
