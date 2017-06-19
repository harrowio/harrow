package domain

type MapPayload map[string]string

func (self MapPayload) Get(key string) string { return self[key] }
