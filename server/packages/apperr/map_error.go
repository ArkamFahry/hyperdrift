package apperr

import (
	"errors"
	"fmt"
	"strings"
)

type Map map[string][]error

func (m Map) Get(key string) []string {
	if errs, ok := m[key]; ok {
		errStrings := make([]string, len(errs))
		for i, err := range errs {
			errStrings[i] = err.Error()
		}
		return errStrings
	}

	return nil
}

func (m *Map) Has(key string) bool {
	_, ok := (*m)[key]
	return ok
}

func (m *Map) Set(key string, msg interface{}) {
	if *m == nil {
		*m = make(Map)
	}

	var errs []error
	switch msg := msg.(type) {
	case error:
		if msg != nil {
			errs = append(errs, msg)
		}

	case string:
		errs = append(errs, errors.New(msg))

	default:
		panic("want error or string message")
	}

	if len(errs) > 0 {
		(*m)[key] = append((*m)[key], errs...)
	}
}

func (m Map) Error() string {
	if m == nil {
		return "<nil>"
	}

	var allErrors []string
	for key, errs := range m {
		errStrings := make([]string, len(errs))
		for i, err := range errs {
			errStrings[i] = err.Error()
		}
		allErrors = append(allErrors, fmt.Sprintf("%s: [%s]", key, strings.Join(errStrings, ", ")))
	}

	return strings.Join(allErrors, "; ")
}

func (m Map) String() string {
	return m.Error()
}

func (m Map) MarshalJSON() ([]byte, error) {
	errs := make([]string, 0, len(m))
	for key, errList := range m {
		errStrings := make([]string, len(errList))
		for i, err := range errList {
			errStrings[i] = err.Error()
		}
		errs = append(errs, fmt.Sprintf("%q: [%s]", key, strings.Join(errStrings, ", ")))
	}

	return []byte(fmt.Sprintf("{%s}", strings.Join(errs, ", "))), nil
}
