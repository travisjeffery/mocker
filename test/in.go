package test

type Iface interface {
	One(str string, variadic ...string) (string, []string)
	Two(int, int) int
}
