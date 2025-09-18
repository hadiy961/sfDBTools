package fs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"sfDBTools/internal/logger"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
	"golang.org/x/term"
)

// fileOperations mengimplementasikan FileOperations interface
type fileOperations struct {
	fs     afero.Fs
	logger *logger.Logger
}

// newFileOperations membuat instance file operations baru
func newFileOperations(fs afero.Fs, logger *logger.Logger) FileOperations {
	return &fileOperations{
		fs:     fs,
		logger: logger,
	}
}

// Copy menyalin file dari src ke dst dengan permission default
func (f *fileOperations) Copy(src, dst string) error {
	info, err := f.fs.Stat(src)
	if err != nil {
		return fmt.Errorf("gagal stat file src %s: %w", src, err)
	}
	return f.CopyWithInfo(src, dst, info)
}

// CopyWithInfo menyalin file dengan informasi FileInfo yang diberikan
func (f *fileOperations) CopyWithInfo(src, dst string, info os.FileInfo) error {
	if err := f.EnsureDir(filepath.Dir(dst)); err != nil {
		return fmt.Errorf("gagal pastikan parent dir: %w", err)
	}

	srcFile, err := f.fs.Open(src)
	if err != nil {
		return fmt.Errorf("gagal buka file src: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := f.fs.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("gagal buat file dst: %w", err)
	}
	defer func() {
		if cerr := dstFile.Close(); cerr != nil {
			f.logger.Warn("Gagal tutup file dst", logger.String("file", dst), logger.Error(cerr))
		}
	}()

	// Decide between interactive CLI progress bar or logger-based progress.
	total := info.Size()

	useProgressBar := false
	if total > 0 {
		if term.IsTerminal(int(os.Stdout.Fd())) {
			useProgressBar = true
		}
	}

	if useProgressBar {
		// Create a progress bar that writes to stdout. Throttle updates to
		// avoid excessive redraws.
		// Include human-readable size in the description so user sees "Copying foo (1.2 MiB)"
		desc := fmt.Sprintf("Copying %s (%s)", filepath.Base(src), humanizeBytes(uint64(total)))
		bar := progressbar.NewOptions64(
			total,
			progressbar.OptionSetWriter(os.Stdout),
			progressbar.OptionSetWidth(40),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetDescription(desc),
			progressbar.OptionThrottle(100*time.Millisecond),
			progressbar.OptionSpinnerType(14),
		)

		// Use io.TeeReader so reads update the progress bar automatically.
		if _, err := io.Copy(dstFile, io.TeeReader(srcFile, bar)); err != nil {
			return fmt.Errorf("gagal copy konten file dengan progressbar: %w", err)
		}
		// Ensure bar finishes
		_ = bar.Close()
		fmt.Println()
		// f.logger.Info("Copy completed ", logger.String("src", src), logger.String("dst", dst), logger.Int64("total", total))
	} else {
		// Non-interactive or unknown terminal: fallback to logger-based chunked copy
		buf := make([]byte, 32*1024)
		var copied int64
		var lastLoggedPercent int64 = -1

		for {
			n, rerr := srcFile.Read(buf)
			if n > 0 {
				wn, werr := dstFile.Write(buf[:n])
				if werr != nil {
					return fmt.Errorf("gagal tulis ke file dst: %w", werr)
				}
				if wn != n {
					return fmt.Errorf("incomplete write to dst: wrote %d of %d", wn, n)
				}
				copied += int64(wn)

				if total > 0 {
					percent := copied * 100 / total
					if percent != lastLoggedPercent {
						if percent%5 == 0 || percent == 100 {
							lastLoggedPercent = percent
							f.logger.Info("Copy progress", logger.String("src", src), logger.String("dst", dst), logger.Int64("percent", percent), logger.Int64("copied", copied), logger.Int64("total", total))
						} else {
							lastLoggedPercent = percent
						}
					}
				}
			}

			if rerr != nil {
				if rerr == io.EOF {
					break
				}
				return fmt.Errorf("gagal baca src: %w", rerr)
			}
		}

		if total == 0 {
			f.logger.Info("Copy progress", logger.String("src", src), logger.String("dst", dst), logger.Int64("percent", 100), logger.Int64("copied", 0), logger.Int64("total", 0))
		} else if copied >= total {

			f.logger.Info("Copy completed", logger.String("src", src), logger.String("dst", dst), logger.Int64("copied", copied), logger.Int64("total", total))
		}
	}

	// Set permission via filesystem jika didukung
	if err := f.fs.Chmod(dst, info.Mode().Perm()); err != nil {
		f.logger.Warn("Gagal set permission file", logger.String("file", dst), logger.Error(err))
	}

	// Coba preserve ownership menggunakan syscall
	if statT, ok := info.Sys().(*syscall.Stat_t); ok {
		if err := os.Chown(dst, int(statT.Uid), int(statT.Gid)); err != nil {
			f.logger.Debug("Chown failed (ignored)", logger.String("file", dst), logger.Error(err))
		}
	}

	return nil
}

// FormatSize mengubah ukuran byte menjadi string yang lebih mudah dibaca
// Move memindahkan file dari src ke dst
func (f *fileOperations) Move(src, dst string) error {
	if err := f.Copy(src, dst); err != nil {
		return err
	}
	return f.fs.Remove(src)
}

// EnsureDir memastikan direktori ada, buat jika tidak ada
func (f *fileOperations) EnsureDir(path string) error {
	if path == "" || path == "." {
		return nil
	}

	normalizedPath := filepath.Clean(path)
	return f.fs.MkdirAll(normalizedPath, 0755)
}

// WriteJSON menulis data sebagai JSON ke file
func (f *fileOperations) WriteJSON(path string, data interface{}) error {
	if err := f.EnsureDir(filepath.Dir(path)); err != nil {
		return fmt.Errorf("gagal pastikan parent dir: %w", err)
	}

	file, err := f.fs.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("gagal buat file JSON: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("gagal encode JSON: %w", err)
	}

	return nil
}
