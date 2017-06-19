package http

type Getter interface {
	Get(key string) string
}
