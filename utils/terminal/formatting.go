package terminal

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
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

// SpinnerStyle defines different spinner animation styles
type SpinnerStyle int

const (
	SpinnerDots SpinnerStyle = iota
	SpinnerArrows
	SpinnerClassic
	SpinnerBraille
	SpinnerPulse
	SpinnerBouncingBall
	SpinnerMinimal
)

// ProgressSpinner represents an enhanced spinner for showing progress
type ProgressSpinner struct {
	style    SpinnerStyle
	chars    []string
	colors   []string
	current  int
	message  string
	active   bool
	stopChan chan bool
	done     chan bool
	mu       sync.RWMutex
	interval time.Duration
	prefix   string
	suffix   string
}

// NewProgressSpinner creates a new progress spinner with default style
func NewProgressSpinner(message string) *ProgressSpinner {
	return NewProgressSpinnerWithStyle(message, SpinnerDots)
}

// NewProgressSpinnerWithStyle creates a new progress spinner with specified style
func NewProgressSpinnerWithStyle(message string, style SpinnerStyle) *ProgressSpinner {
	ps := &ProgressSpinner{
		style:    style,
		current:  0,
		message:  message,
		active:   false,
		stopChan: make(chan bool),
		done:     make(chan bool),
		interval: 100 * time.Millisecond,
		prefix:   "",
		suffix:   "",
	}

	ps.setStyle(style)
	return ps
}

// setStyle configures the spinner characters and colors based on style
func (ps *ProgressSpinner) setStyle(style SpinnerStyle) {
	switch style {
	case SpinnerDots:
		ps.chars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		ps.colors = []string{ColorCyan, ColorBlue, ColorPurple, ColorYellow}
		ps.interval = 80 * time.Millisecond
	case SpinnerArrows:
		ps.chars = []string{"→", "↘", "↓", "↙", "←", "↖", "↑", "↗"}
		ps.colors = []string{ColorGreen, ColorCyan}
		ps.interval = 120 * time.Millisecond
	case SpinnerClassic:
		ps.chars = []string{"|", "/", "-", "\\"}
		ps.colors = []string{ColorWhite}
		ps.interval = 100 * time.Millisecond
	case SpinnerBraille:
		ps.chars = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
		ps.colors = []string{ColorGreen, ColorYellow, ColorRed, ColorPurple}
		ps.interval = 90 * time.Millisecond
	case SpinnerPulse:
		ps.chars = []string{"●", "○", "●", "○"}
		ps.colors = []string{ColorRed, ColorYellow, ColorGreen, ColorBlue}
		ps.interval = 200 * time.Millisecond
	case SpinnerBouncingBall:
		ps.chars = []string{"( ●    )", "(  ●   )", "(   ●  )", "(    ● )", "(     ●)", "(    ● )", "(   ●  )", "(  ●   )", "( ●    )", "(●     )"}
		ps.colors = []string{ColorGreen}
		ps.interval = 100 * time.Millisecond
	case SpinnerMinimal:
		ps.chars = []string{".", "..", "...", "...."}
		ps.colors = []string{ColorCyan}
		ps.interval = 500 * time.Millisecond
	}
}

// SetPrefix sets a prefix to display before the spinner
func (ps *ProgressSpinner) SetPrefix(prefix string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.prefix = prefix
}

// SetSuffix sets a suffix to display after the message
func (ps *ProgressSpinner) SetSuffix(suffix string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.suffix = suffix
}

// SetInterval changes the animation speed
func (ps *ProgressSpinner) SetInterval(interval time.Duration) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.interval = interval
}

// Start begins the spinner animation
func (ps *ProgressSpinner) Start() {
	ps.mu.Lock()
	if ps.active {
		ps.mu.Unlock()
		return
	}
	ps.active = true
	ps.mu.Unlock()

	HideCursor()

	// Register as active spinner so print functions can coordinate
	spinnerMu.Lock()
	activeSpinner = ps
	spinnerMu.Unlock()

	go func() {
		ticker := time.NewTicker(ps.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ps.stopChan:
				ps.done <- true
				return
			case <-ticker.C:
				ps.render()
				ps.advance()
			}
		}
	}()
}

// StartWithNewline begins the spinner animation with a newline before it
func (ps *ProgressSpinner) StartWithNewline() {
	fmt.Println() // Add newline before starting spinner
	ps.Start()
}

// render draws the current spinner frame
func (ps *ProgressSpinner) render() {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if !ps.active {
		return
	}

	// Move cursor to beginning of line and clear it
	fmt.Print("\r\033[2K")

	// Get current color (cycle through colors if multiple)
	color := ColorWhite
	if len(ps.colors) > 0 {
		colorIndex := ps.current % len(ps.colors)
		color = ps.colors[colorIndex]
	}

	// Build the spinner line
	var line strings.Builder

	// Add prefix if set
	if ps.prefix != "" {
		line.WriteString(ps.prefix)
		line.WriteString(" ")
	}

	// Add colored spinner character
	if len(ps.chars) > 0 {
		spinnerChar := ps.chars[ps.current%len(ps.chars)]
		line.WriteString(color)
		line.WriteString(spinnerChar)
		line.WriteString(ColorReset)
		line.WriteString(" ")
	}

	// Add message
	line.WriteString(ps.message)

	// Add suffix if set
	if ps.suffix != "" {
		line.WriteString(" ")
		line.WriteString(ps.suffix)
	}

	fmt.Print(line.String())
}

