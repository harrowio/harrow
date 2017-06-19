PACKAGES_WITH_TESTS = $(shell find '${GOPATH}/src' -path '*vendor*' -prune -o -name '*_test.go' -exec dirname {} \; | sort -u | sed 's\#${GOPATH}/src/\#\#' )

t:
	HARROW_ENV=test go test $(TEST_FLAGS) -run $(t) github.com/harrowio/harrow/$(p)

tt:
	HARROW_ENV=test go test -v $(TEST_FLAGS) -run $(t) github.com/harrowio/harrow/$(p)

test: migrate-test
	HARROW_ENV=test go test $(TEST_FLAGS) $(PACKAGES_WITH_TESTS)

ttest: migrate-test
	HARROW_ENV=test go test -v $(TEST_FLAGS) $(PACKAGES_WITH_TESTS)

migrate-test: ${mkfile_dir}/bin/harrow-debug
	HARROW_ENV=test ${mkfile_dir}/bin/harrow-debug migrate

list-testable-packages:
	@echo ${PACKAGES_WITH_TESTS} | tr ' ' '\n'

.PHONY: t tt test ttest migrate-test
