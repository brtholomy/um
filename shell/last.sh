#!/bin/bash

# Get the last file name of any numbered series of text files.

# file format:
# in regex terms:
#   [0-9]+\.[.*\.]*md|txt
# in plain terms:
#   [leading numbers with consistent width].[optional string descriptor and . separator][md|txt]

UMALLFILES=`ls | grep -E '^[[:digit:]].*\.(md|txt)'`
LASTSTATUS=$?

# NOTE: if we don't stop and signal here, tail will exit 0
if [ "$LASTSTATUS" -ne 0 ]; then
    echo "no files found" >&2
    exit 1
fi

# NOTE: if var isn't quoted echo will change newlines to spaces
echo "$UMALLFILES" | tail -n 1
