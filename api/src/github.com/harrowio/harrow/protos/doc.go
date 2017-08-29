//go:generate protoc -I ./src/github.com/harrowio/harrow/protos ./src/github.com/harrowio/harrow/protos/limits.proto --go_out=plugins=grpc:src/github.com/harrowio/harrow/protos

package protos
