# Rencana Migrasi ke utils/dir

## Tahap 1: Buat Paket utils/dir
1. **Buat struktur dasar**:
   ```bash
   mkdir -p utils/dir
   touch utils/dir/{manager.go,permissions.go,scanner.go,cleanup.go,types.go}
   ```

2. **Pindahkan DirectoryManager dari utils/file/directory_operation.go**:
   - Copy dan refactor ke `utils/dir/manager.go`
   - Tambahkan method tambahan yang dibutuhkan
   - Pertahankan backward compatibility functions

## Tahap 2: Konsolidasi Fungsi Duplikat
1. **Dari utils/disk/disk_other.go**:
   ```go
   // Ganti implementasi dengan delegasi ke utils/dir
   func CreateOutputDirectory(outputDir string) error {
       return dir.Create(outputDir)
   }
   
   func ValidateOutputDir(outputDir string) error {
       return dir.Validate(outputDir)
   }
   ```

2. **Dari utils/file/file_permission.go**:
   - Pindahkan `CreateDirectoryWithPermissions` ke `utils/dir/permissions.go`
   - Pindahkan `SetDirectoryPermissions` ke `utils/dir/permissions.go`

## Tahap 3: Refactor Scanner Operations
1. **Dari utils/dbconfig/filemanager.go**:
   ```go
   // Ganti os.ReadDir dengan utils/dir/scanner
   func (fm *FileManager) ListConfigFiles() ([]*FileInfo, error) {
       scanner := dir.NewScanner()
       entries, err := scanner.FindByExtension(fm.configDir, ".cnf.enc")
       // ... rest of logic
   }
   ```

2. **Dari utils/common/config_utils.go**:
   ```go
   func FindEncryptedConfigFiles(dirPath string) ([]string, error) {
       scanner := dir.NewScanner()
       return scanner.FindByExtension(dirPath, ".cnf.enc")
   }
   ```

3. **Dari utils/backup/list.go**:
   ```go
   func ResolveDBListFile(cmd *cobra.Command) (string, error) {
       scanner := dir.NewScanner()
       files, err := scanner.FindByExtension("./config/db_list", ".txt")
       // ... rest of logic
   }
   ```

## Tahap 4: Refactor Cleanup Operations
1. **Dari utils/backup/cleanup.go**:
   ```go
   func CleanupOldBackups(outputDir string, retentionDays int) ([]string, error) {
       manager := dir.NewManager()
       return manager.CleanupOldDirectories(outputDir, retentionDays, "2006_01_02")
   }
   ```

2. **Dari utils/dbconfig/filemanager.go**:
   ```go
   func (fm *FileManager) CleanupBackups(days int) (int, error) {
       manager := dir.NewManager()
       removed, err := manager.CleanupOldFiles(fm.configDir, days, ".backup.")
       return len(removed), err
   }
   ```

## Tahap 5: Update Import Statements
1. **Update semua file yang menggunakan fungsi direktori**:
   ```go
   // Sebelum
   import "sfDBTools/utils/file"
   file.CreateDir(path)
   
   // Sesudah  
   import "sfDBTools/utils/dir"
   dir.Create(path)
   ```

2. **Update cmd/ files**:
   - Cari semua file di `cmd/` yang menggunakan operasi direktori
   - Update import dan function calls

## Tahap 6: Deprecation dan Cleanup
1. **Tandai fungsi lama sebagai deprecated**:
   ```go
   // Deprecated: Use utils/dir.Create instead
   func CreateDir(path string) error {
       return dir.Create(path)
   }
   ```

2. **Setelah migrasi selesai, hapus fungsi duplikat**:
   - Hapus `utils/file/directory_operation.go`
   - Hapus fungsi direktori dari `utils/disk/disk_other.go`
   - Cleanup import statements

## Checklist Migrasi
- [ ] Buat struktur utils/dir
- [ ] Pindahkan DirectoryManager 
- [ ] Konsolidasi fungsi create/validate
- [ ] Refactor scanner operations
- [ ] Refactor cleanup operations
- [ ] Update import statements di cmd/
- [ ] Update import statements di internal/
- [ ] Testing semua fungsi
- [ ] Hapus kode duplikat
- [ ] Update dokumentasi

## Testing Strategy
1. **Unit tests untuk utils/dir**
2. **Integration tests untuk operasi direktori**
3. **Regression tests untuk backward compatibility**
4. **Performance tests untuk operasi besar**