# Get the last file name of any numbered series of text files.

# file format:
# in regex terms:
#   [0-9]+\.[.*\.]*md|txt
# in plain terms:
#   [leading numbers with consistent width].[optional string descriptor and . separator][md|txt]

ls | egrep '^\d.*\.md|txt' | tail -n 1
