# um

An (U)ltralight database for (M)arkdown composition.

`um` is an ultralight database design and emacs toolkit for organizing writing into larger compositions. It uses only unique filenames, POSIX filesystem conventions, the builtin emacs `project` package, and standard POSIX commandline tools to build out the CLI.

It consists of two parts:

1. Elisp for functionality within emacs. Most of it is integrated directly with the `project` interface, and consists the ability to find a file from a filename.

2. Shell scripts for the commandline interface. I prefer a CLI for file creation and management rather than more emacs functions, because chaining shell commands is easy and I think of the commandline as the point of reference for all filesystem management.

Conceptually, `um` is somewhat like org-roam, except without any database dependency. (And it assumes Markdown rather than org, which I don't care for.)

The "database" depends on a few simple ideas:

1. A numbered filename specification which serves as unique id. The files are their own representation.

2. Placing that same id into the file header, so that concatenated files have reference back to their source. See the `um cat` command.

3. Using a source "journal/" directory, where source files should first be composed and where we can assume a file exists if not elsewhere.

4. Using the built-in emacs `project` package to organize compositions built from these source files.

## installation

Clone it:

```
git clone https://github.com/brtholomy/um.git ~/um
```

Symlink the `um` CLI somewhere in your `PATH`:

```
cd ~/um
ln -s ~/um/shell/um /usr/local/bin/um
```

## emacs config

My config looks like this:

```elisp
(use-package um
  :after (project embark)

  :config
  (defalias 'journal 'uldb-journal-header)

  (advice-add 'embark-target-file-at-point :around 'uldb-file-at-point-advice)

  :load-path "~/um/elisp"
)
```
