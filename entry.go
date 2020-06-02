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

type ObjectPoint struct {
	Object interface{} `json:"object"`
	Time   int64       `json:"time"`
}

func NewObjectPoint(val interface{}, time int64) ObjectPoint {
	return ObjectPoint{
		Object: val,
		Time:   time,
	}
}
