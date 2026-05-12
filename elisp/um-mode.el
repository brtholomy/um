;;; um-mode.el -*- lexical-binding: t -*-

(defun um-inhibit-read-only (cmd &optional args)
  (let ((inhibit-read-only t))
    (apply cmd args)))

(defun um-drag-stuff-up () (interactive) (um-inhibit-read-only 'drag-stuff-up '(1)))
(defun um-drag-stuff-down () (interactive) (um-inhibit-read-only 'drag-stuff-down '(1)))
(defun um-kill-line () (interactive) (um-inhibit-read-only 'kill-line))
(defun um-tag-dwim-inhibit-read-only (arg) (interactive "p")
       (um-inhibit-read-only 'um-tag-dwim (list arg)))

;; NOTE: define-derived-mode defines a sparse keymap with the parent mode. But
;; we can prefill it:
(defvar-keymap um-mode-map
  :doc "Keymap for um-mode."
  "M-p" 'um-drag-stuff-up
  "M-n" 'um-drag-stuff-down
  "M-k" 'um-kill-line
  "t" 'um-tag-dwim-inhibit-read-only
  )

(define-derived-mode um-mode view-mode "um-mode"
  "major mode for *.um file lists. Mostly `view-mode' with a few special commands.

\\{um-mode-map}
"
  (setq-local mode-line-show-line-column nil)
  )

;;;###autoload
(add-to-list 'auto-mode-alist '("\\.um\\'" . um-mode))

(provide 'um-mode)
