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

;; 3. Using a source `um-root-glob' directory, where source files should first be
;; composed and where we can assume a file exists if not elsewhere.

;; 4. Using the built-in `project' package to organize compositions built from
;; these source files.

;; features provided:
;; `um-find-file-at-point' via `find-file': open a file under point in the
;; current project, falling back to a source directory.
;; `um-target-file-at-point-advice' via `embark-dwim': open file under point in all
;; known projects, falling back to a source directory.

;; `um-grep-tag': search files with same tag

(require 'project)

(defcustom um-root-glob ".*/writing/journal"
  "Primary glob for `um-root-path.' This allows various mountpoints."
  :type '(string)
  )

(defcustom um-date-format "%Y-%m-%d"
  "Format passed to `format-time-string' when creating `um-journal-header'. Defaults to ISO8601."
  :type '(string)
  )

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
;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;

(defun um-root-path ()
  (or (car (seq-filter
            (lambda (path)
              ;; this way I don't have to hardcode the full path:
              (string-match um-root-glob path))
            ;; stores all project roots in a list of strings:
            (project-known-project-roots)))
      ;; fallback if not found
      default-directory
      ))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; `find-file' integration
;;
;; combined with the hook into `file-name-at-point-functions', this means we can
;; run `find-file' at point, followed by `next-history-element'. By default:
;; C-x C-f M-n

;;;###autoload
(defun um-find-file-at-point ()
  "Return full file path of thing-at-point, falling back to:

1. `default-directory'
2. `project-root'
3. `um-root-path'
"
  (let ((dirs (list
               default-directory
               (when (project-current) (project-root (project-current)))
               (um-root-path)))
        dir
        file)
    (while (and dirs (not file))
      (setq dir (car dirs)
            dirs (cdr dirs))
      (let* ((default-directory dir)
             (fap (thing-at-point 'existing-filename t)))
        (when fap
          (setq file (expand-file-name fap)))))
    file))

;; NOTE: we now load this by default if the package is loaded.
(add-hook 'file-name-at-point-functions 'um-find-file-at-point)

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; `embark-dwim' integration
;;
;; Load this with an advice:
;; (advice-add 'embark-target-file-at-point :around 'um-target-file-at-point-advice)

;; `embark-target-file-at-point' uses `ffap-file-at-point', which uses
;; `default-directory' to expand the filename.

;; NOTE: decided to use an advice rather than add my own function to
;; `embark-target-finders', since I'd have to remove the file finder and put my
;; function in the same place in the list anyway.

;;;###autoload
(defun um-target-file-at-point-advice (origfunc)
  "Wrapper for `embark-target-file-at-point' that falls back to:

1. `default-directory'
2. `project-root'
3. `um-root-path'
"
  (or
   ;; this does both default-directory and project-root:
   (funcall origfunc)
   (let ((default-directory (um-root-path)))
     (funcall origfunc))
   ))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; tags

;; NOTE: this will get saved by savehist-mode
(defvar um-tags-history nil "History of inserted or searched for tags. Populates
`completing-read'.")

(defconst um-tag-regexp "^\\+ \\(.*\\)$")

(defun um-header-current-buffer ()
  (car (split-string (buffer-substring-no-properties (point-min) (point-max))
                     "\n\n" t
                     )))

(defun um-tag-first-in-current-buffer ()
  (let ((header (um-header-current-buffer)))
    (string-match um-tag-regexp header)
    (match-string 1 header)
    ))

(defun um-grep-tag ()
  "Run `project-find-regexp' on a selection made from `um-tags-history' via `completing-read'.

The initial value provided to `completing-read' is the first tag found in the
current buffer: it will be first in the list and available via \\`M-n'.

NOTE: searches in the current project root by default, but
\\[universal-argument] will allow choice of the base directory as in `project-find-regexp'.

Ultimately this relies on `xref-matches-in-files', which calls `xref-search-program'.
"
  (interactive)
  (project-find-regexp (concat "^\\+ "
                               (completing-read "um tag: " um-tags-history nil
                                                nil nil
                                                ;; NOTE: this means searches
                                                ;; will add to the history:
                                                'um-tags-history
                                                (um-tag-first-in-current-buffer)
                                                )
                               "$")))

(defun um-tag-insert (tag)
  "Insert TAG as + tag\n in current buffer appending to the journal header."
  (goto-char (point-min))
  (search-forward "\n\n")
  (previous-line)
  (insert (concat "+ " tag "\n"))
  )

(defun um-tag-delete (tag)
  "Delete TAG from the journal header in current buffer."
  (goto-char (point-min))
  (search-forward (concat "+ " tag "\n"))
  (previous-line)
  (kill-line)
  )

(defun um-tag-do (tag insert)
  "Insert or delete TAG from the journal header in current buffer.

insert when INSERT > 0, delete otherwise."
  (if (> insert 0)
      (um-tag-insert tag)
    (um-tag-delete tag)))

;;;###autoload
(defun um-tag-insert-dwim (ARG)
  "Run `um-tag-do' on a list of filenames if region active outside
  dired-mode, or if marks exist in dired-mode, or the filename at point, and
  finally in the current buffer if none of those conditions match.

Assumes the files of interest are returned by `um-root-path'.

Negative prefix arg is handled by `um-tag-do', which see.
"
  (interactive "p")
  (let* (
         (tag (completing-read "insert um tag: " um-tags-history nil nil nil 'um-tags-history))
         (marks (if (eq major-mode 'dired-mode) (dired-get-marked-files) nil))
         ;; NOTE: 'existing-filename would be better to avoid bogus strings, but
         ;; in view-mode when in another project, we can't verify it exists yet:
         (fap (thing-at-point 'filename))
         (files (cond
                 ((and (region-active-p) (not (eq major-mode 'dired-mode)))
                  (string-split (buffer-substring (region-beginning) (region-end))))
                 (marks marks)
                 (fap (list fap))
                 ))
         ;; save-excursion is not working below, why?
         (buf (current-buffer)))

    (if files (progn (dolist (f files)
                       ;; TODO: this should use the fallback logic:
                       (find-file (expand-file-name f (um-root-path)))
                       (um-tag-do tag ARG)
                       )
                     (switch-to-buffer buf)
                     (save-some-buffers t)
                     )
      (um-tag-do tag ARG))
    (let ((history-delete-duplicates t))
      (add-to-history 'um-tags-history tag))))

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
