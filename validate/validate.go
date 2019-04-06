package validate

type Validator struct {
	e Errors
}

func New() *Validator {
	v := &Validator{}
	v.e = make(Errors)
	return v
}

func (v *Validator) Presence(attr string, value string) {
	if value == "" {
		v.e.Add(attr, PresenceError{attr: attr})
	}
}

func (v *Validator) Err() error {
	if len(v.e) == 0 {
		return nil
	}

	return v.e
}
