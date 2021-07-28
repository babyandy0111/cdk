#!/bin/sh
env=$1
cdk deploy --all --outputs-file "$1-cdk.json" && \
  go run generate-full-env.go -env "$1" && \
  node ${PWD}/tools/sodium/index.js "$1"