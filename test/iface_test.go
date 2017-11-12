package test

import (
	"fmt"
	"reflect"
	"testing"
)

func TestIface(t *testing.T) {
	iface := &IfaceMock{
		OneFunc: func(str string, variadic ...string) (string, []string) {
			return str, variadic
		},
		TwoFunc: func(x, y int) int {
			return x + y
		},
	}
	type one struct {
		str      string
		variadic []string
		err      error
	}
	ones := []one{
		{
			str:      "firststr",
			variadic: []string{"one", "two"},
			err:      fmt.Errorf("firsterr"),
		},
		{
			str:      "secondstr",
			variadic: []string{"one", "two"},
			err:      fmt.Errorf("seconderr"),
		},
	}
	for _, o := range ones {
		actstr, actvariadic := iface.One(o.str, o.variadic...)
		if actstr != o.str {
			t.Errorf("str = %v, want %v", actstr, o.str)
		}
		if !reflect.DeepEqual(actvariadic, o.variadic) {
			t.Errorf("variadic = %v, want %v", actvariadic, o.variadic)
		}
	}
	if len(iface.OneCalls()) != len(ones) {
		t.Errorf("onecalls = %v, want %v", len(iface.OneCalls()), len(ones))
	}
	z := iface.Two(1, 2)
	if z != 3 {
		t.Errorf("z = %v, want %v", z, 3)
	}
	iface.Reset()
	if len(iface.OneCalls()) != 0 {
		t.Errorf("onecalls = %v, want %v", len(iface.OneCalls()), 0)
	}
	if len(iface.TwoCalls()) != 0 {
		t.Errorf("twocalls = %v, want %v", len(iface.TwoCalls()), 0)
	}
}
