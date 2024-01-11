# why can't i use bash if statement, before appending an empty arg string?

last=`ls | egrep '^\d.*\.md|txt' | tail -n 1`

if ($* != "")
    last=last + $*;

awk '
{
print $0
}
END {print NR}' last