// advance moves to the next animation frame
func (ps *ProgressSpinner) advance() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.current++
}

// Stop stops the spinner animation
func (ps *ProgressSpinner) Stop() {
	ps.mu.Lock()
	if !ps.active {
		ps.mu.Unlock()
		return
	}
	ps.active = false
	ps.mu.Unlock()

	// Signal stop
	ps.stopChan <- true

	// Wait for goroutine to finish
	<-ps.done

	// Clear the spinner line and add newline for clean output
	fmt.Print("\r\033[K")
	fmt.Println() // Add newline for cleaner output
	ShowCursor()

	// Unregister active spinner if it's this one
	spinnerMu.Lock()
	if activeSpinner == ps {
		activeSpinner = nil
	}
	spinnerMu.Unlock()
}

// StopWithMessage stops the spinner and displays a final message
func (ps *ProgressSpinner) StopWithMessage(message string) {
	ps.mu.Lock()
	if !ps.active {
		ps.mu.Unlock()
		return
	}
	ps.active = false
	ps.mu.Unlock()

	// Signal stop
	ps.stopChan <- true

	// Wait for goroutine to finish
	<-ps.done

	// Clear the spinner line and show final message with newline
	fmt.Print("\r\033[K")
	fmt.Println(message)
	ShowCursor()

	// Unregister active spinner if it's this one
	spinnerMu.Lock()
	if activeSpinner == ps {
		activeSpinner = nil
	}
	spinnerMu.Unlock()
}

// UpdateMessage updates the spinner message thread-safely
func (ps *ProgressSpinner) UpdateMessage(message string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.message = message
}

// NewLoadingSpinner creates a spinner optimized for loading operations
func NewLoadingSpinner(message string) *ProgressSpinner {
	return NewProgressSpinnerWithStyle(message, SpinnerDots)
}

// NewProcessingSpinner creates a spinner optimized for processing operations
func NewProcessingSpinner(message string) *ProgressSpinner {
	spinner := NewProgressSpinnerWithStyle(message, SpinnerBraille)
	spinner.SetPrefix("🔄")
	return spinner
}

// NewDownloadSpinner creates a spinner optimized for download operations
func NewDownloadSpinner(message string) *ProgressSpinner {
	spinner := NewProgressSpinnerWithStyle(message, SpinnerArrows)
	spinner.SetPrefix("⬇️")
	return spinner
}

// NewInstallSpinner creates a spinner optimized for installation operations
func NewInstallSpinner(message string) *ProgressSpinner {
	spinner := NewProgressSpinnerWithStyle(message, SpinnerPulse)
	spinner.SetPrefix("📦")
	return spinner
}

// NewRemoveSpinner creates a spinner optimized for removal operations
func NewRemoveSpinner(message string) *ProgressSpinner {
	spinner := NewProgressSpinnerWithStyle(message, SpinnerBraille)
	spinner.SetPrefix("🗑️")
	return spinner
}

// NewCleanupSpinner creates a spinner optimized for cleanup operations
func NewCleanupSpinner(message string) *ProgressSpinner {
	spinner := NewProgressSpinnerWithStyle(message, SpinnerClassic)
	spinner.SetPrefix("🧹")
	return spinner
}

// StopWithSuccess stops the spinner and shows a success message
func (ps *ProgressSpinner) StopWithSuccess(message string) {
	successMsg := ColorGreen + "✅ " + message + ColorReset
	ps.StopWithMessage(successMsg)
}

// StopWithError stops the spinner and shows an error message
func (ps *ProgressSpinner) StopWithError(message string) {
	errorMsg := ColorRed + "❌ " + message + ColorReset
	ps.StopWithMessage(errorMsg)
}

// StopWithWarning stops the spinner and shows a warning message
func (ps *ProgressSpinner) StopWithWarning(message string) {
	warningMsg := ColorYellow + "⚠️ " + message + ColorReset
	ps.StopWithMessage(warningMsg)
}

// temporaryStop temporarily stops the spinner for external output
func (ps *ProgressSpinner) temporaryStop() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.active {
		return
	}

	// Clear the current spinner line
	fmt.Print("\r\033[K")
}

// temporaryResume resumes the spinner after external output (does nothing, as spinner continues running)
func (ps *ProgressSpinner) temporaryResume() {
	// The spinner goroutine continues running, so we don't need to do anything
	// The next render() call will redraw the spinner
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
	// Pause active spinner to avoid overlapping output
	s := pauseActiveSpinner()
	fmt.Print(ColorText(text, color))
	resumeSpinner(s)
}

// PrintColoredLine prints a line with the specified color
func PrintColoredLine(text, color string) {
	s := pauseActiveSpinner()
	fmt.Println(ColorText(text, color))
	resumeSpinner(s)
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
	PrintColoredLine(message, ColorBlue)
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
