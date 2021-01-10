bin/rinako: *.go
	go build -o bin/rinako

release/rinako: *.go
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o release/rinako

.PHONY: run

run: bin/rinako config.toml

release: release/rinako config.toml

clean: 
	rm bin/*