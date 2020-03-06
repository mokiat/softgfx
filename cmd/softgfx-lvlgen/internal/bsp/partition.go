package bsp

func Partition(walls []*Wall, precision float64) *Wall {
	if len(walls) == 0 {
		return nil
	}

	var root *Wall
	for _, wall := range walls {
		if root == nil {
			root = wall
		} else {
			root.Insert(wall, precision)
		}
	}
	return root
}
