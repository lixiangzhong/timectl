package timectl

import (
	"sort"
	"time"
)

type Point struct {
	Entry Entry `json:"entry"`
	Ctime int64 `json:"ctime"`
}

func NewPoint(v Entry) Point {
	return Point{
		Entry: v,
		Ctime: time.Now().Unix(),
	}
}

func FillZero(src []Point, empty Entry, start, end, interval int64) []Point {
	sort.Slice(src, func(i, j int) bool {
		return src[i].Ctime < src[j].Ctime
	})
	var result = make([]Point, 0)
	for _, v := range src {
		if v.Ctime > end {
			break
		}
		for start < v.Ctime {
			result = append(result, Point{Ctime: start, Entry: empty})
			start += interval
		}
		result = append(result, v)
		start += interval
	}

	for start < end {
		result = append(result, Point{Ctime: start, Entry: empty})
		start += interval
	}
	return result
}
