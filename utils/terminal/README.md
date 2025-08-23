# Terminal Utilities

Package terminal menyediakan utilities untuk operasi terminal termasuk pembersihan layar, manipulasi cursor, indikator progress, dan formatting teks.

## Fitur Utama

### 1. Screen Clearing
- `Clear()` - Membersihkan layar terminal
- `ClearScreen()` - Membersihkan layar menggunakan system command
- `ClearScreenANSI()` - Membersihkan layar menggunakan ANSI escape sequences
- `ClearAndHome()` - Membersihkan layar dan pindah cursor ke home
- `ClearWithMessage()` - Membersihkan layar dan menampilkan pesan

### 2. Cursor Manipulation
- `MoveCursor(row, col)` - Memindahkan cursor ke posisi tertentu
- `MoveCursorHome()` - Memindahkan cursor ke home (1,1)
- `SaveCursorPosition()` - Menyimpan posisi cursor
- `RestoreCursorPosition()` - Mengembalikan posisi cursor
- `HideCursor()` - Menyembunyikan cursor
- `ShowCursor()` - Menampilkan cursor

### 3. Line Operations
- `ClearLines(n)` - Membersihkan n baris dari posisi cursor
- `ClearCurrentLine()` - Membersihkan baris saat ini
- `ClearToEndOfLine()` - Membersihkan dari cursor sampai akhir baris

### 4. Progress Indicators

#### Progress Spinner
```go
spinner := terminal.NewProgressSpinner("Loading...")
spinner.Start()
// Do some work...
spinner.UpdateMessage("Processing...")
// Do more work...
spinner.Stop()
```

#### Progress Bar
```go
bar := terminal.NewProgressBar(100, "Downloading")
for i := 0; i <= 100; i++ {
    bar.Update(i)
    time.Sleep(50 * time.Millisecond)
}
bar.Finish()
```

### 5. Colored Output
```go
terminal.PrintSuccess("Operation successful")
terminal.PrintError("An error occurred")
terminal.PrintWarning("Warning message")
terminal.PrintInfo("Information message")

// Custom colors
terminal.PrintColoredLine("Red text", terminal.ColorRed)
terminal.PrintColoredText("Blue text", terminal.ColorBlue)
```

### 6. Text Formatting
- `CenterText(text, width)` - Memusatkan teks
- `PadLeft(text, width)` - Padding kiri
- `PadRight(text, width)` - Padding kanan
- `TruncateText(text, width)` - Memotong teks

### 7. Headers and Separators
```go
terminal.PrintHeader("Main Title")
terminal.PrintSubHeader("Sub Title")
terminal.PrintSeparator()
terminal.PrintDashedSeparator()
```

### 8. Table Formatting
```go
headers := []string{"Name", "Status", "Date"}
rows := [][]string{
    {"Config 1", "Active", "2025-08-23"},
    {"Config 2", "Inactive", "2025-08-22"},
}
terminal.FormatTable(headers, rows)
```

### 9. Interactive Helpers
- `WaitForEnter()` - Menunggu user menekan Enter
- `PauseAndClear()` - Pause kemudian clear screen
- `ConfirmAndClear(question)` - Konfirmasi kemudian clear screen
- `ShowMenuAndClear(title, options)` - Menampilkan menu, ambil pilihan, clear screen

### 10. Interactive Menu
```go
menu := &terminal.InteractiveMenu{
    Title: "Main Menu",
    Options: []string{
        "Option 1",
        "Option 2", 
        "Option 3",
    },
    OnSelect: func(choice int) error {
        // Handle selection
        return nil
    },
    OnExit: func() error {
        // Handle exit
        return nil
    },
}
menu.Show()
```

## Terminal Information
- `GetTerminalSize()` - Mendapatkan ukuran terminal (width, height)

## Konstanta Warna
```go
const (
    ColorReset  = "\033[0m"
    ColorRed    = "\033[31m"
    ColorGreen  = "\033[32m"
    ColorYellow = "\033[33m"
    ColorBlue   = "\033[34m"
    ColorPurple = "\033[35m"
    ColorCyan   = "\033[36m"
    ColorWhite  = "\033[37m"
    ColorBold   = "\033[1m"
)
```

## Platform Support
- **Linux/macOS**: Menggunakan command `clear` dan ANSI escape sequences
- **Windows**: Menggunakan command `cls` dengan fallback ke ANSI sequences
- **Universal**: ANSI escape sequences sebagai fallback untuk semua platform

## Contoh Penggunaan dalam Command Delete

Berikut contoh bagaimana mengintegrasikan terminal utilities ke dalam command delete:

```go
// Di dalam fungsi deleteConfigWithSelection() 
func deleteConfigWithSelection() error {
    // Clear screen and show header
    terminal.ClearAndShowHeader("Delete Database Configuration")

    // Show loading spinner
    spinner := terminal.NewProgressSpinner("Scanning configuration files...")
    spinner.Start()
    
    // ... scan files ...
    
    spinner.Stop()

    // Display files in table format
    terminal.PrintSubHeader("Available Configuration Files")
    headers := []string{"#", "Filename", "Size"}
    rows := [][]string{
        {"1", "config_dev.cnf.enc", "2.1 KB"},
        {"2", "config_prod.cnf.enc", "2.3 KB"},
    }
    terminal.FormatTable(headers, rows)

    // Enhanced user interaction
    terminal.PrintInfo("Select files to delete (1,2 or 'all'):")
    
    // ... handle selection ...

    // Show progress during deletion
    bar := terminal.NewProgressBar(len(selectedFiles), "Deleting")
    for i, file := range selectedFiles {
        // ... delete file ...
        bar.Update(i + 1)
    }
    bar.Finish()

    terminal.PrintSuccess("Files deleted successfully!")
    terminal.WaitForEnter()
    return nil
}
```

## Integrasi dengan Command Lain

```go
// Clear screen before showing menu
terminal.ClearAndShowHeader("Database Configuration Manager")

// Show colored status messages
terminal.PrintSuccess("Configuration validated successfully")
terminal.PrintError("Failed to connect to database")
terminal.PrintWarning("Configuration file is outdated")

// Interactive confirmation
confirmed, err := terminal.ConfirmAndClear("Delete this configuration?")
if confirmed {
    // proceed with deletion
}
```

## File Structure
```
utils/terminal/
├── doc.go          # Package documentation
├── terminal.go     # Core terminal operations
├── formatting.go   # Text formatting and colors
├── helpers.go      # Helper functions and interactive utilities
└── examples.go     # Usage examples
```

## Dependencies
- `sfDBTools/internal/logger` - For logging operations
- Standard Go libraries: `os`, `exec`, `fmt`, `strings`, `time`

## Error Handling
Semua fungsi mengembalikan error yang dapat ditangani. Jika system command gagal, fungsi akan fallback ke ANSI escape sequences.

## Thread Safety
Progress spinner menggunakan goroutine dan channel untuk thread-safe operations. Fungsi lain tidak thread-safe dan harus digunakan dari single goroutine.
