package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"sfDBTools/internal/logger"
)

// ClearScreen clears the terminal screen using platform-specific commands
func ClearScreen() error {
	lg, _ := logger.Get()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		// Linux, macOS, and other Unix-like systems
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		lg.Warn("Failed to clear screen using system command, falling back to ANSI escape sequences",
			logger.String("os", runtime.GOOS),
			logger.Error(err))
		return ClearScreenANSI()
	}

	lg.Debug("Terminal screen cleared", logger.String("method", "system_command"))
	return nil
}

// ClearScreenANSI clears the terminal screen using ANSI escape sequences
func ClearScreenANSI() error {
	lg, _ := logger.Get()

	// ANSI escape sequence to clear screen and move cursor to top-left
	_, err := fmt.Print("\033[2J\033[H")
	if err != nil {
		lg.Error("Failed to clear screen using ANSI escape sequences", logger.Error(err))
		return err
	}

	lg.Debug("Terminal screen cleared", logger.String("method", "ansi_escape"))
	return nil
}

// ClearLines clears the specified number of lines from the current cursor position
func ClearLines(lines int) error {
	lg, _ := logger.Get()

	if lines <= 0 {
		return nil
	}

	// Move cursor up and clear each line
	for i := 0; i < lines; i++ {
		// Move cursor up one line and clear the line
		fmt.Print("\033[1A\033[2K")
	}

	lg.Debug("Cleared lines from terminal", logger.Int("lines", lines))
	return nil
}

// ClearCurrentLine clears the current line and moves cursor to the beginning
func ClearCurrentLine() error {
	lg, _ := logger.Get()

	// Clear current line and move cursor to beginning
	_, err := fmt.Print("\033[2K\r")
	if err != nil {
		lg.Error("Failed to clear current line", logger.Error(err))
		return err
	}

	lg.Debug("Current line cleared")
	return nil
}

// ClearToEndOfLine clears from cursor position to end of line
func ClearToEndOfLine() error {
	lg, _ := logger.Get()

	// Clear from cursor to end of line
	_, err := fmt.Print("\033[K")
	if err != nil {
		lg.Error("Failed to clear to end of line", logger.Error(err))
		return err
	}

	lg.Debug("Cleared to end of line")
	return nil
}

// MoveCursor moves the cursor to the specified row and column (1-indexed)
func MoveCursor(row, col int) error {
	lg, _ := logger.Get()

	if row < 1 || col < 1 {
		return fmt.Errorf("cursor position must be >= 1, got row=%d, col=%d", row, col)
	}

	_, err := fmt.Printf("\033[%d;%dH", row, col)
	if err != nil {
		lg.Error("Failed to move cursor",
			logger.Int("row", row),
			logger.Int("col", col),
			logger.Error(err))
		return err
	}

	lg.Debug("Cursor moved", logger.Int("row", row), logger.Int("col", col))
	return nil
}

// MoveCursorHome moves the cursor to the top-left corner (1,1)
func MoveCursorHome() error {
	return MoveCursor(1, 1)
}

// SaveCursorPosition saves the current cursor position
func SaveCursorPosition() error {
	lg, _ := logger.Get()

	_, err := fmt.Print("\033[s")
	if err != nil {
		lg.Error("Failed to save cursor position", logger.Error(err))
		return err
	}

	lg.Debug("Cursor position saved")
	return nil
}

// RestoreCursorPosition restores the previously saved cursor position
func RestoreCursorPosition() error {
	lg, _ := logger.Get()

	_, err := fmt.Print("\033[u")
	if err != nil {
		lg.Error("Failed to restore cursor position", logger.Error(err))
		return err
	}

	lg.Debug("Cursor position restored")
	return nil
}

// HideCursor hides the terminal cursor
func HideCursor() error {
	lg, _ := logger.Get()

	_, err := fmt.Print("\033[?25l")
	if err != nil {
		lg.Error("Failed to hide cursor", logger.Error(err))
		return err
	}

	lg.Debug("Cursor hidden")
	return nil
}

