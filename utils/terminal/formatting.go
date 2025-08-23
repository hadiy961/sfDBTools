package terminal

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"sfDBTools/internal/logger"

	"github.com/olekukonko/tablewriter"
)

// Colors for terminal output
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

// ProgressSpinner represents a simple spinner for showing progress
type ProgressSpinner struct {
	chars    []string
	current  int
	message  string
	active   bool
	stopChan chan bool
}

// NewProgressSpinner creates a new progress spinner
func NewProgressSpinner(message string) *ProgressSpinner {
	return &ProgressSpinner{
		chars:    []string{"|", "/", "-", "\\"},
		current:  0,
		message:  message,
		active:   false,
		stopChan: make(chan bool),
	}
}

// Start begins the spinner animation
func (ps *ProgressSpinner) Start() {
	if ps.active {
		return
	}

	ps.active = true
	HideCursor()

	go func() {
		for {
			select {
			case <-ps.stopChan:
				return
			default:
				ClearCurrentLine()
				fmt.Printf("%s %s", ps.chars[ps.current], ps.message)
				ps.current = (ps.current + 1) % len(ps.chars)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Start spinner without debug logging for cleaner output
}

// Stop stops the spinner animation
func (ps *ProgressSpinner) Stop() {
	if !ps.active {
		return
	}

	ps.active = false
	ps.stopChan <- true
	ClearCurrentLine()
	ShowCursor()

	// Stop spinner without debug logging for cleaner output
}

// UpdateMessage updates the spinner message
func (ps *ProgressSpinner) UpdateMessage(message string) {
	ps.message = message
}

// ProgressBar represents a progress bar
type ProgressBar struct {
	total   int
	current int
	width   int
	message string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int, message string) *ProgressBar {
	width, _, _ := GetTerminalSize()
	if width <= 0 {
		width = 80
	}
	// Reserve space for message and percentage
	barWidth := width - len(message) - 20
	if barWidth < 10 {
		barWidth = 10
	}

	return &ProgressBar{
		total:   total,
		current: 0,
		width:   barWidth,
		message: message,
	}
}

// Update updates the progress bar
func (pb *ProgressBar) Update(current int) {
	lg, _ := logger.Get()

	pb.current = current
	if pb.current > pb.total {
		pb.current = pb.total
	}

	percentage := float64(pb.current) / float64(pb.total) * 100
	filled := int(float64(pb.width) * float64(pb.current) / float64(pb.total))

	ClearCurrentLine()

	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.width-filled)
	fmt.Printf("%s [%s] %.1f%% (%d/%d)", pb.message, bar, percentage, pb.current, pb.total)

	lg.Debug("Progress bar updated",
		logger.Int("current", pb.current),
		logger.Int("total", pb.total),
		logger.Float64("percentage", percentage))
}

// Finish completes the progress bar
func (pb *ProgressBar) Finish() {
	pb.Update(pb.total)
	fmt.Println() // Move to next line
}

// ColorText applies color to text
func ColorText(text, color string) string {
	return color + text + ColorReset
}

// PrintColoredText prints text with the specified color
func PrintColoredText(text, color string) {
	fmt.Print(ColorText(text, color))
}

// PrintColoredLine prints a line with the specified color
func PrintColoredLine(text, color string) {
	fmt.Println(ColorText(text, color))
}

// PrintSuccess prints success message in green
func PrintSuccess(message string) {
	PrintColoredLine("✅ "+message, ColorGreen)
}

// PrintError prints error message in red
func PrintError(message string) {
	PrintColoredLine("❌ "+message, ColorRed)
}

// PrintWarning prints warning message in yellow
func PrintWarning(message string) {
	PrintColoredLine("⚠️ "+message, ColorYellow)
}

// PrintInfo prints info message in blue
func PrintInfo(message string) {
	PrintColoredLine("ℹ️ "+message, ColorBlue)
}

// PrintHeader prints a header with border
func PrintHeader(title string) {
	width, _, _ := GetTerminalSize()
	if width <= 0 {
		width = 80
	}

	// Calculate padding
	titleLen := len(title)
	if titleLen+4 > width {
		width = titleLen + 4
	}

	border := strings.Repeat("=", width)
	padding := (width - titleLen - 2) / 2
	leftPad := strings.Repeat(" ", padding)
	rightPad := strings.Repeat(" ", width-titleLen-2-padding)

	fmt.Println()
	PrintColoredLine(border, ColorCyan)
	PrintColoredLine("|"+leftPad+title+rightPad+"|", ColorCyan)
	PrintColoredLine(border, ColorCyan)
	fmt.Println()
}

// PrintSubHeader prints a sub-header
func PrintSubHeader(title string) {
	fmt.Println()
	PrintColoredLine("📋 "+title, ColorBold)
	PrintDashedSeparator()
}

// CenterText centers text within the specified width
func CenterText(text string, width int) string {
	textLen := len(text)
	if textLen >= width {
		return text
	}

	padding := (width - textLen) / 2
	leftPad := strings.Repeat(" ", padding)
	rightPad := strings.Repeat(" ", width-textLen-padding)
	return leftPad + text + rightPad
}

// PadLeft pads text to the left
func PadLeft(text string, width int) string {
	textLen := len(text)
	if textLen >= width {
		return text
	}
	return strings.Repeat(" ", width-textLen) + text
}

// PadRight pads text to the right
func PadRight(text string, width int) string {
	textLen := len(text)
	if textLen >= width {
		return text
	}
	return text + strings.Repeat(" ", width-textLen)
}

// TruncateText truncates text to fit within the specified width considering display width
func TruncateText(text string, width int) string {
	displayWidth := GetDisplayWidth(text)
	if displayWidth <= width {
		return text
	}
	if width <= 3 {
		// Remove ANSI codes and truncate
		ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
		cleanText := ansiRegex.ReplaceAllString(text, "")
		if len(cleanText) <= width {
			return cleanText
		}
		return cleanText[:width]
	}

	// For text with ANSI codes, we need to be more careful
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	cleanText := ansiRegex.ReplaceAllString(text, "")

	if len(cleanText) <= width-3 {
		return text // Original text fits
	}

	// Truncate clean text and add ellipsis
	truncated := cleanText[:width-3] + "..."
	return truncated
}

// GetDisplayWidth calculates the actual display width of text, ignoring ANSI escape sequences
func GetDisplayWidth(text string) int {
	// Remove ANSI escape sequences to get actual display width
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	cleanText := ansiRegex.ReplaceAllString(text, "")
	return len(cleanText)
}

// PadRightWithDisplay pads text to the right considering ANSI color codes
func PadRightWithDisplay(text string, width int) string {
	displayWidth := GetDisplayWidth(text)
	if displayWidth >= width {
		return text
	}
	padding := width - displayWidth
	return text + strings.Repeat(" ", padding)
}

// FormatTable formats data as a table using tablewriter library for better appearance
func FormatTable(headers []string, rows [][]string) {
	if len(headers) == 0 || len(rows) == 0 {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)

	// Set table headers using the correct method
	headerInterface := make([]interface{}, len(headers))
	for i, v := range headers {
		headerInterface[i] = v
	}
	table.Header(headerInterface...)

	// Add all rows
	for _, row := range rows {
		// Convert row to interface slice
		rowInterface := make([]interface{}, len(row))
		for i, v := range row {
			rowInterface[i] = v
		}
		table.Append(rowInterface...)
	}

	// Render the table
	table.Render()
}
