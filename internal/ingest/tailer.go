package ingest

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"go-log-aggregator/internal/config"
)

func StartTailer(ctx context.Context, source config.Source, out chan<- Event, errs chan<- error) error {
	if out == nil {
		return fmt.Errorf("event channel is required")
	}

	sourcePath, err := filepath.Abs(source.Path)
	if err != nil {
		return fmt.Errorf("resolve source path: %w", err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}

	dir := filepath.Dir(sourcePath)
	if err := watcher.Add(dir); err != nil {
		_ = watcher.Close()
		return fmt.Errorf("watch directory: %w", err)
	}

	go func() {
		defer watcher.Close()

		var file *os.File
		var partialLine string
		emitLine := func(line string) {
			select {
			case <-ctx.Done():
				return
			case out <- Event{
				SourceName: source.Name,
				SourcePath: sourcePath,
				Line:       line,
				ReceivedAt: time.Now(),
			}:
			}
		}

		closeFile := func() {
			if file == nil {
				return
			}
			_ = file.Close()
			file = nil
		}

		openFile := func(startAtEnd bool) error {
			closeFile()

			f, err := os.Open(sourcePath)
			if err != nil {
				return err
			}

			if startAtEnd {
				if _, err := f.Seek(0, io.SeekEnd); err != nil {
					_ = f.Close()
					return err
				}
			} else if _, err := f.Seek(0, io.SeekStart); err != nil {
				_ = f.Close()
				return err
			}

			file = f
			partialLine = ""
			return nil
		}

		if err := openFile(true); err != nil && !errors.Is(err, os.ErrNotExist) {
			notifyError(errs, fmt.Errorf("open %s: %w", source.Name, err))
		}

		for {
			select {
			case <-ctx.Done():
				closeFile()
				return
			case err := <-watcher.Errors:
				if err != nil {
					notifyError(errs, fmt.Errorf("watcher %s: %w", source.Name, err))
				}
			case event := <-watcher.Events:
				if !sameFile(event.Name, sourcePath) {
					continue
				}

				if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
					closeFile()
					continue
				}

				if event.Op&fsnotify.Create != 0 && file == nil {
					if err := openFile(false); err != nil && !errors.Is(err, os.ErrNotExist) {
						notifyError(errs, fmt.Errorf("open %s: %w", source.Name, err))
					} else if file != nil {
						if err := readAvailable(file, &partialLine, emitLine); err != nil {
							notifyError(errs, fmt.Errorf("read %s: %w", source.Name, err))
						}
					}
				}

				if event.Op&fsnotify.Write != 0 {
					if file == nil {
						if err := openFile(false); err != nil {
							if !errors.Is(err, os.ErrNotExist) {
								notifyError(errs, fmt.Errorf("open %s: %w", source.Name, err))
							}
							continue
						}
					}

					if err := readAvailable(file, &partialLine, emitLine); err != nil {
						notifyError(errs, fmt.Errorf("read %s: %w", source.Name, err))
					}
				}
			}
		}
	}()

	return nil
}

func readAvailable(file *os.File, partialLine *string, emit func(string)) error {
	reader := bufio.NewReader(file)
	for {
		chunk, err := reader.ReadString('\n')
		if len(chunk) > 0 {
			if strings.HasSuffix(chunk, "\n") {
				line := strings.TrimRight(chunk, "\r\n")
				line = *partialLine + line
				*partialLine = ""
				emit(line)
			} else {
				*partialLine += chunk
			}
		}

		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func sameFile(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func notifyError(errs chan<- error, err error) {
	if errs == nil || err == nil {
		return
	}

	select {
	case errs <- err:
	default:
	}
}