// ShowCursor shows the terminal cursor
func ShowCursor() error {
	lg, _ := logger.Get()

	_, err := fmt.Print("\033[?25h")
	if err != nil {
		lg.Error("Failed to show cursor", logger.Error(err))
		return err
	}

	lg.Debug("Cursor shown")
	return nil
}

// GetTerminalSize returns the terminal width and height
func GetTerminalSize() (width, height int, err error) {
	lg, _ := logger.Get()

	// First try using environment variables (more reliable)
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if lines := os.Getenv("LINES"); lines != "" {
			var w, h int
			if _, err := fmt.Sscanf(cols, "%d", &w); err == nil {
				if _, err := fmt.Sscanf(lines, "%d", &h); err == nil {
					lg.Debug("Terminal size from environment", logger.Int("width", w), logger.Int("height", h))
					return w, h, nil
				}
			}
		}
	}

	// Try using tput command (more reliable than stty)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// On Windows, try powershell
		cmd = exec.Command("powershell", "-Command", "(Get-Host).UI.RawUI.WindowSize")
	default:
		// Try tput first (more reliable)
		if width, height, err := getTputSize(); err == nil {
			return width, height, nil
		}

		// Fallback to stty
		cmd = exec.Command("stty", "size")
	}

	if cmd == nil {
		lg.Debug("No terminal size command available, using defaults")
		return 80, 24, nil
	}

	output, err := cmd.Output()
	if err != nil {
		lg.Debug("Failed to get terminal size via command, using defaults", logger.Error(err))
		// Return default values without error
		return 80, 24, nil
	}

	outputStr := strings.TrimSpace(string(output))

	if runtime.GOOS == "windows" {
		// Parse Windows PowerShell output (simplified)
		lg.Debug("Windows terminal size detection using defaults")
		return 80, 24, nil
	} else {
		// Parse Unix stty output: "height width"
		var h, w int
		if _, err := fmt.Sscanf(outputStr, "%d %d", &h, &w); err != nil {
			lg.Debug("Failed to parse terminal size, using defaults", logger.Error(err))
			return 80, 24, nil
		}
		lg.Debug("Terminal size detected via stty", logger.Int("width", w), logger.Int("height", h))
		return w, h, nil
	}
}

// getTputSize tries to get terminal size using tput command
func getTputSize() (width, height int, err error) {
	lg, _ := logger.Get()

	// Get columns
	colsCmd := exec.Command("tput", "cols")
	colsOutput, err := colsCmd.Output()
	if err != nil {
		return 0, 0, err
	}

	// Get lines
	linesCmd := exec.Command("tput", "lines")
	linesOutput, err := linesCmd.Output()
	if err != nil {
		return 0, 0, err
	}

	var w, h int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(colsOutput)), "%d", &w); err != nil {
		return 0, 0, err
	}

	if _, err := fmt.Sscanf(strings.TrimSpace(string(linesOutput)), "%d", &h); err != nil {
		return 0, 0, err
	}

	lg.Debug("Terminal size detected via tput", logger.Int("width", w), logger.Int("height", h))
	return w, h, nil
}

// PrintBorder prints a horizontal border across the terminal width
func PrintBorder(char rune, width int) {
	if width <= 0 {
		width, _, _ = GetTerminalSize()
		if width <= 0 {
			width = 80 // fallback
		}
	}

	border := strings.Repeat(string(char), width)
	fmt.Println(border)
}

// PrintSeparator prints a separator line
func PrintSeparator() {
	PrintBorder('=', 0)
}

// PrintDashedSeparator prints a dashed separator line
func PrintDashedSeparator() {
	PrintBorder('-', 0)
}

// WaitForEnter waits for the user to press Enter
func WaitForEnter() {
	fmt.Print("Press Enter to continue...")
	fmt.Scanln()
}

// WaitForEnterWithMessage waits for the user to press Enter with a custom message
func WaitForEnterWithMessage(message string) {
	fmt.Print(message)
	fmt.Scanln()
}
