
# Costanza

A discord bot implementing quote responses and a dice notation roller & expression parser.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/dmtaylor/costanza)
[![Apache 2](https://img.shields.io/github/license/dmtaylor/costanza)](https://github.com/dmtaylor/costanza/LICENSE)
![Release](https://img.shields.io/github/v/release/dmtaylor/costanza?include_prereleases&sort=semver)
![Build](https://img.shields.io/github/actions/workflow/status/dmtaylor/costanza/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dmtaylor/costanza)](https://goreportcard.com/report/github.com/dmtaylor/costanza)

## Getting started
This project requires [go](https://golang.org/) and postgres.

To run locally, run the commands after setting up `config.toml` file and postgres:
```
$ go get
$ go build
$ ./costanza listen
```

You can also build the binary by running `make` or `mage build`.

Database migrations are contained under `migrations`. [go-migrate](https://github.com/golang-migrate/migrate) can be used to run them against the
DB (and is used by docker-compose to run the migrations in the containerized setup), but any migration tool supporting up/down operations should be fine.

To spin it up in a container (including the DB & running all migrations), use `docker-compose up`. Naturally you will need Docker &
docker-compose installed for this. Running via docker-compose requires directories `/var/costanza/db_data` & `/var/costanza_dev/db_data`
for the Postgres volume mounts. Using the Makefile targets for docker-compose based deployments should ensure these directories are created as needed.

There is a Makefile & [magefiles](https://magefile.org) for commonly used targets.

## Usage
Costanza has the following subcommands:
- listen: listen to incoming Discord events & respond appropriately. This is the main mode of operation
- register: registers the slash commands for the application.
- roll: runs the dice roller using the positional arguments. This is useful for testing out changes to the parser on the command line
- quote: prints a quote to stdout. This is useful for testing changes to quote retrieval.
- cfg: loads configuration from environment. This is useful in debugging issues loading configuration.
- report: send usage reports for the given month to the configured channels
  - remove: remove usage status for the given month from the db

The following behaviors are present in listen mode:
- If Costanza is @-ed, it will respond with a random quote from a slightly curated list of George Costanza quotes
- If a user posts between 12:30 AM & 6:00 AM & their user ID is included in `INSOMNIAC_IDS` or they have a role listed in `INSOMNIAC_ROLES`, they get a gentle reminder to sleep
- A welcome message is sent when a user joins the guild.

Costanza has these slash commands:
- `/chelp`: sends brief usage details.
- `/roll {roll value}`: argument text is parsed as a d-notation roll and evaluated.
- `/srroll {roll value}`: argument text is parsed & evaluated as d-notation, and the resulting value is run as a Shadowrun roll.
- `/wodroll {roll value} [chance] [9again] [8again]`: argument text is parsed and evaluated as d-notation, and the resulting value is run as a World of Darkness roll. Optional arguments indicate
if the roll is a chance die, has 8-again, or 9-again. Rolls of < 1 dice are ran as chance rolls.
- `/dhtest {roll value}`: argument text is parsed and evaluated as d-notation, and the resulting value is run as a Dark Heresy/Fantasy Flight Warhammer 40k
RPG skill test (i.e. over or under 1d100)
- `/weather [location]`: gets current weather conditions for given location, or defaults from config file. Uses [wttr.in](https://wttr.in/) for weather data.

## Environment Variables

To run this project, you will need to have a config file named `config.toml` present either in the active directory or
in `/etc/costanza` conforming to the format in `example.config.toml`. It's recommended to copy that file & replace appropriate
values there.

Additionally, for docker-compose compatability reasons, the environment variable `COSTANZA_DB_URL` will overwrite the connection
string config for the Postgres connection. It's recommended to leave this variable unset to avoid confusion.

## Troubleshooting
- docker-compose mounts the postgres db under `/var/costanza/db_data` or `/var/costanza_dev/db_data`.
If you're encountering issues with the db service, check that those directories exist and have proper permissions.
- If you're having trouble getting the application to connect to the DB, verify that the environment variable `COSTANZA_DB_URL`
is unset.

## TODO
- Add initiative tracking system
- Add rolling types for other popular systems (Savage Worlds?)
    - Dark Heresy/FF 40k damage rolls
- Figure out a good way to print chained rolls that shows intermediate results
- Curate the quote list a bit more
- Add more interesting responses to bad rolls

## Why
Constant imposter syndrome mostly.

I named it Costanza because if there was a modern day saint of having the most rotten luck, it would be George. Plus, most of
my experiences as a player in D&D involve some Costanza-like decision-making.
