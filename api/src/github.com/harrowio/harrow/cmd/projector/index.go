package projector

type Index interface {
	Update(func(tx IndexTransaction) error) error
}

type IndexTransaction interface {
	Get(uuid string, dest interface{}) error
	Put(uuid string, src interface{}) error
}
