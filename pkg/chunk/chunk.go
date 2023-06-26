package chunk

func Chunk[T any](array []T, chunkSize int) [][]T {
	var result [][]T
	arrayLen := len(array)

	for i := 0; i < arrayLen; i += chunkSize {
		j := i + chunkSize
		if j > arrayLen {
			j = arrayLen
		}
		result = append(result, array[i:j])
	}
	return result
}
