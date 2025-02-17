package trees

type BaseTree[T any] interface {
	Empty() bool
	Size() int
	Clear()
	Values() []T
	String() string
}
