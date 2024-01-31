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
;; `um-journal-header' defines the standard header
;; `um-project-find-meta' via `project-switch-project': open a file like
;; *meta.md in a given project.

;; (Optional) `um-file-at-point-advice' via `embark-dwim': open file in known projects
;; with fallback to journal/

;; `um-grep-tag': search files with same tag

(require 'project)

;; Primary path glob for the journal. This allows various mountpoints.

(setq um-journal-path-glob ".*/writing/journal")

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; journal header

(defun um-iso-8601 ()
  "Insert a date stamp in ISO 8601 order, but with periods in place of dashes."
  (format-time-string "%Y.%m.%d")
  )

;;;###autoload
(defun um-journal-header ()
  "Create a header composed of:

# filename
: date

"
   (interactive)
   (insert
    (format "# %s\n: %s\n\n" (buffer-name) (um-iso-8601))
    )
   )

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; journal/.project.el

;; normally .git detection works fine. But for my journal/ I don't git.
;; https://michael.stapelberg.ch/posts/2021-04-02-emacs-project-override/
(defun um-project-override (dir)
    "Returns the parent directory containing a .project.el file,
 if any, to override the standard project.el detection logic when
 needed.
"
    (let ((override (locate-dominating-file dir ".project.el")))
      (if override
          (list 'vc nil override)
        nil)))
;; Cannot use :hook because 'project-find-functions does not end in -hook
(add-hook 'project-find-functions #'um-project-override)

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; find-file-at-point

;; the journal/ serves as content origin, so it must be treated specially.
(setq um-journal-path (car (seq-filter
                         (lambda (path)
                           ;; this way I don't have to hardcode the full path:
                           (string-match um-journal-path-glob path))
                         ;; stores all project roots in a list of strings:
                         (project-known-project-roots))))

;;;;;;;;;;;;;;;;;;;;;
;; find-file C-x C-f solution
;; combined with the hook into `file-name-at-point-functions', this means I can
;; run find-file at point, followed by M-n :

;; TODO: reuse same fallback logic as the embark advice if I keep this.
;;;###autoload
(defun um-journal-find-file ()
  (let ((default-directory um-journal-path)
        (thing (thing-at-point 'filename))
        )
    (expand-file-name thing))
)
(add-hook 'file-name-at-point-functions 'um-journal-find-file nil t)

;;;;;;;;;;;;;;;;;;
;; Embark solution (optional)

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
2. all other project roots
3. journal/ path

Since journal/ is the origin and other projects shadow it.

TODO: Use a special directory name for potentials/ or the like, in case it's not
in the root dir.
"
  (or
   ;; current project
   (funcall origfunc)

   ;; all other projects
   (let ((result))
     (cl-loop for project-path in
              (seq-remove
               (lambda (path)
                 (or (string-match um-journal-path path)
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
   (let ((default-directory um-journal-path))
     (funcall origfunc))
   ))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; C-x p p

;;;###autoload
(defun um-project-find-meta ()
  (interactive)
  ;; this works because `project-switch-project' overrides the dir:
  (find-file (expand-file-name "*meta.md" (caddr (project-current))) t)
  )

;; overwrite
(setq project-switch-commands
      '(
        (um-project-find-meta "open meta" "m")
        (project-dired "dired")
        (project-find-file "find file")
        (project-find-regexp "find regexp")
        ))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; tags

;; TODO: do I want to search all projects?
(defun um-grep-tag ()
  "Run project-find-regexp on the current identifer as a journal frontmatter tag.

NOTE: this searches in the current project root only.
"
  (interactive)
  (project-find-regexp (concat "\\+\\ " (thing-at-point 'word)))
  )

;; NOTE: to export files with a given tag:
;; grep -l '\+ TAG' *md | xargs cp -t /tmp/
;; note the -l flag to grep to output files only, and the -t flag to cp to
;; specify a target dir (since it appends the list from xargs I think).
;; Or, in dired, run `find-grep-dired'.


(provide 'um)
