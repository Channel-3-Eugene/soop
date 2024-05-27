package soop

import "time"

const WORKERS = 3
const TIMEOUT = time.Duration(10 * time.Millisecond)

// Define TestInputType and TestOutputType
type TestInputType struct {
	c int
}

type TestOutputType struct {
	n int
}
