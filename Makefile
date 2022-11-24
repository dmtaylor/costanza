APP=costanza
GO_ENTRY=main.go
EXECFILE=costanza

all: ${EXECFILE}

${EXECFILE}: ${GO_ENTRY}
	go build -o ${EXECFILE} .

clean:
	- go clean
	- rm ${DB_FILE}

rebuild: clean all

docker-build: tests
	docker-compose build

docker-run: tests
	docker-compose up

docker-restart: tests
	docker-compose build --no-cache
	docker-compose up --build --force-recreate --no-deps -d

tests: all
	go test -v ./...

.PHONY: all clean rebuild docker-build docker-run docker-restart tests
