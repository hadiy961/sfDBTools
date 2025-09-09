package dir

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// PermissionInfo berisi informasi permission direktori
type PermissionInfo struct {
	Path  string      `json:"path"`
	Mode  os.FileMode `json:"mode"`
	Owner string      `json:"owner"`
	Group string      `json:"group"`
	UID   int         `json:"uid"`
	GID   int         `json:"gid"`
}

// GetModeString mengembalikan mode dalam format string (misal: drwxr-xr-x)
func (p *PermissionInfo) GetModeString() string {
	return p.Mode.String()
}

// GetOctalMode mengembalikan mode dalam format octal (misal: 755)
func (p *PermissionInfo) GetOctalMode() string {
	return fmt.Sprintf("%o", p.Mode.Perm())
}

// UserInfo berisi informasi user system
type UserInfo struct {
	UID      int    `json:"uid"`
	GID      int    `json:"gid"`
	Username string `json:"username"`
	Name     string `json:"name"`
	HomeDir  string `json:"home_dir"`
}

// String implements Stringer interface
func (u *UserInfo) String() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s (%s)", u.Username, u.Name)
	}
	return fmt.Sprintf("%s (UID:%d, GID:%d)", u.Username, u.UID, u.GID)
}

// DirectoryStats berisi statistik direktori
type DirectoryStats struct {
	Path         string        `json:"path"`
	TotalFiles   int           `json:"total_files"`
	TotalDirs    int           `json:"total_dirs"`
	TotalSize    int64         `json:"total_size"`
	LargestFile  *Entry        `json:"largest_file,omitempty"`
	OldestFile   *Entry        `json:"oldest_file,omitempty"`
	NewestFile   *Entry        `json:"newest_file,omitempty"`
	LastScanned  time.Time     `json:"last_scanned"`
	ScanDuration time.Duration `json:"scan_duration"`
}

// GetFormattedTotalSize mengembalikan total size dalam format yang mudah dibaca
func (d *DirectoryStats) GetFormattedTotalSize() string {
	return formatSize(d.TotalSize)
}

// GetAverageFileSize menghitung rata-rata ukuran file
func (d *DirectoryStats) GetAverageFileSize() int64 {
	if d.TotalFiles == 0 {
		return 0
	}
	return d.TotalSize / int64(d.TotalFiles)
}

// GetFormattedAverageSize mengembalikan rata-rata ukuran dalam format yang mudah dibaca
func (d *DirectoryStats) GetFormattedAverageSize() string {
	return formatSize(d.GetAverageFileSize())
}

// SortOrder untuk sorting entries
type SortOrder int

const (
	SortByName SortOrder = iota
	SortBySize
	SortByModTime
	SortByType
)

// SortDirection untuk arah sorting
type SortDirection int

const (
	Ascending SortDirection = iota
	Descending
)

// SortEntries mengurutkan slice entries berdasarkan kriteria
func SortEntries(entries []Entry, order SortOrder, direction SortDirection) {
	sort.Slice(entries, func(i, j int) bool {
		var less bool

		switch order {
		case SortByName:
			less = strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
		case SortBySize:
			less = entries[i].Size < entries[j].Size
		case SortByModTime:
			less = entries[i].ModTime.Before(entries[j].ModTime)
		case SortByType:
			// Direktori dulu, kemudian file berdasarkan extension
			if entries[i].IsDir != entries[j].IsDir {
				less = entries[i].IsDir
			} else {
				extI := strings.ToLower(filepath.Ext(entries[i].Name))
				extJ := strings.ToLower(filepath.Ext(entries[j].Name))
				less = extI < extJ
			}
		}

		if direction == Descending {
			return !less
		}
		return less
	})
}

// FilterCombiner untuk menggabungkan multiple filter
type FilterCombiner struct {
	filters []FilterFunc
	mode    CombineMode
}

// CombineMode menentukan cara penggabungan filter
type CombineMode int

const (
	AND CombineMode = iota // Semua filter harus match
	OR                     // Minimal satu filter harus match
)

// NewFilterCombiner membuat combiner baru
func NewFilterCombiner(mode CombineMode) *FilterCombiner {
	return &FilterCombiner{
		filters: make([]FilterFunc, 0),
		mode:    mode,
	}
}

// Add menambahkan filter ke combiner
func (fc *FilterCombiner) Add(filter FilterFunc) *FilterCombiner {
	fc.filters = append(fc.filters, filter)
	return fc
}

// Build membuat FilterFunc yang menggabungkan semua filter
func (fc *FilterCombiner) Build() FilterFunc {
	return func(entry Entry) bool {
		if len(fc.filters) == 0 {
			return true
		}

		switch fc.mode {
		case AND:
			for _, filter := range fc.filters {
				if !filter(entry) {
					return false
				}
			}
			return true
		case OR:
			for _, filter := range fc.filters {
				if filter(entry) {
					return true
				}
			}
			return false
		}
		return true
	}
}

// DirectoryTreeNode untuk representasi tree struktur direktori
type DirectoryTreeNode struct {
	Entry    Entry                `json:"entry"`
	Children []*DirectoryTreeNode `json:"children,omitempty"`
	Parent   *DirectoryTreeNode   `json:"-"` // Exclude dari JSON untuk avoid circular reference
	Level    int                  `json:"level"`
}

