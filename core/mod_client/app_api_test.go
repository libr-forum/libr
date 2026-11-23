package main

import (
	"os"
	"reflect"
	"regexp"
	"sort"
	"testing"
)

// TestBindingsMatchApp checks the generated Wails bindings the frontend imports
// (frontend/wailsjs/go/main/App.d.ts) still match App's exported methods, so moving
// logic into core/libr cannot silently drop a binding.
func TestBindingsMatchApp(t *testing.T) {
	src, err := os.ReadFile("frontend/wailsjs/go/main/App.d.ts")
	if err != nil {
		t.Fatal(err)
	}
	var bound []string
	for _, m := range regexp.MustCompile(`export function (\w+)\(`).FindAllStringSubmatch(string(src), -1) {
		bound = append(bound, m[1])
	}
	var methods []string
	ty := reflect.TypeOf(&App{})
	for i := 0; i < ty.NumMethod(); i++ {
		methods = append(methods, ty.Method(i).Name)
	}
	sort.Strings(bound)
	sort.Strings(methods)
	if !reflect.DeepEqual(bound, methods) {
		t.Fatalf("App methods and App.d.ts differ (run `wails generate module` if intended)\nApp.d.ts: %v\nApp:      %v", bound, methods)
	}
}
