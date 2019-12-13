// +build !appengine,!go1.5

package envconfig

import "syscall"

//nolint:gochecknoglobals
var lookupEnv = syscall.Getenv
