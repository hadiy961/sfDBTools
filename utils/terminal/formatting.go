package terminal

import (
	"fmt"
	"strings"
	"time"

	"sfDBTools/internal/logger"
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
	lg, _ := logger.Get()

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

	lg.Debug("Progress spinner started", logger.String("message", ps.message))
}

// Stop stops the spinner animation
func (ps *ProgressSpinner) Stop() {
	lg, _ := logger.Get()

	if !ps.active {
		return
	}

	ps.active = false
	ps.stopChan <- true
	ClearCurrentLine()
	ShowCursor()

	lg.Debug("Progress spinner stopped")
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

// TruncateText truncates text to fit within the specified width
func TruncateText(text string, width int) string {
	if len(text) <= width {
		return text
	}
	if width <= 3 {
		return text[:width]
	}
	return text[:width-3] + "..."
}

// FormatTable formats data as a table
func FormatTable(headers []string, rows [][]string) {
	if len(headers) == 0 || len(rows) == 0 {
		return
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print headers
	fmt.Print("|")
	for i, header := range headers {
		fmt.Printf(" %s |", PadRight(header, colWidths[i]))
	}
	fmt.Println()

	// Print separator
	fmt.Print("|")
	for _, width := range colWidths {
		fmt.Printf("%s|", strings.Repeat("-", width+2))
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		fmt.Print("|")
		for i, cell := range row {
			if i < len(colWidths) {
				fmt.Printf(" %s |", PadRight(cell, colWidths[i]))
			}
		}
		fmt.Println()
	}
}
