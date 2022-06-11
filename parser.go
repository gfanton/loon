package main

type Parser[output any] interface {
	Parse(line string) output
}
