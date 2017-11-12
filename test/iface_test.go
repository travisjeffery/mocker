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
		actstr, actvariadic, acterr := iface.One(o.str, o.variadic...)
		if actstr != o.str {
			t.Errorf("str = %v, want %v", actstr, o.str)
		}
		if !reflect.DeepEqual(actvariadic, o.variadic) {
			t.Errorf("variadic = %v, want %v", actvariadic, o.variadic)
		}
		if acterr.Error() != o.err.Error() {
			t.Errorf("acterr = %v, want %v", acterr.Error(), o.err.Error())
		}
	}
	if len(iface.OneCalls()) != len(ones) {
		t.Errorf("onecalls = %v, want %v", len(iface.OneCalls()), len(ones))
	}
	z := iface.Two(1, 2)
	if z != 3 {
		t.Errorf("z = %v, want %v", z, 3)
	}
}
