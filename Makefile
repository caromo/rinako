bin/rinako: *.go
	go build -o bin/rinako

.PHONY: run

run: bin/rinako config.toml

clean: 
	rm bin/*