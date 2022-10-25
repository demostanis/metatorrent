package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	. "github.com/demostanis/metatorrent/internal/model"
	"os"
)

func main() {
	p := tea.NewProgram(
		InitialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion())

	if err := p.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
