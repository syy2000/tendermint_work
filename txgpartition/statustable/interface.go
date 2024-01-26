package statustable

type Stringer interface {
	String() string
}

type Table interface {
	get(key string) (Stringer, bool)
	set(key string, value Stringer) bool
	clear()
	hash() []byte
}

type TableOption func(Table)
