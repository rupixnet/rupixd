package database

import (
stderrors "errors"
pkgerrors "github.com/pkg/errors"
"fmt"
)

// ErrNotFound denotes that the requested item was not
// found in the database.
var ErrNotFound = stderrors.New("not found")

// IsNotFoundError checks whether an error is an ErrNotFound.
func IsNotFoundError(err error) bool {
if err == nil {
return false
}
if stderrors.Is(err, ErrNotFound) {
return true
}
cause := pkgerrors.Cause(err)
result := cause == ErrNotFound
if !result {
fmt.Printf("DEBUG IsNotFoundError: err=%T/%s cause=%T/%v result=%v\n", err, err.Error(), cause, cause, result)
}
return result
}
