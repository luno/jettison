#!/bin/bash

test -z "$(gofumpt -d -e . | tee /dev/stderr)"
