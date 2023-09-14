package testutil

// FiberTestTimeoutMillis specifies the timeout for a test request to a Fiber
// HTTP server.
//
// When tests are run with the race detector, the default 1-second
// timeout for a Fiber request may be exceeded.
const FiberTestTimeoutMillis = 10000
