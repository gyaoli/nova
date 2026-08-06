package logo

import (
	"fmt"
)

func Print()  {
	asciiArt := []string{
		" _ __     ___   __   __   __ _" ,
		"| '_ \\   / _ \\  \\ \\ / /  / _` |",
		"| | | | | (_) |  \\ V /  | (_| |",
		"|_| |_|  \\___/    \\_/    \\__,_|",
	}
	for _, line := range asciiArt {
		fmt.Println(line)
	}
}
