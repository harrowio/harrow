package stores

import "github.com/harrowio/harrow/domain"

type LogStore interface {
	FindByOperationUuid(uuid string, tepy string) (*domain.Loggable, error)
	FindByRange(uuid string, tepy string, from, to int) (*domain.Loggable, error)
	PersistLogLine(operationUuid string, logLine *domain.LogLine) error
	OnFinished(operationUuid string) error
}
