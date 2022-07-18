package main

type Parser[output any] interface {
	Parse(sid SourceID, line string) output
}
