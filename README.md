# `lin`

## Overview

A CLI to interact with [Linear](https://linear.app) written in Go.

## Getting started

Create an [API key](https://developers.linear.app/docs/graphql/working-with-the-graphql-api#personal-api-keys) and store it in `~/.config/lin/config.toml` using the following schema:

```toml
[auth]
api_token = "verysecret123"
```

## Usage

```shell
go build -o lin
./lin
```

## Notes

This is still very much work in progress, and currently does very little.

It is inspired in part by [the existing CLI project](https://github.com/evangodon/linear-cli). However, that was written in TypeScript and is no longer maintained, and this is being written in Go, and is not yet no longer maintained.

My goal here is to improve my Go, learn more about GraphQL, and maybe make something fun to use.


