bin/gobeanstalkd: *.go
	go build -o bin/gobeanstalkd .

test:
	./tests/tests
