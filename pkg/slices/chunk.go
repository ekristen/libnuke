package slices

// Chunk splits a slice into chunks of the given size.
// Original Source: https://github.com/rebuy-de/aws-nuke/blob/c3ae17932f058f1867aab382182ecd837090961a/resources/util.go#L40
func Chunk[T any](slice []T, size int) [][]T {
	var chunks [][]T
	for i := 0; i < len(slice); {
		// Clamp the last chunk to the slice bound as necessary.
		end := size
		if l := len(slice[i:]); l < size {
			end = l
		}

		// Set the capacity of each chunk so that appending to a chunk does not
		// modify the original slice.
		chunks = append(chunks, slice[i:i+end:i+end])
		i += end
	}

	return chunks
}
