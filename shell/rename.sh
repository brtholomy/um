# rename : renames the string descriptor used in numbered series.

# Usage:

# > rename 100.md bar
# or
# > rename 100.foo.md bar

# is equivalent to:
# > mv 100.md 100.bar.md
# > mv 100.foo.md 100.bar.md

# split the args by a newline so that awk reads them as separate "records"
mvcmd=`echo "$1
$2" | awk -f $UMBASEPATH/rename.awk`

mv $mvcmd
