package common

import (
	"net/http"
	"time"
)

// HTTPClient bounds every API call so an unresponsive endpoint can't hang terraform forever.
var HTTPClient = &http.Client{Timeout: 60 * time.Second}
