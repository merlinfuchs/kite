package store

import "errors"

var ErrNotFound = errors.New("not found")
var ErrAlreadyExists = errors.New("already exists")
var ErrOutdated = errors.New("outdated")
var ErrLimitReached = errors.New("limit reached")
