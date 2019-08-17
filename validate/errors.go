package validate

import (
	"fmt"
	"strings"
)

type Errors map[string][]error

func (e Errors) Add(attr string, err error) {
	e[attr] = append(e[attr], err)
}

func (e Errors) Error() string {
	if len(e) == 0 {
		return "No errors"
	}

	sb := &strings.Builder{}

	join := false
	for attr, errs := range e {
		for _, err := range errs {
			if join {
				sb.WriteString(" and ")
			}
			fmt.Fprintf(sb, "%s %v", attr, err)
			join = true
		}
	}

	return sb.String()
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
