package cwrapmsg

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
)

func parseFile(t *testing.T, src string) (*ast.File, *token.FileSet) {
	t.Helper()

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	require.NoError(t, err)

	return node, fset
}

func Test_getDiagReport(t *testing.T) {
	t.Parallel()

	type args struct {
		isWrapf  bool
		isErrorf bool
	}

	var tests = []struct {
		name string
		args args
		want string
	}{
		{
			name: "wrap",
			want: "message in errors.Wrap does not match function or method name",
		},
		{
			name: "wrapf",
			args: args{
				isWrapf: true,
			},
			want: "message in errors.Wrapf does not match function or method name",
		},
		{
			name: "errorf",
			args: args{
				isErrorf: true,
			},
			want: "message in fmt.Errorf does not match function or method name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			node := &ast.Ident{NamePos: 10, Name: "err"}

			got := getDiagReport(node, tt.args.isWrapf, tt.args.isErrorf)

			require.Equal(t, analysis.Diagnostic{Pos: 10, End: 13, Message: tt.want}, got)
		})
	}
}

func Test_getFuncCalls(t *testing.T) {
	t.Parallel()

	var (
		errNames = []string{"err"}
		src      = `
		package main

		import (
			"github.com/pkg/errors"
			"fmt"
		)

		var err = errors.New("package level")

		type Item struct{}

		func (i Item) Check(ctx context.Context) error {}

		func test(ctx context.Context) (func(context.Context) error, error) {
			item, err := getItem(ctx)
			if err != nil {
				return nil, fmt.Errorf("getItem: %w", err)
			}

			fn := func(ctx context.Context) error {
				err := genericFunc[Item, int](ctx, item)

				return errors.Wrap(err, "genericFunc")
			}

			err = item.Check(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "Check")
			}

			var err = getErr()

			return fn, err
		}
	`
		wantCallData = map[string][]CallData{
			"err-1": {
				{
					funcName:   "getItem",
					parentFunc: "test",
					line:       16,
				},
				{
					funcName:   "item.Check",
					parentFunc: "test",
					line:       27,
				},
				{
					funcName:   "getErr",
					parentFunc: "test",
					line:       32,
				},
			},
			"err-2": {
				{
					funcName:   "genericFunc",
					parentFunc: "test",
					line:       22,
				},
			},
		}
	)

	file, fset := parseFile(t, src)

	got := getFuncCalls(file, fset, errNames)

	require.Equal(t, wantCallData, got)
}

func Test_getFuncName(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name string
		src  string
		want string
	}{
		{
			name: "func",
			src:  "getItem()",
			want: "getItem",
		},
		{
			name: "method",
			src:  "s.store.GetByIDs()",
			want: "s.store.GetByIDs",
		},
		{
			name: "generic func",
			src:  "generic[int]()",
			want: "generic",
		},
		{
			name: "generic func with several type parameters",
			src:  "generic[string, int]()",
			want: "generic",
		},
		{
			name: "method of call result",
			src:  "c.readConn(ctx).Execx()",
			want: "*ast.CallExpr.Execx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			expr, err := parser.ParseExpr(tt.src)
			require.NoError(t, err)

			got := getFuncName(expr.(*ast.CallExpr).Fun)

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_findCallData(t *testing.T) {
	t.Parallel()

	callDataSl := []CallData{
		{funcName: "first", line: 5},
		{funcName: "second", line: 10},
	}

	_, found := findCallData(callDataSl, 5)
	require.False(t, found)

	got, found := findCallData(callDataSl, 6)
	require.True(t, found)
	require.Equal(t, "first", got.funcName)

	got, found = findCallData(callDataSl, 20)
	require.True(t, found)
	require.Equal(t, "second", got.funcName)
}
