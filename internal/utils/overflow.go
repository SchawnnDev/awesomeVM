package utils

// CheckAdditionOverflow checks if the addition of two signed integers results in an overflow.
func CheckAdditionOverflow[T int64 | int32 | int16 | int8 | byte](a, b, sum T) bool {
	return ((T(a) > 0) && (T(b) > 0) && (sum < 0)) || ((T(a) < 0) && (T(b) < 0) && (sum > 0))
}

// CheckSubtractionOverflow checks if the subtraction of two signed integers results in an overflow.
func CheckSubtractionOverflow[T int64 | int32 | int16 | int8 | byte](a, b, diff T) bool {
	return ((T(a) < 0) && (T(b) > 0) && (diff > 0)) || ((T(a) > 0) && (T(b) < 0) && (diff < 0))
}
