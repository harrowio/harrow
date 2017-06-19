package http

import (
	"encoding/json"
	"sort"

	"github.com/harrowio/harrow/domain"
)

type CardsByProjectUuid struct {
	Cards map[string]*domain.ProjectCard
	Order []string
}

func (self *CardsByProjectUuid) Len() int {
	return len(self.Cards)
}

func (self *CardsByProjectUuid) Less(i int, j int) bool {
	a := self.Cards[self.Order[i]]
	b := self.Cards[self.Order[j]]
	return a.ProjectName < b.ProjectName
}

func (self *CardsByProjectUuid) Swap(i int, j int) {
	self.Order[i], self.Order[j] = self.Order[j], self.Order[i]
}

func (self *CardsByProjectUuid) UnmarshalJSON(data []byte) error {
	wrapper := struct {
		ProjectCards *map[string]*domain.ProjectCard
	}{&self.Cards}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}

	for projectUuid, _ := range self.Cards {
		self.Order = append(self.Order, projectUuid)
	}

	sort.Sort(self)

	return nil
}

func (self *CardsByProjectUuid) ProjectCards() []*domain.ProjectCard {
	result := []*domain.ProjectCard{}
	for _, cardId := range self.Order {
		result = append(result, self.Cards[cardId])
	}
	return result
}
