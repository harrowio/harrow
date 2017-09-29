mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

HARROW_ENV=development

debug: BUILD_OPTS+=-tags debug
debug: TARGET_NAME=${mkfile_dir}/bin/harrow-debug
debug: build

${mkfile_dir}/bin/harrow-debug: debug
${mkfile_dir}/bin/harrow-release: release

release: BUILD_OPTS+=-tags release -ldflags="-s -w"
release: TARGET_NAME=${mkfile_dir}/bin/harrow-release
release: build
	AWS_PROFILE=harrow-build-deploy aws s3api put-object --bucket harrow-builds --key backend/harrow-$(shell git rev-parse --short HEAD) --body ./bin/harrow-release

build:
	go generate -x github.com/harrowio/harrow/cmd/...
	go build -i ${BUILD_OPTS} -o ${TARGET_NAME} github.com/harrowio/harrow/cmd/harrow

include ${mkfile_dir}/test.mk

.PHONY: debug release build
