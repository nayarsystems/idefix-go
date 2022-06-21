package idefixgo

import "errors"

// ErrTimeout gets returned when an operation times out
var ErrTimeout = errors.New("resource not found")

// ErrContextClosed generic context closed error
var ErrContextClosed = errors.New("context closed")

// ErrChannelClosed generic channel closed error
var ErrChannelClosed = errors.New("channel closed")

// ErrMarshall marshall error
var ErrMarshall = errors.New("marshall error")
