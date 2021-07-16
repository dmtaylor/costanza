
# Costanza

A discord bot implementing quote responses and a dice notation roller & expression parser.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/dmtaylor/costanza)
[![Apache 2](https://img.shields.io/github/license/dmtaylor/costanza)](https://github.com/dmtaylor/costanza/LICENSE)

## Getting started
This project requires [go](https://golang.org/).

To run locally, run the commands after setting up `.env` file:
```
$ go get
$ go build
$ ./costanza listen
```

You can also build the binary by running `make`.

To spin it up in a container, use `docker-compose up`. Naturally you will need Docker & docker-compose installed for this.

## Usage
Costanza has the following subcommands:
- listen: listen to incoming Discord events & respond appropriately. This is the main mode of operation
- roll: runs the dice roller using the positional arguments. This is useful for testing out changes to the parser on the command line

The following behaviors are present in listen mode:
- If Costanza is @-ed, it will respond with a random quote from a slightly curated list of George Costanza quotes
- If a user posts between 12:30 AM & 6:00 AM & their user ID is included in `INSOMNIAC_IDS`, they get a gentle reminder to sleep
- If a message is prefixed with `!roll`, the text following is parsed as a d-notation roll and evaluated.
- If a message is prefixed with `!srroll`, the text following is parsed and evaluated as d-notation, and the resulting value is ran as a Shadowrun roll.
- If a message is prefixed with `!wodroll`, the text following is parsed and evaluated as d-notation, and the resulting value is ran as a World of Darkness roll.
    - The roll can be modified with the strings `8again`, `9again`, and `chance`. Rolls of < 1 dice are ran as chance rolls

## Environment Variables

To run this project, you will need to add the following environment variables to your .env file

`DISCORD_TOKEN`
`INSOMNIAC_IDS`

You can use example.env as a skeleton.

## TODO
- Figure out a good way to print chained rolls that shows intermediate results
- Add rolling types for other popular systems (Savage Worlds?)
- Curate the quote list a bit more
- Update quote engine to use sqlite db rather than embedded quotes
- Add more interesting responses to bad rolls

## Why
Constant imposter syndrome mostly.

I named it Costanza because if there was a modern day saint of having the most rotten luck, it would be George. Plus, most of
my experiences as a player in D&D involve some Costanza-like decision making.