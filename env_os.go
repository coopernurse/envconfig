// +build appengine go1.5

package envconfig

import "os"

//nolint:gochecknoglobals
var lookupEnv = os.LookupEnv
