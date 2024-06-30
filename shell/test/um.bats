setup() {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'

    # get the containing directory of this file
    # use $BATS_TEST_FILENAME instead of ${BASH_SOURCE[0]} or $0,
    # as those will point to the bats executable's location or the preprocessed file respectively
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" >/dev/null 2>&1 && pwd )"
    # make executables visible to PATH
    # FIXME: why isn't this working?
    # PATH="$DIR/../:$PATH"
    # commands depend on being in the journal/ path.
    export ORIGPWD=`pwd`
    cd $DIR/mockjournal
    # so that we can run next.sh without dealing with emacsclient.
    export UMNEXTECHO=true
}

teardown() {
    export UMNEXTECHO=false
    cd $ORIGPWD
    git restore $DIR/mockjournal
}

@test "um usage" {
    run um
    assert_output 'usage: um (next|last|cat|rename|tag)'
}

um_next_awk() {
    echo "$1" | awk -f $DIR/../next.awk -v arg="$2"
}

@test "next.awk" {
    run um_next_awk 01.md
    assert_output '02.md'

    run um_next_awk 01.md foo
    assert_output '02.foo.md'
}

@test "um next empty" {
    run um next
    assert_output --partial '100.md'
}

@test "um next empty elisp" {
    run um next
    assert_output '(progn (find-file "100.md") (um-journal-header) (message "creating 100.md"))'
}

@test "um next descriptor" {
    run um next foo
    assert_output --partial '100.foo.md'
}

@test "um next descriptor tag" {
    run um next foo bar
    assert_output --partial '(insert "+ bar\n")'
}

@test "um next descriptor +" {
    run um next foo +
    assert_output --partial '(insert "+ foo\n")'
}

@test "um last" {
    run um last
    assert_output '099.poo.md'
}

@test "um last empty" {
    cd foo
    run um last
    assert_failure
    assert_output 'no files found'
}

@test "um cat" {
    run um cat 001.md 002.blah.md
    assert_output '# 001
: 2024.01.14
+ foo

---

# 002.blah
: 2024.01.14
+ foo'
}

@test "um rename" {
    run um rename 001.md blah
    run ls 001*
    assert_output '001.blah.md'
    mv 001.blah.md 001.md
}

# this doesn't work because bats doesn't handle stdin the way a tty does.
# @test "um rename piped" {
#     run um last | um rename gloo
#     run ls 099*
#     assert_output '099.gloo.md'
#     mv 099.gloo.md 099.poo.md
# }

um_rename_awk() {
    echo "$1
$2" | awk -f $DIR/../rename.awk
}

@test "rename.awk" {
    run um_rename_awk 01.md foo
    assert_output '01.md 01.foo.md'

    run um_rename_awk 01.foo.md bar
    assert_output '01.foo.md 01.bar.md'
}

@test "um tag" {
    run um tag foo
    assert_output '001.md
002.blah.md'
}
