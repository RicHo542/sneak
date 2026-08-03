package ui

import "fmt"

const (
	ColorTeal  = "\033[38;2;70;255;200m"
	ColorDark  = "\033[38;2;26;38;50m"
	ColorWhite = "\033[1;37m"
	ColorGray  = "\033[38;2;126;140;155m"
	ColorReset = "\033[0m"
)

func PrintBanner() {
	textArt := ColorWhite + `
  ███████╗███╗   ██╗███████╗███████╗██╗  ██╗
  ██╔════╝████╗  ██║██╔════╝██╔══██╗██║ ██║
  ███████╗██╔██╗ ██║█████╗  ███████║██║██║
  ╚════██║██║╚██╗██║██╔══╝  ██╔══██║██║ ██║
  ███████║██║ ╚████║███████╗██║  ██║██║  ██║
  ╚══════╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝` + ColorTeal + `__` + ColorReset + `
`

	tagline := ColorGray + "   	 sneak: minimize red tape_" + ColorReset + "\n"

	fmt.Print(textArt)
	fmt.Println(tagline)
}

func Printfln(format string, a ...any) {
	fmt.Printf(format+"\n", a...)
}
