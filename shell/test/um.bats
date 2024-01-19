setup() {
    load 'test_helper/bats-support/load'
    load 'test_helper/bats-assert/load'

    # get the containing directory of this file
    # use $BATS_TEST_FILENAME instead of ${BASH_SOURCE[0]} or $0,
    # as those will point to the bats executable's location or the preprocessed file respectively
    DIR="$( cd "$( dirname "$BATS_TEST_FILENAME" )" >/dev/null 2>&1 && pwd )"
    # make executables visible to PATH
    PATH="$DIR/..:$PATH"
    # commands depend on being in the journal/ path.
    cd $DIR/mockjournal
    # so that we can run next.sh without dealing with emacsclient.
    export UMNEXTECHO=true
}

teardown() {
    export UMNEXTECHO=false
    cd -
}

@test "um usage" {
    run um
    assert_output 'usage: um (next|last|cat|rename|tag)'
}

@test "next empty" {
    run um next
    assert_output '100.md'
}

@test "next descriptor" {
    run um next foo
    assert_output '100.foo.md'
}

@test "next descriptor tag" {
    run um next foo bar
    assert_output --partial '(insert "+ bar\n")'
}

@test "next descriptor +" {
    run um next foo +
    assert_output --partial '(insert "+ foo\n")'
}

@test "um last" {
    run um last
    assert_output '099.poo.md'
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

um_rename_awk() {
    echo "$1
$2" | awk -f ../../rename.awk
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
