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
$2" |
awk '
BEGIN {
  # Builtin variable, "field seperator". The neatness of this is the main reason
  # I use awk to parse the filename.
  FS="."
  oldname = ""
  num = ""
  ext = ""
  newname = ""
}

NR == 1 {
  oldname = $0
  num = $1
  # important that this is the last field, not numbered, to allow for a file
  # without a string descriptor.
  ext = $NF
}

NR == 2 {
  newname = sprintf("%s.%s.%s", num, $0, ext)
}

END {
  print oldname, newname
}
'`
mv $mvcmd
