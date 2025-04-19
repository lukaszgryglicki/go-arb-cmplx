#!/bin/bash
go fmt main.go && go build -o arbcomplex main.go && strip -s ./arbcomplex
