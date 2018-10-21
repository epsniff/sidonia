package govector

import (
	"fmt"
)

// AsVector converts slices of numeric types into a Vector.
func AsVector(any interface{}) (Vector, error) {
	switch x := any.(type) {
	case []uint8:
		return uint8ToVector(x), nil
	case []uint16:
		return uint16ToVector(x), nil
	case []uint32:
		return uint32ToVector(x), nil
	case []uint64:
		return uint64ToVector(x), nil
	case []int:
		return intToVector(x), nil
	case []int8:
		return int8ToVector(x), nil
	case []int16:
		return int16ToVector(x), nil
	case []int32:
		return int32ToVector(x), nil
	case []int64:
		return int64ToVector(x), nil
	case []float32:
		return float32ToVector(x), nil
	case []float64:
		return float64ToVector(x), nil
	default:
		return nil, fmt.Errorf("Unable to coerce input of type %T into vector", any)
	}
}

func uint8ToVector(x []uint8) Vector {
	y := make(Vector, len(x))

	for i, _ := range x {
		y[i] = float64(x[i])
	}
	return y
}

func uint16ToVector(x []uint16) Vector {
	y := make(Vector, len(x))

	for i, _ := range x {
		y[i] = float64(x[i])
	}
	return y
}

func uint32ToVector(x []uint32) Vector {
	y := make(Vector, len(x))

	for i, _ := range x {
		y[i] = float64(x[i])
	}
	return y
}

func uint64ToVector(x []uint64) Vector {
	y := make(Vector, len(x))

	for i, _ := range x {
		y[i] = float64(x[i])
	}
	return y
}

func intToVector(x []int) Vector {
	y := make(Vector, len(x))

	for i, _ := range x {
		y[i] = float64(x[i])
	}
	return y
}

func int8ToVector(x []int8) Vector {
	y := make(Vector, len(x))

	for i, _ := range x {
		y[i] = float64(x[i])
	}
	return y
}

func int16ToVector(x []int16) Vector {
	y := make(Vector, len(x))

	for i, _ := range x {
		y[i] = float64(x[i])
	}
	return y
}

func int32ToVector(x []int32) Vector {
	y := make(Vector, len(x))

	for i, _ := range x {
		y[i] = float64(x[i])
	}
	return y
}

func int64ToVector(x []int64) Vector {
	y := make(Vector, len(x))

	for i, _ := range x {
		y[i] = float64(x[i])
	}
	return y
}

func float32ToVector(x []float32) Vector {
	y := make(Vector, len(x))

	for i, _ := range x {
		y[i] = float64(x[i])
	}
	return y
}

func float64ToVector(x []float64) Vector {
	return Vector(x)
}
