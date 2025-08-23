// Package terminal provides utilities for terminal operations including
// screen clearing, cursor manipulation, progress indicators, and text formatting.
//
// Example usage:
//
//	// Clear the terminal screen
//	terminal.ClearScreen()
//
//	// Show a progress spinner
//	spinner := terminal.NewProgressSpinner("Processing...")
//	spinner.Start()
//	// Do some work...
//	spinner.Stop()
//
//	// Show a progress bar
//	bar := terminal.NewProgressBar(100, "Downloading")
//	for i := 0; i <= 100; i++ {
//		bar.Update(i)
//		time.Sleep(50 * time.Millisecond)
//	}
//	bar.Finish()
//
//	// Print colored messages
//	terminal.PrintSuccess("Operation completed successfully")
//	terminal.PrintError("An error occurred")
//	terminal.PrintWarning("This is a warning")
//	terminal.PrintInfo("Information message")
//
//	// Format and display tables
//	headers := []string{"Name", "Age", "City"}
//	rows := [][]string{
//		{"John", "25", "New York"},
//		{"Jane", "30", "Los Angeles"},
//	}
//	terminal.FormatTable(headers, rows)
package terminal
