package validate

import "fmt"

type Errors map[string][]error

func (e Errors) Add(attr string, err error) {
	e[attr] = append(e[attr], err)
}

func (e Errors) Error() string {
	if len(e) == 0 {
		return "No errors"
	}

	return fmt.Sprintf("%#v", e)
}

func (e Errors) Get(attr string) []error {
	if e == nil {
		return nil
	}

	return e[attr]
}

type PresenceError struct {
	attr string
}

func (e PresenceError) Error() string {
	return fmt.Sprintf("%s cannot be blank", e.attr)
}

type MinLengthError struct {
	attr      string
	minLength int
}

func (e MinLengthError) Error() string {
	return fmt.Sprintf("%s must have a minimum length of %d", e.attr, e.minLength)
}
