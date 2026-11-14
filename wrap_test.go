package cwrapmsg

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
)

func Test_findWrapCalls(t *testing.T) {
	t.Parallel()

	var (
		src = `
		package main

		import (
			"github.com/pkg/errors"
			"fmt"
		)

		type Item struct{}

		func (i Item) Check(ctx context.Context) error {}

		func (i *Item) test(ctx context.Context) error {
			item, err := getItem(ctx)
			if err != nil {
				return fmt.Errorf("getItem: %w", err)
			}

			err = item.Check(ctx)
			if err != nil {
				return errors.Wrapf(err, "Check: %d", 5)
			}

			return run(func() error {
				return errors.Wrap(err, "ошибка")
			})
		}
	`
		wantWrapData = []WrapData{
			{
				line:          16,
				errName:       "err",
				message:       "getItem: %w",
				parentFunc:    "Item_test",
				nestingsLevel: 1,
				isErrorf:      true,
			},
			{
				line:          21,
				errName:       "err",
				message:       "Check: %d",
				parentFunc:    "Item_test",
				nestingsLevel: 1,
				isWrapf:       true,
			},
			{
				line:          25,
				errName:       "err",
				message:       "ошибка",
				parentFunc:    "Item_test",
				nestingsLevel: 2,
			},
		}
		wantErrsMap = map[string]struct{}{
			"err": {},
		}
	)

	file, fset := parseFile(t, src)

	pass := &analysis.Pass{
		Fset:  fset,
		Files: []*ast.File{file},
	}

	gotWrapData, gotErrsMap := findWrapCalls(pass, file, fset, Config{}, "errors", "fmt")

	require.Len(t, gotWrapData, len(wantWrapData))
	for idx := range wantWrapData {
		gotWrapData[idx].node = nil
		require.Equal(t, wantWrapData[idx], gotWrapData[idx])
	}

	require.Equal(t, wantErrsMap, gotErrsMap)

	// messages with cyrillic symbols are skipped with IgnoreCyrillic
	gotWrapData, _ = findWrapCalls(pass, file, fset, Config{IgnoreCyrillic: true}, "errors", "fmt")
	require.Len(t, gotWrapData, 2)
}

func Test_isWrapMsgSuitable(t *testing.T) {
	t.Parallel()

	type args struct {
		config   Config
		funcName string
		message  string
	}

	var tests = []struct {
		name string
		args args
		want bool
	}{
		{
			name: "complete match",
			args: args{
				funcName: "s.store.GetByIDs",
				message:  "s.store.GetByIDs",
			},
			want: true,
		},
		{
			name: "part of name",
			args: args{
				funcName: "s.store.GetByIDs",
				message:  "store.GetByIDs: %w",
			},
			want: true,
		},
		{
			name: "underscores in message",
			args: args{
				funcName: "getUserIDs",
				message:  "get_User_IDs",
			},
			want: true,
		},
		{
			name: "underscores in func name",
			args: args{
				funcName: "get_user_ids",
				message:  "getuserids",
			},
			want: true,
		},
		{
			name: "case-sensitive by default",
			args: args{
				funcName: "getUserIDs",
				message:  "getuserids",
			},
			want: false,
		},
		{
			name: "case insensitive",
			args: args{
				config: Config{
					CaseInsensitive: true,
				},
				funcName: "getUserIDs",
				message:  "get_user_ids: %w",
			},
			want: true,
		},
		{
			name: "method of call result",
			args: args{
				funcName: "*ast.CallExpr.Execx",
				message:  "readConn.Execx",
			},
			want: true,
		},
		{
			name: "method of call result, case insensitive",
			args: args{
				config: Config{
					CaseInsensitive: true,
				},
				funcName: "*ast.CallExpr.Execx",
				message:  "execx",
			},
			want: true,
		},
		{
			name: "method of call result, wrong name",
			args: args{
				funcName: "*ast.CallExpr.Execx",
				message:  "readConn",
			},
			want: false,
		},
		{
			name: "not suitable",
			args: args{
				funcName: "getUserIDs",
				message:  "get user ids, %w",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isWrapMsgSuitable(tt.args.config, tt.args.funcName, tt.args.message)

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_checkStandard(t *testing.T) {
	t.Parallel()

	type args struct {
		funcName string
		message  string
	}

	var tests = []struct {
		name string
		args args
		want bool
	}{
		{
			name: "complete matches",
			args: args{
				message:  "entity.getUserIDs",
				funcName: "entity.getUserIDs",
			},
			want: true,
		},
		{
			name: "partly matches",
			args: args{
				message:  "getUserIDs",
				funcName: "entity.getUserIDs",
			},
			want: true,
		},
		{
			name: "extra word",
			args: args{
				message:  "getUserIDs.extra",
				funcName: "entity.getUserIDs",
			},
			want: false,
		},
		{
			name: "not match",
			args: args{
				message:  "e.getUserIDs",
				funcName: "entity.getUserIDs",
			},
			want: false,
		},
		{
			name: "exceed",
			args: args{
				message:  "extra.entity.getUserIDs",
				funcName: "entity.getUserIDs",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := checkStandard(Config{}, tt.args.funcName, tt.args.message)

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_cutMessage(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "no suffix",
			message: "store.GetByIDs",
			want:    "store.GetByIDs",
		},
		{
			name:    "colon",
			message: "GetByIDs: %w",
			want:    "GetByIDs",
		},
		{
			name:    "space",
			message: "get by ids",
			want:    "get",
		},
		{
			name:    "dash",
			message: "GetByIDs-failed",
			want:    "GetByIDs",
		},
		{
			name:    "symbol at the beginning",
			message: ",GetByIDs",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := cutMessage(tt.message)

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_getWIndex(t *testing.T) {
	t.Parallel()

	type wants struct {
		wantInt  int
		wantBool bool
	}

	var tests = []struct {
		name    string
		message string
		wants   wants
	}{
		{
			name:    "no 'w' verb",
			message: "user id: %d, val: %s",
			wants: wants{
				wantInt:  0,
				wantBool: false,
			},
		},
		{
			name:    "more than one 'w' verb",
			message: "err: %w, err2: %w",
			wants: wants{
				wantInt:  1,
				wantBool: false,
			},
		},
		{
			name:    "'w' verb at the end",
			message: "id: %d, err: %w",
			wants: wants{
				wantInt:  1,
				wantBool: true,
			},
		},
		{
			name:    "'w' verb at the beginning",
			message: "err: %w, id: %d",
			wants: wants{
				wantInt:  0,
				wantBool: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotInt, gotBool := getWIndex(tt.message)

			require.Equal(t, tt.wants.wantInt, gotInt)
			require.Equal(t, tt.wants.wantBool, gotBool)
		})
	}
}

func Test_hasCyrillic(t *testing.T) {
	t.Parallel()

	require.True(t, hasCyrillic("ошибка: %w"))
	require.True(t, hasCyrillic("getItem: ошибка"))
	require.False(t, hasCyrillic("getItem: %w"))
}

func Test_getPackageNameOrAlias(t *testing.T) {
	t.Parallel()

	src := `
		package main

		import (
			pkgerrors "github.com/pkg/errors"
			"fmt"
		)
	`

	file, _ := parseFile(t, src)

	require.Equal(t, "pkgerrors", getPackageNameOrAlias(file, "errors", "github.com/pkg/errors"))
	require.Equal(t, "fmt", getPackageNameOrAlias(file, "fmt", "fmt"))
}
