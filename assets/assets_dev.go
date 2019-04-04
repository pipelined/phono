// +build dev

package assets

import "net/http"

var Assets = http.Dir("../web")
