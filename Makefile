GO_TEST_ARGS := ""
FILE_PATTERN := 'go\|Makefile'

test:
	go test ./... $(GO_TEST_ARGS)

test_watch:
	find . | grep $(FILE_PATTERN) | entr bash -c 'clear; make test'
