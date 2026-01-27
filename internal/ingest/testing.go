package ingest

import "os"

// ReadAvailableForTest exposes readAvailable for tests in /tests.
func ReadAvailableForTest(file *os.File, partialLine *string, emit func(string)) error {
	return readAvailable(file, partialLine, emit)
}
