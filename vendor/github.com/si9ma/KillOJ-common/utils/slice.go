package utils

func ContainsInt(arr []int, target int) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}

	return false
}

func ContainsInt64(arr []int64, target int64) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}

	return false
}

func ContainsFloat64(arr []float64, target float64) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}

	return false
}

func ContainsString(arr []string, target string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}

	return false
}
