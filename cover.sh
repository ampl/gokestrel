#!/bin/bash
go test -v -coverprofile cover.out ./cmd/gokestrel
go tool cover -html=cover.out -o cover.html
