package utils

import "errors"

var ErrNotReady = errors.New("not ready")
var ErrDisabled = errors.New("disabled")
var ErrNotAvailable = errors.New("not available")
var ErrNotFound = errors.New("not found")
var ErrUnknownFormat = errors.New("unknown format")
var ErrUpdateInProgress = errors.New("database update is already in progress")
