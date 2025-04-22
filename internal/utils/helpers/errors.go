package helpers

// ErrNoStagedChanges represents an error when no staged changes are found
type ErrNoStagedChanges struct{}

func (e ErrNoStagedChanges) Error() string {
	return "no staged changes found"
}

// IsErrNoStagedChanges checks if an error is of type ErrNoStagedChanges
func IsErrNoStagedChanges(err error) bool {
	_, ok := err.(ErrNoStagedChanges)
	return ok
}