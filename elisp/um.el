;;; um.el --- An Ultralight database for Markdown composition. -*- lexical-binding: t -*-

;; by bth

;; Um is an ultralight database design and emacs toolkit for organizing writing
;; into larger compositions. It uses only unique filenames, POSIX filesystem
;; conventions, the builtin `project' package, and standard POSIX commandline
;; tools to build out the CLI.

;; It consists of two parts:
;; 1. The elisp for functionality within emacs.
;; 2. Bash scripts for the commandline interface.

;; This is somewhat like org-roam, except without any database dependency. (And
;; it assumes Markdown rather than org, which I don't care for.)

;; This depends on a few simple ideas:

;; 1. A numbered filename specification which serves as unique id.

;; 2. Placing that same id into the file header, so that concatenated files
;; have reference back to their source. See the `um cat` command.

;; 3. Using a source "journal/" directory, where source files should first be
;; composed and where we can assume a file exists if not elsewhere.

;; 4. Using the built-in `project' package to organize compositions built from
;; these source files.

;; features provided:
;; `um-journal-find-file' via `find-file': open a file under point in the
;; current project, falling back to a source directory.
;; `um-file-at-point-advice' via `embark-dwim': open file under point in all
;; known projects, falling back to a source directory.

;; `um-grep-tag': search files with same tag

(require 'project)

(defvar um-journal-path-glob ".*/writing/journal"
  "Primary path glob for the journal. This allows various mountpoints.")

(defvar um-date-format "%Y-%m-%d"
  "Format passed to `format-time-string' when creating `um-journal-header'. Defaults to ISO8601.")

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; journal header

;;;###autoload
(defun um-journal-header ()
  "Create a header composed of:

# filename
: date

"
  (interactive)
  (insert
   (format "# %s\n: %s\n\n" (buffer-name) (format-time-string um-date-format))
   ))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; find-file-at-point

;; the journal/ serves as content origin, so it must be treated specially.
(defun um-journal-path()
  (or (car (seq-filter
            (lambda (path)
              ;; this way I don't have to hardcode the full path:
              (string-match um-journal-path-glob path))
            ;; stores all project roots in a list of strings:
            (project-known-project-roots)))
      ;; fallback if not found
      default-directory
      ))

;;;;;;;;;;;;;;;;;;;;;
;; find-file C-x C-f solution
;; combined with the hook into `file-name-at-point-functions', this means I can
;; run find-file at point, followed by M-n :

;; TODO: reuse same fallback logic as the embark advice if I keep this.
;;;###autoload
(defun um-journal-find-file ()
  (let ((default-directory (um-journal-path))
        (thing (thing-at-point 'filename))
        )
    (expand-file-name thing))
)
(add-hook 'file-name-at-point-functions 'um-journal-find-file nil t)

;;;;;;;;;;;;;;;;;;
;; Embark solution

;; Load this with an advice:
;; (advice-add 'embark-target-file-at-point :around 'um-file-at-point-advice)

;; `embark-target-file-at-point' uses `ffap-file-at-point', which uses
;; `default-directory' to expand the filename.

;; NOTE: decided to use an advice rather than add my own function to
;; `embark-target-finders', since I'd have to remove the file finder and put my
;; function in the same place in the list anyway.

;;;###autoload
(defun um-file-at-point-advice (origfunc)
  "Like find-file-at-point, but checks all known project paths.

In this order:
1. current project root
2. all subdirectories in current project
3. all other project roots
4. journal/ path

Since journal/ is the origin and other projects shadow it.
"
  (or
   ;; current project
   (funcall origfunc)

   ;; NOTE: Recurse down into current project directories.
   ;; project-find-file does it fine, just don't want interactive
   ;; completing-read call.
   (when (project-current)
     (let ((result) (tap (thing-at-point 'filename t)))
       ;; TODO: this is procedural thinking and not lisp-like:
       (cl-loop for file in
                (project-files (project-current))
                do (let ((basename (file-name-nondirectory file)))
                     (setq result (equal tap basename)))
                if result
                ;; see `embark-target-finders':
                return (cons 'file file)
                )))

   ;; all other projects
   (let ((result))
     (cl-loop for project-path in
              (seq-remove
               (lambda (path)
                 (or (string-match (um-journal-path) path)
                     (when (project-current)
                       (string-match (caddr (project-current)) path))
                     ))
               (project-known-project-roots))
              do (let ((default-directory project-path))
                   (setq result (funcall origfunc)))
              if result
              return result
              ))

   ;; journal path
   (let ((default-directory (um-journal-path)))
     (funcall origfunc))
   ))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; tags

;; TODO: do I want to search all projects?
(defun um-grep-tag ()
  "Run project-find-regexp on the current identifer as a journal frontmatter tag.

NOTE: this searches in the current project root only.
"
  (interactive)
  (project-find-regexp (concat "^\\+ " (thing-at-point 'word) "$"))
  )

;; NOTE: to export files with a given tag:
;; grep -l '\+ TAG' *md | xargs cp -t /tmp/
;; note the -l flag to grep to output files only, and the -t flag to cp to
;; specify a target dir (since it appends the list from xargs I think).
;; Or, in dired, run `find-grep-dired'.

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; shell integration

;;;###autoload
(defun um-next (next-file &optional tags skip-header)
  "Takes the next-file generated by the shell script and opens it.

Optional TAGS string may contain more than one tag separated by a comma.
The SKIP-HEADER option says not to insert the um-journal-header at all,
in which case TAGS is also ignored.
"
  (find-file next-file)
  (unless skip-header
    (um-journal-header)
    (if tags
        (progn (forward-line -1)
               (dolist (tag (split-string tags ","))
                 (insert (format "+ %s\n"
                                 (if (equal "+" tag)
                                     (cadr (split-string next-file "\\."))
                                   tag))))
               (forward-line))))
  (message (format "creating %s" next-file))
  )

(defun um-next-shell (dir &optional descriptor skip-header)
  "Calls um next in the provided dir.

With the option not to leave behind the um-journal-header.

Returns the value of next-file computed by um next
"
  (let ((default-directory dir) (next-file))
    ;; https://emacs.stackexchange.com/a/19878
    (setq next-file (shell-command-to-string
                     (concat "export UMNEXTPRINT=true; um next"
                             (concat " " descriptor))))
    (um-next next-file nil skip-header)
    next-file
    )
  )

(provide 'um)
