package domain

type Watchable interface {
	Id() string
	WatchableType() string
	WatchableEvents() []string
}
