all:	
	cd src; CGO_ENABLED=0 go install -ldflags '-X "gr"' -v gr.v1/...

install: all
	@echo

test:
	cd src;go test -race gr.v1/...

testv:
	cd src;go test -v -race gr.v1/...

clean:
	cd src;go clean -i ./...; cd ..; rm -rf pkg bin

