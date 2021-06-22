APP=costanza
GO_ENTRY=main.go
EXECFILE=costanza

all: ${EXECFILE}

${EXECFILE}: ${GO_ENTRY}
	go build -o ${EXECFILE} .

clean:
	- go clean

rebuild: clean all

docker-build:
	docker-compose build

docker-run:
	docker-compose up

.PHONY: all clean rebuild docker-build docker-run
