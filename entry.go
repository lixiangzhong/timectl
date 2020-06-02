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

type any interface{}

type ObjectPoint struct {
	any
	Time int64 `json:"time"`
}

func NewObjectPoint(val interface{}, time int64) ObjectPoint {
	return ObjectPoint{
		any:  val,
		Time: time,
	}
}
