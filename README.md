# WIP: Loon

A realtime filter pager

## Install

```sh
go install github.com/gfanton/loon@latest
```

## Usage

```
USAGE
  loon [flags] <files...>

FLAGS
  -bgcolor=false                  enable background color on multiple sources
  -config /Users/asdf/.loonrc     root config project
  -fgcolor=true                   enable forground color on multiple sources
  -linesize 10000                 If non-zero, split longer lines into multiple lines
  -noansi=false                   do not parse ansi sequence
  -nocolor=false                  disable color
  -ringsize 100000                ring line capacity
```


## Commands

`arrows` -> move arround

`alt+arrows` -> move arround faster

`enter` -> go to the end of the buffer

`ctrl+e` -> go to end of the line

`ctrl+a` -> go to the beginning of the line

`ctrl+l` -> clear the buffer
