package domain

type mockArchiver struct {
	archived map[string]int
}

func NewMockArchiver() *mockArchiver {
	return &mockArchiver{
		archived: map[string]int{},
	}
}

func (self *mockArchiver) ArchiveByUuid(uuid string) error {
	_, found := self.archived[uuid]
	if !found {
		return &NotFoundError{}
	}

	self.archived[uuid]++
	return nil
}
