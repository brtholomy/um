# Get the file name in any numbered series of text files and open with emacs

# Usage:
#     $ next [descriptor]

# Handles an optional string descriptor. Like this:
#    $ ls
#    011.blah.md
#    $ next foo
#    012.foo.md

# File format:
# in regex terms:
#   [0-9]+\.[.*\.]*md|txt
# in plain terms:
#   [leadings numbers with consistent width].[optional string descriptor and . separator][md|txt]

# NOTE the optional $1 arg passed to awk with -v:
NEXTFILE=`$UMBASEPATH/last.sh | awk -f $UMBASEPATH/next.awk -v arg=$1`

# HACK: emacs won't accept piped input as a file name, so we use --eval to open
# the file via find-file.
emacsclient --eval "(progn (find-file \"$NEXTFILE\") (um-journal-header) (message \"creating $NEXTFILE\"))"
