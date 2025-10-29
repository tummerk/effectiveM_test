package persistent

import (
	"fmt"
)

func createSubscriptionError(err error) error {
	return fmt.Errorf("create Subscription Error: %w", err)
}

func getByIdError(err error) error {
	return fmt.Errorf("get by id Subscription Error: %w", err)
}

func deleteSubscriptionError(err error) error {
	return fmt.Errorf("delete Subscription Error: %w", err)
}

func calculatingTotalCostError(err error) error {
	return fmt.Errorf("calculating Total Cost Error: %w", err)
}

func scanningRowError(err error) error {
	return fmt.Errorf("failed to scan subscription row: %w", err)
}

func iterationRowsError(err error) error {
	return fmt.Errorf("error during rows iteration: %w", err)
}

func updateSubscriptionError(err error) error {
	return fmt.Errorf("update Subscription Error: %w", err)
}
