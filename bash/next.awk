# awk script to increment the digits, preserve the zero prefix, add string descriptor if supplied, and add back extension

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
if (arg != "")
    finalstr = sprintf("%s.%s", finalstr, arg)

# append last field by default
finalstr = sprintf("%s.%s", finalstr, $NF)

}

END {
print finalstr
}
