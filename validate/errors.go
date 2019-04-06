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

	return "TODO"
}

type PresenceError struct {
	attr string
}

func (e PresenceError) Error() string {
	return fmt.Sprintf("%s cannot be blank", e.attr)
}
