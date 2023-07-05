package db

const TextQuoteType = "quote"
const FileQuoteType = "file"

type Quote struct {
	Id   uint
	Data string
	Type string
}
