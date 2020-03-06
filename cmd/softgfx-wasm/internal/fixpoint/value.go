package fixpoint

const precisionBits = 12

type Value int32

func (v Value) Floor() int {
	return int(v >> precisionBits)
}

func (v Value) Times(count int) Value {
	return v * Value(count)
}

func FromInt(value int) Value {
	return Value(value << precisionBits)
}

func FromFloat32(value float32) Value {
	return Value(value * (1 << precisionBits))
}
