package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

const (
	// GCM nonce size (12 bytes is standard for GCM)
	GCMNonceSize = 12
	// GCM tag size (16 bytes)
	GCMTagSize = 16
)

// GCMEncryptingWriter implements io.WriteCloser for streaming AES-GCM encryption
type GCMEncryptingWriter struct {
	writer     io.Writer
	gcm        cipher.AEAD
	nonce      []byte
	buffer     []byte
	closed     bool
	cipherBuf  []byte
	totalBytes int64
}

// NewGCMEncryptingWriter creates a new streaming GCM encrypting writer
func NewGCMEncryptingWriter(w io.Writer, key []byte) (*GCMEncryptingWriter, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, GCMNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Write nonce to the beginning of the output
	if _, err := w.Write(nonce); err != nil {
		return nil, fmt.Errorf("failed to write nonce: %w", err)
	}

	return &GCMEncryptingWriter{
		writer:    w,
		gcm:       gcm,
		nonce:     nonce,
		buffer:    make([]byte, 0, 64*1024),         // 64KB buffer
		cipherBuf: make([]byte, 64*1024+GCMTagSize), // Buffer for encrypted data
	}, nil
}

// Write encrypts and writes data to the underlying writer
func (w *GCMEncryptingWriter) Write(p []byte) (int, error) {
	if w.closed {
		return 0, fmt.Errorf("writer is closed")
	}

	originalLen := len(p)
	w.buffer = append(w.buffer, p...)
	w.totalBytes += int64(len(p))

	// Process data in chunks when buffer gets large enough
	if len(w.buffer) >= 32*1024 { // 32KB threshold
		if err := w.flushBuffer(false); err != nil {
			return 0, err
		}
	}

	return originalLen, nil
}

// flushBuffer encrypts and writes buffered data
func (w *GCMEncryptingWriter) flushBuffer(final bool) error {
	if len(w.buffer) == 0 && !final {
		return nil
	}

	// For streaming GCM, we need to use a different approach
	// We'll encrypt the data and write it, but save the final tag for Close()
	if final {
		// This is the final flush - encrypt all remaining data
		if len(w.buffer) > 0 {
			// gcm.Seal returns ciphertext + tag together
			encrypted := w.gcm.Seal(nil, w.nonce, w.buffer, nil)

			// Write the encrypted data (ciphertext + tag)
			if _, err := w.writer.Write(encrypted); err != nil {
				return fmt.Errorf("failed to write encrypted data: %w", err)
			}
		}
	} else {
		// For intermediate chunks, we need a different approach
		// Since GCM requires all data at once, we'll buffer until Close()
		// This is a limitation of GCM - it's not truly streaming like CBC
		return nil
	}

	w.buffer = w.buffer[:0]
	return nil
}

// Close finalizes the encryption and writes the authentication tag
func (w *GCMEncryptingWriter) Close() error {
	if w.closed {
		return nil
	}
	w.closed = true

	// Encrypt all buffered data and write it with the tag
	return w.flushBuffer(true)
}

// GCMDecryptingReader implements io.Reader for streaming AES-GCM decryption
type GCMDecryptingReader struct {
	reader    io.Reader
	gcm       cipher.AEAD
	nonce     []byte
	buffer    []byte
	plainBuf  []byte
	remaining []byte
	finished  bool
}

// NewGCMDecryptingReader creates a new streaming GCM decrypting reader
func NewGCMDecryptingReader(r io.Reader, key []byte) (*GCMDecryptingReader, error) {
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Read nonce from the beginning of the input
	nonce := make([]byte, GCMNonceSize)
	if _, err := io.ReadFull(r, nonce); err != nil {
		return nil, fmt.Errorf("failed to read nonce: %w", err)
	}

	return &GCMDecryptingReader{
		reader:   r,
		gcm:      gcm,
		nonce:    nonce,
		buffer:   make([]byte, 0, 64*1024),
		plainBuf: make([]byte, 64*1024),
	}, nil
}

// Read decrypts and returns data from the underlying reader
func (r *GCMDecryptingReader) Read(p []byte) (int, error) {
	if r.finished && len(r.remaining) == 0 {
		return 0, io.EOF
	}

	// If we have remaining decrypted data, return it first
	if len(r.remaining) > 0 {
		n := copy(p, r.remaining)
		r.remaining = r.remaining[n:]
		return n, nil
	}

	// Read all remaining data (GCM requires complete data to decrypt)
	if !r.finished {
		allData, err := io.ReadAll(r.reader)
		if err != nil {
			return 0, fmt.Errorf("failed to read encrypted data: %w", err)
		}

		if len(allData) < GCMTagSize {
			return 0, fmt.Errorf("encrypted data too short")
		}

		// The data format is: ciphertext + tag (as returned by gcm.Seal)
		// We can pass this directly to gcm.Open
		plaintext, err := r.gcm.Open(nil, r.nonce, allData, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to decrypt data: %w", err)
		}

		r.remaining = plaintext
		r.finished = true
	}

	// Return decrypted data
	if len(r.remaining) > 0 {
		n := copy(p, r.remaining)
		r.remaining = r.remaining[n:]
		return n, nil
	}

	return 0, io.EOF
}
