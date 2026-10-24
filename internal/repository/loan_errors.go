package repository

import "errors"

var (
	ErrCopyNotAvailable = errors.New("copy is not available")
	ErrLoanLimitReached = errors.New("loan limit reached")
	ErrAlreadyReturned  = errors.New("loan is already returned")
)

// translateLoanError keeps our own sentinel errors unchanged and translates everything else.
func translateLoanError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrNotFound),
		errors.Is(err, ErrInvalidReference),
		errors.Is(err, ErrCopyNotAvailable),
		errors.Is(err, ErrLoanLimitReached),
		errors.Is(err, ErrAlreadyReturned):
		return err
	}
	return translateError(err)
}
