package bridge

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode"
)

// exportedFuncs parses bridge.go and returns exported func name -> param type list.
func exportedFuncs(t *testing.T) map[string][]string {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), "bridge.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string][]string{}
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || !fn.Name.IsExported() {
			continue
		}
		var params []string
		for _, p := range fn.Type.Params.List {
			typ := typeName(p.Type)
			n := max(len(p.Names), 1)
			for i := 0; i < n; i++ {
				params = append(params, typ)
			}
		}
		out[fn.Name.Name] = params
	}
	return out
}

func typeName(e ast.Expr) string {
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return "?"
}

// TestExportedAPIGolden pins the gomobile API; LibrCore.ts/Kotlin depend on it.
// Regenerate testdata/bridge_api.golden deliberately when the API changes.
func TestExportedAPIGolden(t *testing.T) {
	funcs := exportedFuncs(t)
	var lines []string
	for name, params := range funcs {
		lines = append(lines, name+"("+strings.Join(params, ", ")+")")
	}
	sort.Strings(lines)
	got := strings.Join(lines, "\n") + "\n"
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.WriteFile("testdata/bridge_api.golden", []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile("testdata/bridge_api.golden")
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("bridge API changed.\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestKotlinCallsResolve checks every method LibrCoreModule.kt invokes by reflection
// exists with the same arity and string-only params; a mismatch fails silently on device.
func TestKotlinCallsResolve(t *testing.T) {
	src, err := os.ReadFile("../../android/app/src/main/java/com/libr/app/LibrCoreModule.kt")
	if err != nil {
		t.Fatal(err)
	}
	funcs := exportedFuncs(t)
	calls := regexp.MustCompile(`callBridgeString\("(\w+)"((?:,[^)]*)?)\)`).FindAllStringSubmatch(string(src), -1)
	if len(calls) < 20 {
		t.Fatalf("found only %d bridge calls in Kotlin; regex out of date?", len(calls))
	}
	for _, c := range calls {
		name := string(unicode.ToUpper(rune(c[1][0]))) + c[1][1:]
		args := 0
		if strings.TrimSpace(c[2]) != "" {
			args = strings.Count(c[2], ",")
		}
		params, ok := funcs[name]
		if !ok {
			t.Errorf("Kotlin calls %s, bridge has no %s", c[1], name)
			continue
		}
		if len(params) != args {
			t.Errorf("%s: Kotlin passes %d args, bridge takes %d", name, args, len(params))
		}
		for _, p := range params {
			if p != "string" {
				t.Errorf("%s: param type %s, Kotlin passes String", name, p)
			}
		}
	}
	for _, h := range []string{"SetNotificationHandler"} {
		if _, ok := funcs[h]; !ok {
			t.Errorf("Kotlin registers %s, bridge has none", h)
		}
	}
}
