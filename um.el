;;; um.el --- An Ultralight database for Markdown composition. -*- lexical-binding: t -*-

;; by bth
;; Version: 0.1

;; Um is an ultralight database design and emacs toolkit for organizing writing
;; into larger compositions. It consists of two parts:
;;
;; 1. Elisp for functionality within emacs.
;; 2. Go for the commandline interface.
;;
;; This is somewhat like org-roam, except without any database dependency. (And
;; it assumes Markdown rather than org, which I don't care for.)
;;
;; This depends on a few simple ideas:
;;
;; 1. A numbered filename specification which serves as unique id. The
;; filesystem is the database.
;;
;; 2. A file header consisting of the title, date, and optional tags. These tags
;; can be used by the CLI to construct expressive queries.
;;
;; 3. Using a "root" project defined by `um-root-glob', where source files
;; should first be composed and where we can assume a file exists if not
;; elsewhere.
;;
;; 4. Using the built-in `project' package and the CLI to organize compositions
;; built from these source files.
;;
;; features provided:
;; `um-mode': for *.um files, which are lists of content files produced by "um tag".
;; `um-minor-mode': for *.md files, which are the content files.
;; `um-find-file-at-point' via `find-file': open a file under point in the
;; current project, falling back to a source directory.
;; `um-target-file-at-point-advice' via `embark-dwim': open file under point in all
;; known projects, falling back to a source directory.
;; `um-tag-grep': search files with some tag
;; `um-tag-dwim': insert or delete a tag in dired and other contexts.

(require 'project)
(require 'dired)

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;; defcustom

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
  "Seperator used in date strings, used by `um--date-format' and `um--date-re'."
  :type '(string)
  :group 'um
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

(defconst um--date-format (concat "%Y" um-date-separator "%m" um-date-separator "%d")
  "Format passed to `format-time-string' when creating
 `um--header'. ISO8601 with a custom `um-date-separator'.")

(defconst um--date-re (rx
                       line-start
                       (literal ": ")
                       (group (repeat 4 digit)
                              (literal um-date-separator)
                              (repeat 1 2 digit)
                              (literal um-date-separator)
                              (repeat 1 2 digit))
                       line-end
                       )
  "um date regexp. built from `um-date-separator'."
  )

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;; find-file-at-point

(defun um--root-path ()
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
;;;; `find-file' integration
;;
;; combined with the hook into `file-name-at-point-functions', this means we can
;; run `find-file' at point, followed by `next-history-element'. By default:
;; C-x C-f M-n

;;;###autoload
(defun um-find-file-at-point ()
  "Return full file path of thing-at-point, falling back to:

1. `default-directory'
2. `project-root'
3. `um--root-path'
"
  (let ((dirs (list
               default-directory
               (when (project-current) (project-root (project-current)))
               (um--root-path)))
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

(defun um--filelist-dwim ()
  "Return a list of filenames obtained from marks in dired, from the active
  region, or from the file at point, or the current file."
  (let (
        (marks (if (eq major-mode 'dired-mode) (dired-get-marked-files) nil))
        (fap (um-find-file-at-point))
        )
    (cond
     ((and (region-active-p) (not (eq major-mode 'dired-mode)))
      (string-split (buffer-substring-no-properties (region-beginning) (region-end))))
     (marks marks)
     (fap (list fap))
     (t (list (buffer-file-name)))
     )))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;;; `embark-dwim' integration
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
3. `um--root-path'
"
  (or
   ;; this does both default-directory and project-root:
   (funcall origfunc)
   (let ((default-directory (um--root-path)))
     (funcall origfunc))
   ))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;; tags

(defconst um--tag-regexp "^\\+ \\([[:alpha:]\\_\\-]+\\)$"
  "um tag regexp. Allows hypens and underscores within the tag.")

(defconst um--header-end "\n\n" "um header is delimited by the first two newlines.")

;; NOTE: this will get saved by savehist-mode
(defvar um-tags-history nil "History of inserted or searched for tags. Populates
`completing-read'.")

(defun um--header-end-pos ()
  (save-excursion
    (goto-char (point-min))
    (search-forward um--header-end)))

(defun um--tag-first-in-buffer ()
  (save-excursion
    (goto-char (point-min))
    (when (search-forward um--tag-regexp (um--header-end-pos) t)
      (match-string 1))))

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
                                                (um--tag-first-in-buffer))
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
              (setq found (search-forward-regexp um--tag-regexp bound t))
              (when found (add-to-list 'tags (buffer-substring-no-properties
                                              (match-beginning 1) (match-end 1))))
              )))))
    (reverse tags)))

