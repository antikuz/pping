package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingResultContainError(t *testing.T) {
	zeroReceivedErr := errors.New("1 packets transmitted, 0 received, 100% packet loss, time 0ms")
	timedOutErr := errors.New("Request timed out.")
	unreachableErr := errors.New("Reply from 127.0.0.1: Destination host unreachable.")
	unknownDnsErr := errors.New("Ping request could not find host google.com. Please check the name and try again.")

	assert.Equal(t, false, pingResultContainError(zeroReceivedErr))
	assert.Equal(t, false, pingResultContainError(timedOutErr))
	assert.Equal(t, false, pingResultContainError(unreachableErr))
	assert.Equal(t, true, pingResultContainError(unknownDnsErr))
}
