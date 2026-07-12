package core

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/term"
)

func runSelector(cmds []Command) (*Command, error) {
	if len(cmds) == 0 {
		return nil, nil
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fallbackSelector(cmds)
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	sel := 0
	totalLines := len(cmds) * 2 + 2

	moveUp := func() { fmt.Printf("\033[%dA\033[J", totalLines) }

	render := func() {
		fmt.Print("\r\n")
		for i, c := range cmds {
			marker := " "
			if i == sel {
				marker = ">"
			}
			fmt.Printf("\r%s [%d] (%s) %s\n", marker, i+1, c.Context, c.Title)
			fmt.Printf("\r  %s\n", c.Command)
		}
		fmt.Print("\r\n")
	}

	clear := func() { moveUp(); fmt.Print("\033[J") }

	fmt.Print("\033[J")
	render()

	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return nil, err
		}

		switch {
		case n == 1 && buf[0] == '\r':
			clear()
			return &cmds[sel], nil

		case n == 1 && (buf[0] == 'q' || buf[0] == '\x1b' || buf[0] == 3):
			clear()
			return nil, nil

		case n == 3 && buf[0] == '\x1b' && buf[1] == '[':
			old := sel
			switch buf[2] {
			case 'A':
				if sel > 0 {
					sel--
				}
			case 'B':
				if sel < len(cmds)-1 {
					sel++
				}
			}
			if sel != old {
				clear()
				render()
			}
		}
	}
}

func fallbackSelector(cmds []Command) (*Command, error) {
	fmt.Println()
	for i, c := range cmds {
		fmt.Printf("[%d] (%s) %s\n     %s\n", i+1, c.Context, c.Title, c.Command)
	}
	fmt.Println()

	choice := promptLine("Copy which? (number, ENTER cancels): ")
	if choice == "" {
		return nil, nil
	}
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(cmds) {
		fmt.Println("Invalid choice.")
		return nil, nil
	}
	return &cmds[idx-1], nil
}
