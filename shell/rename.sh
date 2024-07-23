#!/bin/bash

# rename : renames the string descriptor used in numbered series.

# Usage:

# > um rename 100.md foo
# or
# > um rename 100.foo.md bar
# or
# > um last | um rename baz

# is equivalent to:
# > mv 100.md 100.foo.md
# > mv 100.foo.md 100.bar.md
# > mv 100.bar.md 100.baz.md

# accept piped input
# https://unix.stackexchange.com/a/537581
if [[ ! -t 0 ]]; then
    FROM=$(cat -)
    TO=$1
else
    FROM=$1
    TO=$2
fi

# split the args by a newline so that awk reads them as separate "records"
mvcmd=`echo "$FROM
$TO" | awk -f $UMBASEPATH/rename.awk`

echo mv $mvcmd

mv $mvcmd
