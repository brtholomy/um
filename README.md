# um

An (U)ltralight database for (M)arkdown composition.

`um` is an ultralight *zettelkasten* design and emacs toolkit for organizing writing into larger compositions. It uses unique filenames, simple plaintext tags, the builtin emacs `project` package, and a Go CLI.

It tries to be stupid-simple on the filesystem side, while offering powerful conveniences on the tooling side. The idea is to prioritize the moment of creation and get all noise out of the way.

It consists of two parts:

1. Elisp for functionality within emacs. Much of it is integrated directly with the `project` interface, and consists of the ability to find a file from a filename and search for tags.

2. A commandline interface written in Go. I prefer a CLI for file creation and management rather than more emacs functions, because chaining shell commands is easy and I think of the commandline as the point of reference for all filesystem management.

Conceptually, `um` is somewhat like org-roam, except without any database dependency. And it assumes Markdown rather than org, which I find too heavy-handed.

This "database" depends on a few simple ideas:

1. A sequentially numbered filename specification which serves as unique id, like this: `001.foo.md`.  The filesystem is the database. Note that the string descriptor is optional. In regex terms:

    ```sh
    ls | egrep '^[[:digit:]]+\.?.*\.md|txt'
    ```

    Or:

    ```
    digits.[descriptor.]md
    ```

2. A very simple file header consisting of the title and date, with optional tags and place marker.

    ```markdown
    # 001.foo.md
    : 2024.01.14
    - place_optional
    + tag_optional
    ```

3. Using a "root" project defined by `um-root-glob`, where source files should first be composed and where we can assume a file exists if not elsewhere. This matters when trying to navigate back to a source file.

4. Using the built-in emacs `project` package and a powerful CLI to organize compositions built from these source files.

# installation

Clone it:

```sh
git clone https://github.com/brtholomy/um.git ~/.emacs.d/um
```

Build the `um` binary:

```sh
cd ~/.emacs.d/um/go
go build -o um .
```

Symlink somewhere in your `PATH`:

```sh
ln -s ~/.emacs.d/um/go/um /usr/local/bin/um
```

## config

My config looks like this:

```elisp
(use-package um
  :after (project embark)

  :custom (um-root-glob ".*/writing/journal")

  :bind
  ("M-s t" . um-tag-grep)
  ("M-r t" . um-tag-dwim)

  :config
  (advice-add 'embark-target-file-at-point :around 'um-target-file-at-point-advice)

  :load-path "um/elisp"
)
```

Note the optional setup of `embark`, which allows us to find files intelligently, explained below.

# ultralight "database"

What I consider the killer feature of `um` is the ability to jump to files from filenames listed in any "child" project back to the "source" project, defined via `um-root-glob`. This means that child projects can be initially defined as simple lists of filenames, which I do to compose larger pieces. A simple filename anywhere can also serve as a "link".

There are two primary ways this is accessed:

1. `find-file` and `next-history-element`, which will work as `C-x C-f M-n` out of the box. See `um-find-file-at-point`.
2. By invoking `emark-dwim` on any filename. Requires `embark` and the `advice-add` shown above.

# CLI

The "command line interface" is a set of conveniences: since this is all plain text, we could just as easily create everything manually.

Try:

```
um help
um tag --help
```

## seed

To get started, create an empty directory to serve as content origin. It doesn't matter where or what it's called, since the CLI only assumes a sequentially numbered collection of files. Then create your first file, while seeding the zero-width. 4 zeros is plenty, since that means 10k files. My zettelkasten is 20 years old and has about 3000 entries with almost a million words:

```
touch 0000.md
```

## um next

To create a new file:

```sh
um next
```

This will create the file, open it with emacsclient, and run `um-journal-header` in that file:

```markdown
# 01.md
: 2024.01.14
```

If you save this file and run `um next` again you get:

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

Or send a list of tags separated by commas. `+` still works:

```sh
um next foo +,bar,baz
```

Yields:

```markdown
# 05.foo.md
: 2024.01.14
+ foo
+ bar
+ baz
```

## um last

`um last` will print the name of the last numbered file.

## um mv

The `um mv` command makes it easier to add or change the string descriptor, while also updating the header:

```sh
um mv 02.foo.md bar
```

And we get `02.bar.md`:

```markdown
# 02.bar.md
: 2024.01.14
```

## um tag

The file header allows for an optional list of tags, one per line, marked by a leading `+`:

```markdown
# 02.md
: 2024.01.14
+ foo
+ bar
```

```markdown
# 03.md
: 2024.01.14
+ foo
```

This is the most powerful aspect of `um`: a simple list of tags applied to source files. It encourages small files organized from the bottom up, rather than topdown management - which gets in the way of good creative moods.

We can then search for files containing these tags. This list just goes to stdout, so the idea is to pipe it into a file for reordering within emacs:

```sh
um tag foo > ./some/filelist.md
```

The query supports union as `+` and intersection as `,`:

```sh
> um tag foo+bar

02.md

> um tag foo,bar

02.md
03.md
```

And the complement:

```sh
> um tag foo --invert

03.md
```

Pipe them together to use a big union from which to subtract:

```sh
um tag foo,bar | um tag baz --invert
```

Run `um tag --help` to see what it can do.

## um sort

When working with the filelists produced by `um tag`, we'll want to rearrange the order of files and add or remove tags. Then when we update our filelist by rerunning `um tag`, we want the output to respect our updated order. `um sort` does this:

```sh
um tag foo+bar | um sort --key some/filelist.md
```

## um cat

This command is designed to work with the filelists produced by `um tag`. It separates files with a Markdown horizontal rule `---` while stripping their headers:

```sh
um tag foo | um cat
```

As the last step in composing larger pieces, it accepts a filelist and a base directory to find those files. We can then pipe the output wherever we like:

```sh
um cat filelist.md --base ../ > finished.md
```

And there you have the virtue of the Unix philosophy.
