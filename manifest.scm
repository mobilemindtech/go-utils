(use-modules (guix packages)
             (guix download)
             (guix gexp)
             (guix utils)
             (guix licenses)
             (gnu packages)
             (gnu packages base)
             (gnu packages compression)
             (guix build-system trivial)
             ((gnu packages) #:select (specifications->manifest))
             (ice-9 eval-string)
             (system base compile))

(define-public go-1.24
  (package
    (name "go")
    (version "1.24.3")
    (source (origin
              (method url-fetch)
              (uri (string-append "https://go.dev/dl/go" version ".linux-amd64.tar.gz"))
              (sha256 (base32 "1n5xj7iyhcwspkk11366np3ah13jmjjfm5c80zlp3adgagmgccrk"))))
    (build-system trivial-build-system) ; apenas extrai sem compilar
    (inputs (list coreutils)) ; general require
    (native-inputs (list coreutils tar gzip)) ; require by "build"
    (outputs '("out"))
    (arguments
     '(#:builder
        (begin

          ; set path with deps
          (setenv "PATH" (string-append (assoc-ref %build-inputs "coreutils") "/bin:"
                                        (assoc-ref %build-inputs "tar") "/bin:"
                                        (assoc-ref %build-inputs "gzip") "/bin:"
                                        (getenv "PATH")))
          (define out (assoc-ref %outputs "out"))
          (mkdir out)  ; Criar diretório de saída

          ; extract tar.gz to out
          (system* "tar" "xzf" (assoc-ref %build-inputs "source") "-C" out)  ; Extração do Go

          ; go links
          (let ((bin-dir (string-append out "/bin")))
            (mkdir bin-dir)
            (system* "ln" "-sf" (string-append out "/go/bin/go") (string-append bin-dir "/go"))
            (system* "ln" "-sf" (string-append out "/go/bin/godoc") (string-append bin-dir "/godoc"))
            (system* "ln" "-sf" (string-append out "/go/bin/gofmt") (string-append bin-dir "/gofmt"))))))
    (synopsis "Go programming language")
    (description "Go is an open source programming language that makes it easy to build simple, reliable, and efficient software.")
    (home-page "https://go.dev/")
    (license bsd-3)))

(packages->manifest
 (list
  (specification->package "bash")
  (specification->package "grep")
  (specification->package "git")
  (specification->package "coreutils")
  (specification->package "node@22")
  go-1.24))




