#!/bin/sh
env=$1
## --profile is used when you have multiple aws account
cdk deploy --all --outputs-file "$1-cdk.json" --profile sideproject && \
  go run generate-full-env.go -env "$1" && \
  node ${PWD}/tools/sodium/index.js "$1"