#!/bin/bash

# Generate the next filename in any numbered series of text files and open with emacs.

# Usage:
#     $ next [descriptor] [tag]

# Handles optional string descriptor and tag. Like this:
#    $ ls
#    011.blah.md
#
#    $ next foo
#    012.foo.md
#
#    $ next foo bar
#    013.foo.md
#
# Which populates the file like this:
#
# # 13.foo.md
# : 2024.01.14
# + bar
#
# See https://github.com/brtholomy/um#next

# NOTE: handle the exit 1 case of last.sh
LASTFILE=`$UMBASEPATH/last.sh || echo $UMDEFAULTINIT`

# NOTE: the optional $1 arg passed to awk with -v:
NEXTFILE=`echo $LASTFILE | awk -f $UMBASEPATH/next.awk -v arg=$1`

# tag is optional second arg to this script
TAGINSERT='(previous-line) (insert "+ %s\\n") (next-line)'
if [ $2 ]; then
    # + means insert the string descriptor as tag
    if [ $2 = '+' ]; then
        UMTAG=$(printf "$TAGINSERT" "$1")
    else
        UMTAG=$(printf "$TAGINSERT" "$2")
    fi
fi

UMELISP="(progn (setq um-next-file \"$NEXTFILE\") (find-file um-next-file) (um-journal-header) $UMTAG (message \"creating $NEXTFILE\"))"

# for testing and piping
if [ $UMNEXTECHO ]; then
    echo $UMELISP
else
    # HACK: emacs won't accept piped input as a file name, so we use --eval to open
    # the file via find-file.
    emacsclient --eval "$UMELISP"
fi

exit 0
