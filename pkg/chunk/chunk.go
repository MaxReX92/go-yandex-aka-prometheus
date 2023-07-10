package chunk

func SliceToChunks[T any](array []T, chunkSize int) [][]T {
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

func ChanToChunks[T any](ch <-chan T, chunkSize int) [][]T {
	var result [][]T
	var chunk []T
	for item := range ch {
		if len(chunk) < chunkSize {
			chunk = append(chunk, item)
			continue
		}

		result = append(result, chunk)
		chunk = []T{}
	}

	if len(chunk) != 0 {
		result = append(result, chunk)
	}

	return result
}
