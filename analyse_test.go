package cwrapmsg_test

import (
	"testing"

	"github.com/mrsubudei/cwrapmsg"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc   string
		config *cwrapmsg.Config
	}{
		{
			desc:   "./cwrapmsg/default/...",
			config: &cwrapmsg.Config{},
		},
		{
			desc: "./cwrapmsg/withflags/...",
			config: &cwrapmsg.Config{
				IgnoreCyrillic:  true,
				CaseInsensitive: true,
			},
		},
		{
			// known limitations: incorrect messages which are not reported
			desc: "./cwrapmsg/limitations/...",
			config: &cwrapmsg.Config{
				IgnoreCyrillic: true,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			a := cwrapmsg.NewAnalyzer(test.config)

			analysistest.Run(t, analysistest.TestData(), a, test.desc)
		})
	}
}

func TestAnalyzerFlags(t *testing.T) {
	t.Parallel()

	a := cwrapmsg.NewAnalyzer(nil)

	require.NotNil(t, a.Flags.Lookup("ignore-cyrillic"))
	require.NotNil(t, a.Flags.Lookup("case-insensitive"))

	require.NoError(t, a.Flags.Set("ignore-cyrillic", "true"))
	require.NoError(t, a.Flags.Set("case-insensitive", "true"))

	analysistest.Run(t, analysistest.TestData(), a, "./cwrapmsg/withflags/...")
}
