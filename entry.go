package timectl

type Entry interface {
	Less(Entry) bool
	Sum(Entry) Entry
}

type Object interface {
	Less(Object) bool
	Sum(Object) Object
	MergeKey() interface{}
	Key() interface{}
}

type Any interface{}

type ObjectPoint struct {
	Any
	Time int64 `json:"time"`
}
