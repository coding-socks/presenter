# Presenter

Presenter displays slide presentations. It runs a web server that presents slide files from a specific directory.

## Introduction

[Markdown](https://en.wikipedia.org/wiki/Markdown) is a very versatile tool. This library builds on that premise and generates presentations from markdown files.

## Production readiness

**This project is still in alpha phase.** In this stage the public API can change between days.

## Documents

### Inspiration

Inspiration for the implementation comes from the following places:

- https://pkg.go.dev/golang.org/x/tools/cmd/present
- https://pkg.go.dev/golang.org/x/tools/present

How is it different from x/tools/present and x/tools/cmd/present?

- There is no legacy layout. Only Markdown formatting is available which allows the library to fully build on https://github.com/yuin/goldmark.
- No "present command invocation", markdown only.
- Some changes on how metadata and author information parsed.
- Not focused on Go. There for, the Go playground is not part of the presentation tools.
- As of now, no notes (#2) and no comments (#3) support.
- More changes are coming...
