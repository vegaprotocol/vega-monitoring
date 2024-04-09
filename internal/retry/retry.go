package retry

import (
	"fmt"
	"time"
)

func RetryReturn[K any](retryAmount int, retryDelay time.Duration, handler func() (K, error)) (K, error) {
	if retryAmount < 1 {
		retryAmount = 1
	}
	if retryDelay < 1 {
		retryDelay = 200 * time.Millisecond
	}

	var (
		allErrors []string
		lastError error
	)

	for i := 0; i < retryAmount; i++ {
		res, err := handler()
		if err != nil {
			lastError = err
			allErrors = append(allErrors, err.Error())
			time.Sleep(retryDelay)
			continue
		}

		return res, nil
	}

	var result K
	return result, fmt.Errorf("failed to run handler for %d times, last error: %w, all errors: %v", retryAmount, lastError, allErrors)
}

func RetryRun(retryAmount int, retryDelay time.Duration, handler func() error) error {
	if retryAmount < 1 {
		retryAmount = 1
	}
	if retryDelay < 1 {
		retryDelay = 200 * time.Millisecond
	}

	var (
		allErrors []string
		lastError error
	)

	for i := 0; i < retryAmount; i++ {
		if err := handler(); err != nil {
			lastError = err
			allErrors = append(allErrors, err.Error())
			time.Sleep(retryDelay)
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to run handler for %d times, last error: %w, all errors: %v", retryAmount, lastError, allErrors)
}
