# Genni

## Introduction

Generate Go structs using Go files

## Contents

- [Install](#install)
- [Usage](#usage)

## Install

```bash
go get github.com/rms1000watt/genni
```

## Usage

```
Genni generates go structs from go files

Usage:
  genni [flags]

Examples:
  genni -i examples/people.go
  genni -i examples/people.go --log-level debug

Flags:
  -h, --help               help for genni
  -i, --in string          Proto file to read in (Required)
      --log-level string   Set log level (debug, info, warn, error, fatal) (default "info")
```
