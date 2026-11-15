package plugin_test

import (
	"path/filepath"
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	_ "github.com/mrsubudei/cwrapmsg/plugin"
)

func TestPlugin(t *testing.T) {
	t.Parallel()

	newPlugin, err := register.GetPlugin("cwrapmsg")
	require.NoError(t, err)

	p, err := newPlugin(map[string]any{
		"ignore-cyrillic":  true,
		"case-insensitive": true,
	})
	require.NoError(t, err)
	require.Equal(t, register.LoadModeSyntax, p.GetLoadMode())

	analyzers, err := p.BuildAnalyzers()
	require.NoError(t, err)
	require.Len(t, analyzers, 1)

	testdata, err := filepath.Abs(filepath.Join("..", "testdata"))
	require.NoError(t, err)

	analysistest.Run(t, testdata, analyzers[0], "./cwrapmsg/withflags/...")
}

func TestPluginDefaultSettings(t *testing.T) {
	t.Parallel()

	newPlugin, err := register.GetPlugin("cwrapmsg")
	require.NoError(t, err)

	p, err := newPlugin(nil)
	require.NoError(t, err)

	analyzers, err := p.BuildAnalyzers()
	require.NoError(t, err)

	testdata, err := filepath.Abs(filepath.Join("..", "testdata"))
	require.NoError(t, err)

	analysistest.Run(t, testdata, analyzers[0], "./cwrapmsg/default/...")
}

func TestPluginUnknownSetting(t *testing.T) {
	t.Parallel()

	newPlugin, err := register.GetPlugin("cwrapmsg")
	require.NoError(t, err)

	_, err = newPlugin(map[string]any{"unknown": true})
	require.Error(t, err)
}
