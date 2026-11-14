# testdata

Test cases for [analysistest](https://pkg.go.dev/golang.org/x/tools/go/analysis/analysistest),
the expected diagnostics are marked with `// want` comments. The directory is a separate module.

- `cwrapmsg/default` — the default settings: `positive.go` contains incorrect messages which are reported,
  `negative.go` contains correct messages and cases which are not checked.
- `cwrapmsg/withflags` — the same with `-ignore-cyrillic` and `-case-insensitive`.
- `cwrapmsg/limitations` — known limitations: incorrect messages which are not reported,
  the lines are marked with `// NOT FOUND:`. These cases are rare and are not handled on purpose,
  see "Known limitations" in the main README.
