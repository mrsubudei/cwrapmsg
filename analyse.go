package cwrapmsg

import (
	"flag"
	"go/ast"
	"regexp"

	"golang.org/x/tools/go/analysis"
)

func NewAnalyzer(cfg *Config) *analysis.Analyzer {
	var flagSet flag.FlagSet
	if cfg == nil {
		cfg = &Config{}
		flagSet = flags(cfg)
	}

	analyser := &analysis.Analyzer{
		Name:  "cwrapmsg",
		Doc:   "Checks messages that are passed to errors.Wrap, errors.Wrapf and fmt.Errorf",
		URL:   "https://github.com/mrsubudei/cwrapmsg",
		Flags: flagSet,
		Run: func(pass *analysis.Pass) (any, error) {
			Run(pass, *cfg)

			return nil, nil
		},
	}

	return analyser
}

func flags(opts *Config) flag.FlagSet {
	fs := flag.NewFlagSet("cwrapmsg", flag.ContinueOnError)

	fs.BoolVar(&opts.IgnoreCyrillic, "ignore-cyrillic", false, "skip messages containing cyrillic symbols")
	fs.BoolVar(&opts.CaseInsensitive, "case-insensitive", false, "perform case-insensitive comparison")

	return *fs
}

func Run(pass *analysis.Pass, cfg Config) {
	for _, file := range pass.Files {
		if isGenerated(pass, file) {
			continue
		}

		wrapPackageName := getPackageNameOrAlias(file, "errors", "github.com/pkg/errors")
		fmtPackageName := getPackageNameOrAlias(file, "fmt", "fmt")

		handle(pass, file, cfg, wrapPackageName, fmtPackageName)
	}
}

var generatedFileNameRegexp = regexp.MustCompile(`\.pb\.|_gen\.go$`)

func isGenerated(pass *analysis.Pass, file *ast.File) bool {
	if ast.IsGenerated(file) {
		return true
	}

	return generatedFileNameRegexp.MatchString(pass.Fset.File(file.Pos()).Name())
}
