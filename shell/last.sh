#!/bin/bash

# Get the last file name of any numbered series of text files and open it with emacs.

# file format:
# in regex terms:
#   [0-9]+\.[.*\.]*md|txt
# in plain terms:
#   [leading numbers with consistent width].[optional string descriptor and . separator][md|txt]

UMALLFILES=`ls | grep -E '^[[:digit:]].*\.(md|txt)'`
UMLASTSTATUS=$?
UMLASTFILE=''

if [ "$UMLASTSTATUS" -ne 0 ]; then
    UMLASTFILE="$UMDEFAULTINIT"
else
    # NOTE: if var isn't quoted echo will change newlines to spaces
    UMLASTFILE=`echo "$UMALLFILES" | tail -n 1`
fi

if [ "$UMDISABLEEMACS" = true ] || [ "$UMTEST" = true ] ; then
    echo $UMLASTFILE
    exit 0
fi

echo "emacsclient -n $UMLASTFILE"
emacsclient -n $UMLASTFILE
