package gossdb

var (
	byt              = []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	maxByteSize byte = 58
)

func ToNum(bs []byte) int {
	re := 0
	for _, v := range bs {
		if v >= maxByteSize {
			return 0
		}
		re = re*10 + byt[v]
	}
	return re
}
