# cat for Markdown in my ultralight writing database
# wherein concatenated filed need a "\n---\n\n" between them.

#############################################################################
# cat with file headers | replace with section string | remove first instance

# NOTE: only need to add one newline, because the header inserted by tail
# implicitly adds two : \n==> foo <==\n
tail -n +1 "$@" | sed 's/^==> .* <==$/---\n/g' | tail -n +3


##############################
# failed approaches

# find . -type f -name "$@" -exec cat {} \; -printf "\n---\n\n"

# this leaves one trailing, could cat the whole thing and then cut the tail off
# MDCATALL=`for i in "$@";
# do
#     cat $i
#     echo '\n---\n\n'
# done`
# echo $MDCATALL | head -n +2
