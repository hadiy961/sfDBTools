package terminal

import (
	"fmt"
	"time"
)

// ExampleUsage demonstrates how to use the terminal utilities
func ExampleUsage() {
	// Example 1: Basic screen clearing
	fmt.Println("Current content on screen...")
	time.Sleep(2 * time.Second)
	Clear()
	fmt.Println("Screen cleared!")

	// Example 2: Clear with header
	ClearAndShowHeader("Database Configuration Manager")

	// Example 3: Progress spinner
	spinner := NewProgressSpinner("Loading configuration files...")
	spinner.Start()
	time.Sleep(3 * time.Second)
	spinner.UpdateMessage("Validating files...")
	time.Sleep(2 * time.Second)
	spinner.Stop()
	PrintSuccess("Files loaded successfully")

	// Example 4: Progress bar
	bar := NewProgressBar(100, "Processing")
	for i := 0; i <= 100; i += 5 {
		bar.Update(i)
		time.Sleep(100 * time.Millisecond)
	}
	bar.Finish()

	// Example 5: Colored output
	PrintSuccess("Operation completed successfully")
	PrintError("An error occurred during processing")
	PrintWarning("This is a warning message")
	PrintInfo("Information: Process running normally")

	// Example 6: Interactive menu (commented out for demo)
	/*
		menu := &InteractiveMenu{
			Title: "Database Configuration Menu",
			Options: []string{
				"Generate new config",
				"Edit existing config",
				"Delete config",
				"Show config",
			},
			OnSelect: func(choice int) error {
				switch choice {
				case 1:
					PrintInfo("Generating new configuration...")
				case 2:
					PrintInfo("Editing configuration...")
				case 3:
					PrintInfo("Deleting configuration...")
				case 4:
					PrintInfo("Showing configuration...")
				}
				PauseAndClearWithMessage("Press Enter to continue...")
				return nil
			},
			OnExit: func() error {
				PrintInfo("Goodbye!")
				return nil
			},
		}

		// Uncomment to run interactive menu
		// menu.Show()
	*/
}

// ClearScreenDemo demonstrates different clearing methods
func ClearScreenDemo() {
	fmt.Println("=== Clear Screen Demo ===")
	fmt.Println("This is some content on the screen")
	fmt.Println("We will clear it in 3 seconds...")

	time.Sleep(3 * time.Second)

	// Method 1: Simple clear
	Clear()

	fmt.Println("Screen cleared using Clear()")
	time.Sleep(2 * time.Second)

	// Method 2: Clear with message
	ClearWithMessage("Screen cleared with a message!")
	time.Sleep(2 * time.Second)

	// Method 3: Clear and show header
	ClearAndShowHeader("Application Header")
	fmt.Println("Content below the header...")
}

// ProgressDemo demonstrates progress indicators
func ProgressDemo() {
	ClearAndShowHeader("Progress Indicators Demo")

	// Spinner demo
	PrintInfo("Demonstrating spinner...")
	spinner := NewProgressSpinner("Loading data...")
	spinner.Start()
	time.Sleep(3 * time.Second)
	spinner.UpdateMessage("Processing data...")
	time.Sleep(2 * time.Second)
	spinner.Stop()
	PrintSuccess("Spinner demo completed")

	time.Sleep(1 * time.Second)

	// Progress bar demo
	PrintInfo("Demonstrating progress bar...")
	bar := NewProgressBar(50, "Download")
	for i := 0; i <= 50; i++ {
		bar.Update(i)
		time.Sleep(50 * time.Millisecond)
	}
	bar.Finish()
	PrintSuccess("Progress bar demo completed")
}

// TableDemo demonstrates table formatting
func TableDemo() {
	ClearAndShowHeader("Table Formatting Demo")

	headers := []string{"ID", "Name", "Status", "Created"}
	rows := [][]string{
		{"1", "config_dev.cnf.enc", "Active", "2025-08-23"},
		{"2", "config_prod.cnf.enc", "Active", "2025-08-22"},
		{"3", "config_test.cnf.enc", "Inactive", "2025-08-21"},
	}

	PrintSubHeader("Database Configuration Files")
	FormatTable(headers, rows)

	fmt.Println()
	PrintInfo("Table formatting demo completed")
}
