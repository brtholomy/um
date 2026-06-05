;;; um.el --- An Ultralight database for Markdown composition. -*- lexical-binding: t -*-

;; by bth
;; Version: 0.1

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
;; `um-tag-grep': search files with some tag
;; `um-tag-dwim': insert or delete a tag in dired and other contexts.

(require 'project)
(require 'dired)

(defgroup um nil
  "(u)ltralight (m)arkdown zettelkasten."
  :link '(url-link :tag "Website" "https://github.com/brtholomy/um")
  :group 'files
  :prefix "um-")

(defcustom um-root-glob ".*/writing/journal"
  "Primary glob for `um-root-path.' This allows various mountpoints."
  :type '(string)
  :group 'um
  )

(defcustom um-date-separator "-"
  "Seperator used in date strings, used by `um-date-format' and `um-date-re'."
  :type '(string)
  :group 'um
  )

(defcustom um-date-format (concat "%Y" um-date-separator "%m" um-date-separator "%d")
  "Format passed to `format-time-string' when creating
 `um-journal-header'. Defaults to ISO8601."
  :type '(string)
  :group 'um
  )

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; faces

;; TODO: doesn't currently respect um-date-format:
(defconst um-date-re (rx
                      line-start
                      (literal ": ")
                      (group (repeat 4 digit)
                             (literal um-date-separator)
                             (repeat 1 2 digit)
                             (literal um-date-separator)
                             (repeat 1 2 digit))
                      line-end
                      )
  "um date regexp. built from `um-date-separator'. NOTE: currently assumes
ISO8601."
  )

;; NOTE: regexp-opt wraps the whole expression in (), so I can't omit the -.
(defcustom um-locale-re (regexp-opt '(
                                      "- home"
                                      "- away"
                                      ))
  "um locale regexp. Add your custom places within a `regexp-opt' or
 make it general."
  :type '(string)
  :group 'um
  )

(defconst um-tag-regexp "^\\+ \\([[:alpha:]\\_\\-]+\\)$"
  "um tag regexp. Allows hypens and underscores within the tag.")

(defface font-lock-um-date-face
  `((((type tty) (class mono)))
    (t (
        :inherit shadow
        )))
  "um date face"
  :group 'um
  )

(defface font-lock-um-locale-face
  `((((type tty) (class mono)))
    (t (
        :inherit shadow
        )))
  "um locale face"
  :group 'um
  )

(defface font-lock-um-tag-face
  `((((type tty) (class mono)))
    (t (
        :inherit shadow
        )))
  "um tag face"
  :group 'um
  )

(font-lock-add-keywords 'markdown-mode
                        `(
                          (,um-date-re 1 'font-lock-um-date-face)
                          (,um-locale-re 0 'font-lock-um-locale-face)
                          (,um-tag-regexp 1 'font-lock-um-tag-face)
                          ))

(defvar um-minor-mode-keywords
  `(
    (,um-date-re 1 'font-lock-um-date-face)
    (,um-locale-re 0 'font-lock-um-locale-face)
    (,um-tag-regexp 1 'font-lock-um-tag-face)
    ))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; journal header

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

(defun um-header-current-buffer ()
  (car (split-string (buffer-substring-no-properties (point-min) (point-max))
                     "\n\n" t
                     )))

(defun um--header-end-pos ()
  (save-excursion
    (goto-char (point-min))
    (search-forward "\n\n")))

(defun um-tag-first-in-current-buffer ()
  (let ((header (um-header-current-buffer)))
    (string-match um-tag-regexp header)
    (match-string 1 header)
    ))

;;;###autoload
(defun um-tag-grep ()
  "Run `project-find-regexp' on a selection made from `um-tags-history' via
  `completing-read'.

The initial value provided to `completing-read' is the first tag found in the
current buffer: it will be first in the list and available via \\`M-n'.

NOTE: searches in the current project root by default, but
\\[universal-argument] will allow choice of the base directory as in
`project-find-regexp'.

Ultimately this relies on `xref-matches-in-files', which calls
`xref-search-program'.
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

(defun um--extract-tags (files)
  (let ((tags))
    (dolist (f files)
      (with-current-buffer
          (find-file-noselect f t)
        (save-excursion
          (goto-char (point-min))
          (let* ((bound (um--header-end-pos))
                 (found t))
            (while found
              (setq found (search-forward-regexp um-tag-regexp bound t))
              (when found (add-to-list 'tags (buffer-substring-no-properties
                                              (match-beginning 1) (match-end 1))))
              )))))
    (reverse tags)))

(defun um-tag-insert (tag)
  "Insert TAG as + tag\n in current buffer appending to the journal header.

Emits message if TAG is already present, but does not error."
  (if (member tag (um--extract-tags (list buffer-file-name)))
      ;; NOTE: user-error would stop any iteration over files:
      (message "\"%s\" tag already exists in %s" tag buffer-file-name)
    (goto-char (um--header-end-pos))
    (forward-line -1)
    (insert (concat "+ " tag "\n"))))

(defun um-tag-delete (tag)
  "Delete TAG from the journal header in current buffer."
  (goto-char (point-min))
  (when (search-forward (concat "+ " tag "\n") (um--header-end-pos) t)
    (delete-region (match-beginning 0) (match-end 0))))

(defun um-tag-do (tag insert)
  "Insert or delete TAG from the journal header in current buffer.

insert when INSERT > 0, delete otherwise."
  (if (> insert 0)
      (um-tag-insert tag)
    (um-tag-delete tag)))

;;;###autoload
(defun um-tag-dwim (ARG)
  "Run `um-tag-do' on a list of filenames if region active outside
  dired-mode, or if marks exist in dired-mode, or the filename at point, and
  finally in the current buffer if none of those conditions match.

Assumes the files of interest are returned by `um-root-path'.

Negative prefix arg is handled by `um-tag-do', which see.
"
  (interactive "p")
  (let* (
         (insert (> ARG 0))
         (prompt (format "um tag %s: " (if insert "insert" "delete")))
         (marks (if (eq major-mode 'dired-mode) (dired-get-marked-files) nil))
         ;; NOTE: should avoid bogus strings when in a markdown buffer:
         (fap (um-find-file-at-point))
         (files (cond
                 ((and (region-active-p) (not (eq major-mode 'dired-mode)))
                  (string-split (buffer-substring (region-beginning) (region-end))))
                 (marks marks)
                 (fap (list fap))
                 (t (list (buffer-file-name)))
                 ))
         (collection (if insert um-tags-history (um--extract-tags files)))
         ;; no need for dupes:
         (history-delete-duplicates t)
         ;; override sorting when deleting, because we sort the tags:
         (completions-sort (if insert completions-sort nil))
         (vertico-sort-function (if insert vertico-sort-function nil))
         (tag (completing-read prompt collection nil nil nil
                               'um-tags-history)))

    (dolist (f files)
      (with-current-buffer
          ;; TODO: this should use the fallback logic:
          (find-file-noselect (expand-file-name f (um-root-path)) t)
        (save-excursion
          (um-tag-do tag ARG)
          (save-buffer))))
    ))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; shell integration

;;;###autoload
(defun um-next (next-file &optional tags)
  "Takes the next-file generated by the shell script and opens it.

Optional TAGS string may contain more than one tag separated by a comma.
"
  (find-file next-file)
  (um-journal-header)
  (if tags
      (progn (forward-line -1)
             (dolist (tag (split-string tags ","))
               (insert (format "+ %s\n"
                               (if (equal "+" tag)
                                   (cadr (split-string next-file "\\."))
                                 tag))))
             (forward-line)))
  (message (format "um next: %s" next-file)))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; um-minor-mode : in markdown files under markdown-mode

(defvar-keymap um-minor-mode-map
  :doc "Keymap for `um-minor-mode'."
  "M-s t" #'um-tag-grep
  ;; M-r is not exactly right, but can't think of a better binding:
  "M-r t" #'um-tag-dwim
  )

;;;###autoload
(define-minor-mode um-minor-mode
  "Minor mode for `um' commands.
\\{um-minor-mode-map}"
  :init-value nil
  :lighter " um"
  :keymap um-minor-mode-map
  (if um-minor-mode
      (progn
        (font-lock-add-keywords nil um-minor-mode-keywords)
        (font-lock-flush))
    (font-lock-remove-keywords nil um-minor-mode-keywords)
    (font-lock-flush))
  )

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; um-mode : in .um files

(defun um-inhibit-read-only (cmd &optional args)
  (let ((inhibit-read-only t))
    (apply cmd args)))

(defun um-drag-stuff-up () (interactive) (um-inhibit-read-only 'drag-stuff-up '(1)))
(defun um-drag-stuff-down () (interactive) (um-inhibit-read-only 'drag-stuff-down '(1)))
(defun um-kill-line () (interactive) (um-inhibit-read-only 'kill-line))
(defun um-kill-region () (interactive) (um-inhibit-read-only 'kill-region '(nil nil t)))
(defun um-yank () (interactive) (um-inhibit-read-only 'yank))
(defun um-tag-dwim-inhibit-read-only (arg) (interactive "p")
       (um-inhibit-read-only 'um-tag-dwim (list arg)))

;; NOTE: define-derived-mode defines a sparse keymap with the parent mode. But
;; we can prefill it:
(defvar-keymap um-mode-map
  :doc "Keymap for um-mode."
  "n" 'next-line
  "p" 'previous-line
  "M-p" 'um-drag-stuff-up
  "M-n" 'um-drag-stuff-down
  "C-k" 'um-kill-line
  "C-w" 'um-kill-region
  "C-y" 'um-yank
  "t" 'um-tag-dwim-inhibit-read-only
  )

;;;###autoload
(define-derived-mode um-mode special-mode "um-mode"
  "major mode for *.um file lists. Mostly `special-mode' with a few more commands.

\\{um-mode-map}
"
  (read-only-mode)
  )

(provide 'um)
