package timectl

type Entry interface {
	Less(Entry) bool
	Sum(Entry) Entry
}
