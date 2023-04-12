APP=costanza
GO_ENTRY=main.go
EXECFILE=costanza
DEV_DB_DIR=/var/costanza_dev/db_data
PROD_DB_DIR=/var/costanza/db_data

all: ${EXECFILE}

${EXECFILE}: ${GO_ENTRY} FORCE
	go build -o ${EXECFILE} .

clean:
	- go clean

rebuild: clean all

docker-build-dev: tests
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml -p costanza-dev build

docker-run-dev: tests ${DEV_DB_DIR}
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml -p costanza-dev up

docker-restart-dev: tests ${DEV_DB_DIR}
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml -p costanza-dev -- build --no-cache
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml -p costanza-dev -- up --build --force-recreate --no-deps -d

docker-build-prod: tests
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml -p costanza build

docker-run-prod: tests ${PROD_DB_DIR}
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml -p costanza up

docker-restart-prod: tests ${PROD_DB_DIR}
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml -p costanza -- build --no-cache
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml -p costanza -- up --build --force-recreate --no-deps -d
tests: all
	go test -v ./...

${DEV_DB_DIR}:
	sudo mkdir --parents --verbose ${DEV_DB_DIR}

${PROD_DB_DIR}:
	sudo mkdir --parents --verbose ${PROD_DB_DIR}

FORCE:

.PHONY: all clean rebuild docker-build-dev docker-build-prod docker-run-dev docker-run-prod docker-restart-dev docker-restart-prod tests FORCE
