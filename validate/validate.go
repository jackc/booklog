package validate

type Validator struct {
	e Errors
}

func New() *Validator {
	v := &Validator{}
	v.e = make(Errors)
	return v
}

func (v *Validator) Add(attr string, err error) {
	v.e.Add(attr, err)
}

func (v *Validator) Presence(attr string, value string) {
	if value == "" {
		v.e.Add(attr, PresenceError{attr: attr})
	}
}

func (v *Validator) MinLength(attr string, value string, minLength int) {
	if len(value) < minLength {
		v.e.Add(attr, MinLengthError{attr: attr, minLength: minLength})
	}
}

func (v *Validator) Err() error {
	if len(v.e) == 0 {
		return nil
	}

	return v.e
}
