package helpers

import (
	"errors"
	"github.com/spf13/cobra"
)
// Flag represents a command line flag and its value
type Flag[T any] struct {
	Name  string
	Value T
}

// GetGenerateFlagValues gets provider, candidates, and temperature flags from the generate command
func SetGenerateFlagValues[T1, T2, T3 any](cmd *cobra.Command, provider, candidates, temperature string) (T1, T2, T3, error) {

	value1, err := getFlagValue[T1](cmd, provider)
	if err != nil {
		value1 = any("gemini").(T1)
	}

	value2, err := getFlagValue[T2](cmd, candidates)
	if err != nil {
		value2 = any(2).(T2)
	}

	value3, err := getFlagValue[T3](cmd, temperature)
	if err != nil {
		value3 = any(0.5).(T3)
	}

	return value1, value2, value3, nil
}

// getFlagValue gets a single flag value with type conversion
func getFlagValue[T any](cmd *cobra.Command, flagName string) (T, error) {
	var typedValue T
	switch any(typedValue).(type) {
	case string:
		value, err := cmd.Flags().GetString(flagName)
		if err != nil {
			var zero T
			return zero, err
		}
		return any(value).(T), nil
	case int:
		value, err := cmd.Flags().GetInt(flagName)
		if err != nil {
			var zero T
			return zero, err
		}
		return any(value).(T), nil
	case float32:
		value, err := cmd.Flags().GetFloat32(flagName)
		if err != nil {
			var zero T
			return zero, err
		}
		return any(value).(T), nil
	case float64:
		value, err := cmd.Flags().GetFloat64(flagName)
		if err != nil {
			var zero T
			return zero, err
		}
		return any(value).(T), nil
	case bool:
		value, err := cmd.Flags().GetBool(flagName)
		if err != nil {
			var zero T
			return zero, err
		}
		return any(value).(T), nil
	default:
		var zero T
		return zero, errors.New("unsupported flag type")
	}
}
