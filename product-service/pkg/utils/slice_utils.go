package utils

func SliceContains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func SliceRemove(slice []string, val string) (result []string) {
	for _, item := range slice {
		if item == val {
			continue
		}
		result = append(result, item)
	}
	return
}

func SliceChunk(slice []string, chunkSize int) [][]string {
	var chunks [][]string
	for i := 0; i < len(slice); i += chunkSize {
		chunk := slice[i:min(i+chunkSize, len(slice))]
		chunks = append(chunks, chunk)
	}
	return chunks
}
