# Get the file name in any numbered series of text files and open with emacs

# Usage:
#     $ . next [descriptor]

# Handles an optional string descriptor. Like this:
#    $ ls
#    011.blah.md
#    $ . next foo
#    012.foo.md

# File format:
# in regex terms:
#   [0-9]+\.[.*\.]*md|txt
# in plain terms:
#   [leadings numbers with consistent width].[optional string descriptor and . separator][md|txt]

# what this does:
# ls | grep [regex for number.*.md] | get last | awk script to increment the digits, preserve the zero prefix, add string descriptor if supplied, and add back extension

# Note the egrep '^\d.*\.md|txt' (matches whole string) to account for an optional string between the initial number and the file extension. As in '435.blah.md'.

# old version, relied on python, which didn't handle leading zeros properly, but treated them as binary:
# ls | grep '\d.*\.md' | tail -n 1 | sed s/'\..*md'// | sed s/.*/'print & + 1'/ | python - | sed s/.*/'&.md/'

# BUGS

# 1. Emacs won't accept piped input as a file name. I'd rather keep this as a separate utility, which only printed the right filename, and then use an alias to pipe the filename to emacs, or invoke it manually.

# A.
# What I want, is this syntax:
# next [optional arg]
# To both get the file and open emacs with it.

# B.
# But this doesn't work:
# next | emacs
# Because it sees the stdin the pipe wants to feed it as a substitute for the tty in the new session, rather than as input for startup only, i.e. the filename.

# C.
# Wrapping the whole shell command here works, like this:
# emacs `ls | grep ...`

# But emacs doesn't like being run inside a child level shell, because it can't load .emacs properly. So you must run this file within the current shell using ".", like this:
# . next
# . next foo

# D.
# This does not work, because .profile must be executed, meaning that the "" expression is peeked by the shell:
# alias next="emacs `next.sh $1`"

# E.
# This does not work either, because the arg never gets filled out when evoked:
# alias next='emacs `next.sh $1`'

# F.
# This works and is the current solution:
# alias next='. next.sh'

# Which relies on 2 somewhat obscured facts:
# 1. This script calls emacs itself, wrapping the real work in ``
# 2. The alias is a simple substitution, which fortunately seems to happen before args are processed by the shell, meaning that the shell sees: ". next.sh arg"
# 3. NOTE: the symbolic link in the PATH should not be simply "next", because it
# may in certain situations conflict with the alias. The link should be "next.sh" -> ~/config/..
# 4. NOTE: I use 'e' rather than 'emacs', to use my clever emacs/emacsclient function.


NEXTFILE=`$UMBASEPATH/bash/last/last.sh | awk '

BEGIN {
# set field separator, blank by default
FS = "."
}

{
newnum = $1 + 1
nzeros = length($1) - length(int(newnum))

# could I do this with a simpler regex to select all leading zeros?
leadingzeros = ""
if (nzeros > 0)
  for (i = 1; i <= nzeros; i++)
    leadingzeros = sprintf("%s%s", leadingzeros, "0")

if (nzeros > 0)
  newnum = sprintf("%s%s", leadingzeros, newnum)

# start by prefixing the number, now with leading zeros
finalstr = newnum

# allow an arg input to fill out descriptor string.
# single quote break to allow the shell to fill it out, double quotes so awk will treat it as a string
arg = "'$1'"

if (arg != "")
    finalstr = sprintf("%s.%s", finalstr, arg)

# append last field by default
finalstr = sprintf("%s.%s", finalstr, $NF)

# previous loop to reassemble all fields. But I do not want the old middle string here
# Keeping it here for reference.
# for (i = 2; i <= NF; i++)
#   finalstr = sprintf("%s.%s", finalstr, $i)
# }

}

END {
print finalstr
}'
`

emacsclient --eval "(progn (find-file \"$NEXTFILE\") (uldb-journal-header) (message \"creating $NEXTFILE\"))"
