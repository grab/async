package core

import (
	"reflect"
	"strings"
)

func swallowErrPlanExecutionEndingEarly(err error) error {
	// Execution was intentionally ended by clients
	if err == ErrPlanExecutionEndingEarly {
		return nil
	}

	return err
}

func extractFullNameFromValue(v any) string {
	if reflect.ValueOf(v).Kind() == reflect.Pointer {
		t := reflect.ValueOf(v).Elem().Type()
		return extractFullNameFromType(t)
	}

	t := reflect.TypeOf(v)
	return extractFullNameFromType(t)
}

func extractFullNameFromType(t reflect.Type) string {
	return t.PkgPath() + "/" + t.Name()
}

func extractShortName(fullName string) string {
	shortNameIdx := strings.LastIndex(fullName, "/")
	return fullName[shortNameIdx+1:]
}