(defun um--tag-insert (tag)
  "Insert TAG as + tag\n in current buffer appending to the journal header.

Emits message if TAG is already present, but does not error."
  (if (member tag (um--extract-tags (list buffer-file-name)))
      ;; NOTE: user-error would stop any iteration over files:
      (message "\"%s\" tag already exists in %s" tag buffer-file-name)
    (goto-char (um--header-end-pos))
    (forward-line -1)
    (insert (concat "+ " tag "\n"))))

(defun um--tag-delete (tag)
  "Delete TAG from the journal header in current buffer."
  (goto-char (point-min))
  (if (search-forward (concat "+ " tag "\n") (um--header-end-pos) t)
      (delete-region (match-beginning 0) (match-end 0))
    ;; NOTE: as long as collection was gathered with um--extract-tags, this
    ;; shouldn't happen:
    (message "\"%s\" tag not found in %s" tag buffer-file-name)))

(defun um--tag-do (tag insert)
  "Insert or delete TAG from the journal header in current buffer.

insert when INSERT > 0, delete otherwise."
  (save-excursion
    (if (> insert 0)
        (um--tag-insert tag)
      (um--tag-delete tag))))

;;;###autoload
(defun um-tag-dwim (ARG)
  "Run `um--tag-do' on a list of filenames if region active outside
  dired-mode, or if marks exist in dired-mode, or the filename at point, and
  finally in the current buffer if none of those conditions match.

Assumes the files of interest are returned by `um--root-path'.

Negative prefix arg is handled by `um--tag-do', which see.
"
  (interactive "p")
  (let* (
         (insert (> ARG 0))
         (prompt (format "um tag %s: " (if insert "insert" "delete")))
         (files (um--filelist-dwim))
         (collection (if insert um-tags-history (um--extract-tags files)))
         ;; no need for dupes:
         (history-delete-duplicates t)
         ;; override sorting when deleting, because we sort the tags:
         (completions-sort (if insert completions-sort nil))
         (vertico-sort-function (if (and insert (bound-and-true-p vertico-sort-function))
                                    vertico-sort-function
                                  nil))
         (tag (completing-read prompt collection nil nil nil
                               'um-tags-history)))

    (dolist (f files)
      (with-current-buffer
          ;; TODO: this should use the fallback logic:
          (find-file-noselect (expand-file-name f (um--root-path)) t)
        (um--tag-do tag ARG)
        (save-buffer)))
    ))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;; ispell

(defun um-do-ispell ()
  "Run `ispell-buffer' on a list of files."
  (interactive)
  (dolist (f (um--filelist-dwim))
    (save-window-excursion
      (with-current-buffer
          ;; NOTE: not find-file-noselect because ispell needs to display the window:
          (find-file (expand-file-name f (um--root-path)) t)
        (ispell-buffer)
        ))))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;; CLI

(defun um--header (filename &optional tags)
  "Create a header composed of:

# filename
: date
+ tags

"
  (insert
   (format "# %s\n: %s%s" filename (format-time-string um--date-format) um--header-end))
  (when tags
    (forward-line -1)
    (dolist (tag (split-string tags ","))
      (insert (format "+ %s\n"
                      (if (equal "+" tag)
                          (cadr (split-string filename "\\."))
                        tag))))
    (forward-line)))

;;;###autoload
(defun um-next (filename &optional tags)
  "Takes the filename generated by the CLI, opens it, and adds the header.

Optional TAGS string may contain more than one tag separated by a comma.

A tag with a value of \"+\" is rendered as the descriptor portion of the filename.
"
  (find-file filename)
  (um--header filename tags)
  (message "um next: %s" filename))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;; um-minor-mode
;;
;; intended to be hooked into markdown-mode

(defface um-date-face
  `((((type tty) (class mono)))
    (t (
        :inherit shadow
        )))
  "um date face"
  :group 'um
  )

(defface um-locale-face
  `((((type tty) (class mono)))
    (t (
        :inherit shadow
        )))
  "um locale face"
  :group 'um
  )

(defface um-tag-face
  `((((type tty) (class mono)))
    (t (
        :inherit shadow
        )))
  "um tag face"
  :group 'um
  )

(defvar um--minor-mode-keywords
  `(
    (,um--date-re 1 'um-date-face)
    (,um-locale-re 0 'um-locale-face)
    (,um--tag-regexp 1 'um-tag-face)
    ))

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;;; section

;; Rather than migrate to org-mode or some other more complex specification,
;; I've opted to extend and augment my own idiom for structuring Markdown as
;; befits my style of writing: simple sections separated by a horizontal line.
;; This code is an attempt to make working with those sections easier.

(defconst um--section "\n---\n\n")

(defun um-section ()
  "Insert a horizontal rule between paragraphs, defining a section of text."
  (interactive)
  ;; reduce to either 1 or none: if 1, it'll be pushed to after the string.
  (delete-blank-lines)
  (insert um--section)
  ;; move to the beginning of the following text
  (re-search-forward "[^[:space:]]" nil t 1)
  ;; the previous leaves us just after the first character.
  (backward-char)
  ;; reduces to a single empty line before point.
  (ensure-empty-lines))

(defun um--goto-section (direction)
  (let ((origpos (point))
        (header (save-excursion
                  (goto-char (point-min))
                  (search-forward um--header-end nil t)))
        (pos (save-excursion
               (search-forward um--section nil t direction)))
        (endpos (match-end 0)))

    (cond ((= origpos (point-min)) (goto-char header))
          ((and pos (= 1 direction)) (goto-char pos))
          ;; when going backwards, the pos will be at the beginning of the
          ;; string, while the endpos will be where we want to end up. This allows
          ;; for the backwards movement to take you to that spot if you're within a
          ;; section, but move past it if you're already at this point.
          ((and pos (= -1 direction))
           (goto-char pos)
           (if (= origpos endpos)
               (um--goto-section -1)
             (goto-char endpos)))
          ;; pos will be nil when going back to beginning, so this clause will
          ;; match if the header is found:
          ((and header (= -1 direction)) (goto-char header))
          )))

(defun um-forward-section ()
  "Move to next section marker."
  (interactive) (um--goto-section 1))

(defun um-backward-section ()
  "Move to previous section marker."
  (interactive) (um--goto-section -1))

(defun um-mark-section ()
  "Mark text between section markers. Repeat to expand."
  (interactive)
  (let* ((pos (save-excursion
                (search-backward um--section nil t)))
         (m (if pos (match-end 0) (point))))
    (when (not (region-active-p)) (set-mark m))
    ;; NOTE: don't use um--goto-section because it doesn't go to eobp
    ;; 'noerror will take us to the end of buffer if no next
    (search-forward um--section nil 'noerror)))

(defun um--backward-section-permissive ()
  ;; NOTE: won't work without this hack when point is on line just below the
  ;; section, because the um--section includes 2 newlines.
  (search-backward (substring um--section 0 -1) nil t))

(defun um--delete-section ()
  (um--backward-section-permissive)
  (delete-region (match-beginning 0) (match-end 0))
  ;; to achieve single blank line before:
  (ensure-empty-lines)
  ;; and after:
  (delete-blank-lines))

(defun um-backward-kill-paragraph-or-section ()
  "Kills either a preceding paragraph or `um-section'"
  (interactive)
  (let ((section-pos) (text-pos))
    (save-excursion
      (setq section-pos (um--backward-section-permissive)))
    (save-excursion
      (setq text-pos (re-search-backward "[[:alpha:]]" nil t 1)))
    (if (and section-pos (> section-pos text-pos))
        (um--delete-section)
      (backward-kill-paragraph 1))))

(defun um-transpose-section (&optional arg)
  "Like transpose-paragraphs but for `um-section'.

A negative prefix argument moves it backward.
"
  (interactive "p")
  (let* ((direction (if (< arg 0)
                        -1 1
                        ))
         (origpos (point))
         (last (save-excursion
                 (not (search-forward um--section nil t))))
         (secondlast (save-excursion
                       (not (search-forward um--section nil t 2)))))
    (cond ((and last (= -1 direction))
           (um--goto-section -1)
           (um-transpose-section 1))
          ((and secondlast (= 1 direction))
           (um-mark-section)
           (kill-region nil nil t)
           (goto-char (point-max))
           (um-section)
           (yank)
           (um--delete-section))
          (t
           (um-mark-section)
           (kill-region nil nil t)
           (um--goto-section direction)
           (yank)))
    ;; to get us back to the top of the section region
    (set-mark-command 1)
    ))

(defvar-keymap um-minor-mode-map
  :doc "Keymap for `um-minor-mode'."
  "M-s t" #'um-tag-grep
  ;; M-r is not exactly right, but can't think of a better binding:
  "M-r t" #'um-tag-dwim
  ;; TODO: find a better default binding?
  "C-c s" #'um-section
  "M-o" #'um-backward-kill-paragraph-or-section
  "C-M-n" #'um-forward-section
  "C-M-p" #'um-backward-section
  "C-M-h" #'um-mark-section
  "C-M-t" #'um-transpose-section
  )

(bind-keys
 :repeat-map um-section-repeat-map
 :repeat-docstring "Keymap to repeat um key sequences
 `um-backward-section' and `um-forward-section'."
 ("p" . um-backward-section)
 ("n" . um-forward-section))

;;;###autoload
(define-minor-mode um-minor-mode
  "Minor mode for `um' commands.
\\{um-minor-mode-map}"
  :init-value nil
  :lighter " um"
  :keymap um-minor-mode-map
  (if um-minor-mode
      (progn
        (font-lock-add-keywords nil um--minor-mode-keywords)
        (font-lock-flush))
    (font-lock-remove-keywords nil um--minor-mode-keywords)
    (font-lock-flush))
  )

;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;;; um-mode : .um files

(defun um-next-line ()
  (interactive) (move-beginning-of-line nil) (next-line))

(defun um-previous-line ()
  (interactive) (move-beginning-of-line nil) (previous-line))

(defun um--reset-cursor-intangible-property ()
  (let ((modified (buffer-modified-p)))
    ;; NOTE: remove all first, because otherwise we might miss a newline:
    (remove-text-properties (point-min) (point-max)
                            '(
                              cursor-intangible t
                              ;; NOTE: no faces in yanked text:
                              face nil))
    (save-excursion
      (goto-char (point-min))
      (while (not (eobp))
        ;; NOTE: actually the newline is the only position not included, but this
        ;; positions the cursor visually at the bol:
        (add-text-properties (pos-bol) (pos-eol) '(cursor-intangible t))
        (forward-line 1)))
    ;; we don't want to fuck with the modified state just to change text
    ;; properties, but we also don't want to obscure it if it was already modified.
    (set-buffer-modified-p modified)))

(defun um--iro (cmd &optional args)
  (let ((inhibit-read-only t))
    (apply cmd args)
    (um--reset-cursor-intangible-property)))

(defun um-drag-stuff-up () (interactive) (um--iro 'drag-stuff-up '(1)))
(defun um-drag-stuff-down () (interactive) (um--iro 'drag-stuff-down '(1)))
(defun um-kill-line () (interactive) (um--iro 'kill-line))
(defun um-kill-region () (interactive) (um--iro 'kill-region '(nil nil t)))
;; NOTE: parallel to dired-copy-filename-as-kill
(defun um-copy-line-as-kill () (interactive) (kill-ring-save (pos-bol) (pos-eol))
       (message (buffer-substring-no-properties (pos-bol) (pos-eol))))
(defun um-yank () (interactive) (um--iro 'yank))
(defun um-tag-dwim-iro (arg) (interactive "p")
       (um--iro 'um-tag-dwim (list arg)))

;; NOTE: define-derived-mode defines a sparse keymap with the parent mode. But
;; we can prefill it:
(defvar-keymap um-mode-map
  :doc "Keymap for um-mode."
  "n" 'um-next-line
  "p" 'um-previous-line
  "M-p" 'um-drag-stuff-up
  "M-n" 'um-drag-stuff-down
  "C-k" 'um-kill-line
  "k" 'um-kill-line
  "C-w" 'um-kill-region
  "w" 'um-copy-line-as-kill
  "C-y" 'um-yank
  "t" 'um-tag-dwim-iro
  )

;;;###autoload
(define-derived-mode um-mode special-mode "um-mode"
  "major mode for *.um file lists. Mostly `special-mode' with a few more commands.

\\{um-mode-map}
"
  (let ((inhibit-read-only t))
    (um--reset-cursor-intangible-property))
  (cursor-intangible-mode)
  (read-only-mode))

(provide 'um)

;; Local Variables:
;; outline-regexp: ";;;+ [^;]+"
;; eval: (outline-minor-mode 1)
;; End:
