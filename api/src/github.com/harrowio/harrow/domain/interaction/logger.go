package interaction

type Logger interface {
	Infof(pattern string, args ...interface{}) error
	Errf(pattern string, args ...interface{}) error
}
