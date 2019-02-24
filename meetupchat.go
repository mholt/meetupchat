package meetupchat

import "time"

// Message represents a chat message.
type Message struct {
	From string
	Body string
	Time time.Time
}
