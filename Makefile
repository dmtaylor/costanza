APP=costanza
GO_ENTRY=main.go
EXECFILE=costanza
DB_FILE=db.sqlite3
SQL_DIR=sql
SQL_FILES=${SQL_DIR}/create_tables.sql ${SQL_DIR}/load_quotes.sql

all: ${EXECFILE} ${DB_FILE}

db: ${DB_FILE}

${DB_FILE}: ${SQL_FILES}
	- rm -f ${DB_FILE} # remove any existing database & build from scratch
	sqlite3 ${DB_FILE} < ${SQL_DIR}/create_tables.sql
	sqlite3 ${DB_FILE} < ${SQL_DIR}/load_quotes.sql

${EXECFILE}: ${GO_ENTRY}
	go build -o ${EXECFILE} .

clean:
	- go clean
	- rm ${DB_FILE}

rebuild: clean all

docker-build: ${DB_FILE}
	docker-compose build

docker-run: ${DB_FILE}
	docker-compose up

.PHONY: all clean rebuild docker-build docker-run
