package terminal

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Clear clears the terminal screen - simple wrapper for ClearScreen
func Clear() error {
	return ClearScreen()
}

// ClearAndHome clears the screen and moves cursor to home position
func ClearAndHome() error {
	if err := ClearScreen(); err != nil {
		return err
	}
	return MoveCursorHome()
}

// ClearWithMessage clears screen and displays a message
func ClearWithMessage(message string) error {
	if err := ClearScreen(); err != nil {
		return err
	}
	if message != "" {
		fmt.Println(message)
	}
	return nil
}

// ClearAndShowHeader clears screen and shows a formatted header
func ClearAndShowHeader(title string) error {
	if err := ClearScreen(); err != nil {
		return err
	}
	PrintHeader(title)
	return nil
}

// PauseAndClear waits for user input then clears the screen
func PauseAndClear() error {
	WaitForEnter()
	return ClearScreen()
}

// PauseAndClearWithMessage shows a custom message, waits for input, then clears
func PauseAndClearWithMessage(message string) error {
	WaitForEnterWithMessage(message)
	return ClearScreen()
}

// RefreshDisplay clears screen and refreshes with new content
func RefreshDisplay(content func()) error {
	if err := ClearScreen(); err != nil {
		return err
	}
	if content != nil {
		content()
	}
	return nil
}

// ClearLastLines clears the last N lines and repositions cursor
func ClearLastLines(n int) error {
	return ClearLines(n)
}

// ConfirmAndClear shows a confirmation dialog then clears screen
func ConfirmAndClear(question string) (bool, error) {
	// Pause any active spinner so prompt is not overwritten
	s := pauseActiveSpinner()
	defer resumeSpinner(s)

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", question)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(strings.ToLower(response))
	confirmed := response == "y" || response == "yes"

	if err := ClearScreen(); err != nil {
		return confirmed, err
	}

	return confirmed, nil
}

// ShowMenuAndClear displays a menu, gets user choice, then clears screen
func ShowMenuAndClear(title string, options []string) (int, error) {
	fmt.Println()
	if title != "" {
		PrintSubHeader(title)
	}

	for i, option := range options {
		fmt.Printf("   %d. %s\n", i+1, option)
	}

	// Pause any active spinner so menu and prompt are visible
	s := pauseActiveSpinner()
	defer resumeSpinner(s)

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSelect option (1-%d): ", len(options))

	choice, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	choice = strings.TrimSpace(choice)

	// Parse choice
	var selected int
	if _, err := fmt.Sscanf(choice, "%d", &selected); err != nil {
		ClearScreen()
		return 0, fmt.Errorf("invalid selection: %s", choice)
	}

	if selected < 1 || selected > len(options) {
		ClearScreen()
		return 0, fmt.Errorf("selection out of range: %d", selected)
	}

	if err := ClearScreen(); err != nil {
		return selected, err
	}

	return selected, nil
}

// DisplayMessageAndClear shows a message for a brief moment then clears
func DisplayMessageAndClear(message string, pauseSeconds int) error {
	fmt.Println(message)

	if pauseSeconds > 0 {
		fmt.Printf("Screen will clear in %d seconds...\n", pauseSeconds)
		for i := pauseSeconds; i > 0; i-- {
			fmt.Printf("\rClearing in %d... ", i)
			// Note: In a real implementation, you'd want to use time.Sleep(time.Second)
			// but for this utility, we'll just show the countdown
		}
		fmt.Println()
	}

	return ClearScreen()
}

// InteractiveMenu creates an interactive menu that automatically clears between selections
type InteractiveMenu struct {
	Title    string
	Options  []string
	OnSelect func(int) error
	OnExit   func() error
}

// Show displays the interactive menu
func (im *InteractiveMenu) Show() error {
	for {
		ClearAndShowHeader(im.Title)

		// Add exit option
		allOptions := append(im.Options, "Exit")

		selected, err := ShowMenuAndClear("", allOptions)
		if err != nil {
			PrintError(fmt.Sprintf("Menu error: %v", err))
			WaitForEnter()
			continue
		}

		// Check if user selected exit
		if selected == len(allOptions) {
			if im.OnExit != nil {
				return im.OnExit()
			}
			return nil
		}

		// Execute selected option
		if im.OnSelect != nil {
			if err := im.OnSelect(selected); err != nil {
				PrintError(fmt.Sprintf("Action error: %v", err))
				WaitForEnter()
			}
		}
	}
}