// AddChild menambahkan child node
func (n *DirectoryTreeNode) AddChild(child *DirectoryTreeNode) {
	child.Parent = n
	child.Level = n.Level + 1
	n.Children = append(n.Children, child)
}

// GetPath mengembalikan full path dari root ke node ini
func (n *DirectoryTreeNode) GetPath() string {
	if n.Parent == nil {
		return n.Entry.Name
	}
	return filepath.Join(n.Parent.GetPath(), n.Entry.Name)
}

// IsLeaf mengecek apakah node ini adalah leaf (tidak punya children)
func (n *DirectoryTreeNode) IsLeaf() bool {
	return len(n.Children) == 0
}

// CountNodes menghitung total nodes dalam subtree
func (n *DirectoryTreeNode) CountNodes() int {
	count := 1 // Count self
	for _, child := range n.Children {
		count += child.CountNodes()
	}
	return count
}

// Walk melakukan traversal DFS pada tree
func (n *DirectoryTreeNode) Walk(visitFunc func(*DirectoryTreeNode) error) error {
	if err := visitFunc(n); err != nil {
		return err
	}

	for _, child := range n.Children {
		if err := child.Walk(visitFunc); err != nil {
			return err
		}
	}

	return nil
}

// ProgressCallback untuk operasi yang membutuhkan progress reporting
type ProgressCallback func(current, total int, currentItem string)

// OperationResult untuk hasil operasi yang kompleks
type OperationResult struct {
	Success        bool          `json:"success"`
	ProcessedCount int           `json:"processed_count"`
	SuccessCount   int           `json:"success_count"`
	ErrorCount     int           `json:"error_count"`
	Errors         []string      `json:"errors,omitempty"`
	Warnings       []string      `json:"warnings,omitempty"`
	Duration       time.Duration `json:"duration"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
}

// AddError menambahkan error ke result
func (r *OperationResult) AddError(err error) {
	r.ErrorCount++
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}
}

// AddWarning menambahkan warning ke result
func (r *OperationResult) AddWarning(warning string) {
	r.Warnings = append(r.Warnings, warning)
}

// Finish mengakhiri operasi dan menghitung durasi
func (r *OperationResult) Finish() {
	r.EndTime = time.Now()
	r.Duration = r.EndTime.Sub(r.StartTime)
	r.Success = r.ErrorCount == 0
}

// GetSuccessRate menghitung success rate dalam persen
func (r *OperationResult) GetSuccessRate() float64 {
	if r.ProcessedCount == 0 {
		return 0
	}
	return float64(r.SuccessCount) / float64(r.ProcessedCount) * 100
}

// NewOperationResult membuat OperationResult baru
func NewOperationResult() *OperationResult {
	return &OperationResult{
		StartTime: time.Now(),
		Errors:    make([]string, 0),
		Warnings:  make([]string, 0),
	}
}

// Platform-specific constants
var (
	// PathSeparator adalah separator path untuk platform saat ini
	PathSeparator = string(filepath.Separator)

	// MaxPathLength adalah panjang maksimum path untuk platform saat ini
	MaxPathLength = getMaxPathLength()

	// ReservedNames adalah nama file/direktori yang reserved per platform
	ReservedNames = getReservedNames()
)

// getMaxPathLength mengembalikan panjang path maksimum per platform
func getMaxPathLength() int {
	switch runtime.GOOS {
	case "windows":
		return 260 // MAX_PATH di Windows (tapi bisa lebih dengan long path support)
	case "darwin":
		return 1024 // macOS PATH_MAX
	default:
		return 4096 // Linux dan Unix lainnya PATH_MAX
	}
}

// getReservedNames mengembalikan nama yang reserved per platform
func getReservedNames() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			"CON", "PRN", "AUX", "NUL",
			"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
			"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
		}
	default:
		return []string{} // Unix systems umumnya tidak punya reserved names
	}
}

// IsReservedName mengecek apakah nama adalah reserved name
func IsReservedName(name string) bool {
	upperName := strings.ToUpper(name)
	for _, reserved := range ReservedNames {
		if upperName == reserved {
			return true
		}
	}
	return false
}

// ValidatePath melakukan validasi path berdasarkan platform
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path tidak boleh kosong")
	}

	if len(path) > MaxPathLength {
		return fmt.Errorf("path terlalu panjang: %d karakter (maksimum: %d)", len(path), MaxPathLength)
	}

	// Check reserved names
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for _, part := range parts {
		if IsReservedName(part) {
			return fmt.Errorf("nama '%s' adalah reserved name di platform ini", part)
		}
	}

	// Platform-specific validations
	if runtime.GOOS == "windows" {
		return validateWindowsPath(path)
	}

	return nil
}

// validateWindowsPath melakukan validasi khusus Windows
func validateWindowsPath(path string) error {
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(path, char) {
			return fmt.Errorf("path mengandung karakter tidak valid untuk Windows: %s", char)
		}
	}

	// Check untuk trailing spaces atau dots (tidak diizinkan Windows)
	parts := strings.Split(path, string(filepath.Separator))
	for _, part := range parts {
		if part != "" && (strings.HasSuffix(part, " ") || strings.HasSuffix(part, ".")) {
			return fmt.Errorf("nama direktori tidak boleh diakhiri dengan spasi atau titik di Windows: '%s'", part)
		}
	}

	return nil
}
