package storage

import "github.com/ovatu/redalert/events"

type EventStorage interface {
	Store(*events.Event) error
	Last() (*events.Event, error)
	GetRecent() ([]*events.Event, error)
}
