# generates a string of "oldname newname" from 2 records, for use with mv. Main
# problem solved here generating the newname string using the number prefix and
# filetype affix.

BEGIN {
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
