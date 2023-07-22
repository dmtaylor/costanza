package model

const TextQuoteType = "quote"
const FileQuoteType = "file"

type InvalidQuoteTypeError string

func (i InvalidQuoteTypeError) Error() string {
	return "invalid quote type " + string(i)
}

type Quote struct {
	Id   int
	Data string
	Type string
}
