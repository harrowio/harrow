package stores

type DataSource interface {
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
}
