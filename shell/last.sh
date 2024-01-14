# Get the last file name of any numbered series of text files.

# file format:
# in regex terms:
#   [0-9]+\.[.*\.]*md|txt
# in plain terms:
#   [leading numbers with consistent width].[optional string descriptor and . separator][md|txt]

# TODO: signal no file found, so that next knows to create one.
# Grep exits 1, but tail exits 0.
ls | egrep '^[[:digit:]].*\.md|txt' | tail -n 1
