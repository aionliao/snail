all:	
	cd src; CGO_ENABLED=0 go install -ldflags '-X "gr"' -v bzcom/...

install: all
	@echo

test:
	cd src;go test -race bzcom/...

testv:
	cd src;go test -v -race bzcom/...

clean:
	cd src;go clean -i ./...; cd ..; rm -rf pkg bin

