package utils

// CheckOverflow checks if the addition of two signed integers results in an overflow.
func CheckOverflow[T int64 | int32 | int16 | int8 | byte](a, b, sum T) bool {
	return ((T(a) > 0) && (T(b) > 0) && (sum < 0)) || ((T(a) < 0) && (T(b) < 0) && (sum > 0))
}
