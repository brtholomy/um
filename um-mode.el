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
  "n" 'next-line
  "p" 'previous-line
  "M-p" 'um-drag-stuff-up
  "M-n" 'um-drag-stuff-down
  "M-k" 'um-kill-line
  "t" 'um-tag-dwim-inhibit-read-only
  )

;;;###autoload
(define-derived-mode um-mode special-mode "um-mode"
  "major mode for *.um file lists. Mostly `special-mode' with a few more commands.

\\{um-mode-map}
"
  (read-only-mode)
  )

(provide 'um-mode)
