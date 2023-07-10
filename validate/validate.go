package validate

import (
	"fmt"

	"github.com/jackc/errortree"
)

type Validator struct {
	e *errortree.Node
}

func New() *Validator {
	v := &Validator{
		e: &errortree.Node{},
	}
	return v
}

func (v *Validator) Add(attr string, err error) {
	v.e.Add([]any{attr}, err)
}

type PresenceError struct {
	attr string
}

func (e PresenceError) Error() string {
	return fmt.Sprintf("%s cannot be blank", e.attr)
}

func (v *Validator) Presence(attr string, value string) {
	if value == "" {
		v.e.Add([]any{attr}, PresenceError{attr: attr})
	}
}

type MinLengthError struct {
	attr      string
	minLength int
}

func (e MinLengthError) Error() string {
	return fmt.Sprintf("%s must have a minimum length of %d", e.attr, e.minLength)
}

func (v *Validator) MinLength(attr string, value string, minLength int) {
	if len(value) < minLength {
		v.e.Add([]any{attr}, MinLengthError{attr: attr, minLength: minLength})
	}
}

func (v *Validator) Err() error {
	if len(v.e.AllErrors()) == 0 {
		return nil
	}

	return v.e
}
