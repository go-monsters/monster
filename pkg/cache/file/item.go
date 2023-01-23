package file

import "time"

type Item struct {
	Data       interface{}
	LastAccess time.Time
	Expired    time.Time
}
