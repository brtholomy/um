# um

An (U)ltralight database for (M)arkdown composition.

`um` is an ultralight database design and emacs toolkit for organizing writing into larger compositions. It uses only unique filenames, POSIX filesystem conventions, the builtin emacs `project` package, and standard POSIX commandline tools to build out the CLI.

It tries to be stupid-simple. Because the "CLI" is just sh and awk, it really only has emacs as a dependency.

It consists of two parts:

1. Elisp for functionality within emacs. Most of it is integrated directly with the `project` interface, and consists the ability to find a file from a filename.

2. Shell scripts for the commandline interface. I prefer a CLI for file creation and management rather than more emacs functions, because chaining shell commands is easy and I think of the commandline as the point of reference for all filesystem management.

Conceptually, `um` is somewhat like org-roam, except without any database dependency. (And it assumes Markdown rather than org, which I don't care for.)

This "database" depends on a few simple ideas:

1. A sequentially numbered filename specification which serves as unique id, like this: `001.foo.md`.  The filesystem is the database. Note that the string descriptor is optional. In regex terms:

    ```sh
    ls | egrep '^[[:digit:]]+\.?.*\.md|txt'
    ```

    In pseudocode:

    ```
    [leadings numbers].[optional string descriptor and . separator][md|txt]
    ```

2. Placing that same id into the file header, so that concatenated files have reference back to their source - like this:

    ```markdown
    # 001.foo.md
    : 2024.01.14
    - place_optional
    + tag_optional
    ```

    See the `um cat` command for why this matters.

3. Using a source "journal/" directory, where source files should first be composed and where we can assume a file exists if not elsewhere. This matters when trying to navigate back to a source file.

4. Using the built-in emacs `project` package to organize compositions built from these source files.

# installation

Clone it:

```sh
git clone https://github.com/brtholomy/um.git ~/um
```

Symlink the `um` CLI somewhere in your `PATH`:

```sh
ln -s ~/um/shell/um /usr/local/bin/um
```

## emacs config

My config looks like this:

```elisp
(use-package um
  :after (project embark)

  :config
  (defalias 'journal 'um-journal-header)

  (advice-add 'embark-target-file-at-point :around 'um-file-at-point-advice)

  :load-path "~/um/elisp"
)
```

Note the optional setup of `embark`, which allows us to visit files just by invoking `emark-dwim` on a file header. This will also work with `C-x C-f M-n` out of the box.

# CLI usage

The "command line interface" is just a small set of convenient scripts: since these are just text files, we could just as easily create everything manually.

1. To get started, create an empty directory to serve as content origin. It doesn't matter where or what it's called, since the shell scripts only assume a sequentially numbered collection of files.

2. Decide on a zero-width for your series and seed the first file:

  ```sh
  touch 01.md
  ```

  The rest of the commands will now work with this width.

## next

Create a new file, open it with emacsclient, and run `um-journal-header`:

```sh
um next
```

Yields:

```markdown
# 02.md
: 2024.01.14
```

Create a new file with an optional descriptor:

```sh
um next foo
```

Yields:

```markdown
# 03.foo.md
: 2024.01.14
```

Create a new file with an optional descriptor and tag:

```sh
um next foo bar
```

Yields:

```markdown
# 04.foo.md
: 2024.01.14
+ bar
```

Or just append `+` to add the descriptor as a tag:

```sh
um next foo +
```

Yields:

```markdown
# 05.foo.md
: 2024.01.14
+ foo
```

## last

`um last` will print the name of the last numbered file. This can be useful on its own, or like this:

```sh
um last | xargs emacs
```

## cat

One of the advantages of the header specification, is that it allows us to `cat` files together without losing track of their origin. This is probably the most important design decision, since the more complicated option would have been to keep all source files intact, and compose larger projects by reference back to the original file only - like org-roam does. This would, however, introduce considerably more complexity, and it would mean that project files become obscure lists of files, instead of content - which would largely defeat the value of git and plaintext in general.

So rather than overengineer, like most engineers do, I choose to stick with plaintext files everywhere, and let the source file serve as history relative to the larger project. It means that the source of truth for my writing projects travels downstream, which is not ideal but largely fine - since git makes history of everything anyway.

All this stupid little command does, is concatenate files together while placing a Markdown horizontal rule between them, like this:

```sh
um cat 01.md 02.md

# 01.md
: 2024.01.14

---

# 02.md
: 2024.01.14
```

Which means we can pipe the output wherever we like:

```sh
um cat 01.md 02.md > /tmp/foo.md
```

## rename

The `um rename` command makes it easier to add or change the string descriptor:

```sh
um rename 02.md foo
```

And we get `02.foo.md`.

## tag

The file header allows for an optional block of tags, marked by a `+` like this:

```markdown
# 02.md
: 2024.01.14
+ foo
+ bar
```

We can then grep for these tags. This command just makes it easier to list them:

```sh
um tag foo
02.md
```

So we can pipe it:

```sh
um tag foo | xargs um cat > /tmp/foo.md
```

And there you have the virtue of the Unix philosophy.
