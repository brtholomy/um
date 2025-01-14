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

# NOTE: we only want the filename, not to run emacs yet.
LASTFILE=`UMDISABLEEMACS=true; $UMBASEPATH/last.sh`

# NOTE: the optional $1 arg passed to awk with -v:
NEXTFILE=`echo $LASTFILE | awk -f $UMBASEPATH/next.awk -v arg=$1`

# tag is optional second arg to this script
if [ $2 ]; then
    UMELISP=$(printf '(um-next "%s" "%s")' "$NEXTFILE" "$2")
else
    UMELISP=$(printf '(um-next "%s")' "$NEXTFILE")
fi

# for testing and piping
# NOTE: bash booleans are such a pain, because they don't exist. https://stackoverflow.com/a/2953673
if [ "$UMNEXTPRINT" = true ]; then
    # NOTE: printf instead of echo, because echo inserts a newline, and we pipe
    # this back into elisp for um-next-shell:
    printf $NEXTFILE
elif [ "$UMTEST" = true ]; then
    echo $UMELISP
else
    # NOTE: emacs won't accept piped input as a file name, so we use --eval to run
    # um-next as elisp. Which also lets us add the header and tag.
    emacsclient --eval "$UMELISP"
fi

exit 0
