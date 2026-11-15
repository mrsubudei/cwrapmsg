// Package plugin registers cwrapmsg as a golangci-lint module plugin,
// see https://golangci-lint.run/docs/plugins/module-plugins/.
package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/mrsubudei/cwrapmsg"
)

//nolint:gochecknoinits // the registration of the plugin requires init
func init() {
	register.Plugin("cwrapmsg", New)
}

type cwrapmsgPlugin struct {
	cfg cwrapmsg.Config
}

// New creates the plugin with the settings from the golangci-lint configuration.
func New(settings any) (register.LinterPlugin, error) {
	cfg, err := register.DecodeSettings[cwrapmsg.Config](settings)
	if err != nil {
		return nil, err
	}

	return &cwrapmsgPlugin{cfg: cfg}, nil
}

func (p *cwrapmsgPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{cwrapmsg.NewAnalyzer(&p.cfg)}, nil
}

func (p *cwrapmsgPlugin) GetLoadMode() string {
	return register.LoadModeSyntax
}
