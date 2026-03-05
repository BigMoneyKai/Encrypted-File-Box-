package logo

import (
    "fmt"
	"time"
)

func DrawLogo() error {
    logo := []string{
		"███████╗███████╗██████╗ ",
		"██╔════╝██╔════╝██╔══██╗",
		"███████╗█████╗  ██████╔╝",
		"╚════██║██╔══╝  ██╔══██╗",
		"███████║██║     ██████╔╝",
		"╚══════╝╚═╝     ╚═════╝ ",
	}
	for frame := 0; frame < 10; frame++ {
		fmt.Print("\033[H\033[2J") // 清屏
		for _, line := range logo {
			for i, ch := range line {
				// 🌈 渐变算法（核心）
				r := (i*2 + frame*8) % 256
				g := (150 + i*3 + frame*5) % 256
				b := (255 - i*4 + frame*6) % 256

				fmt.Printf("\033[38;2;%d;%d;%dm%c", r, g, b, ch)
			}
			fmt.Print("\033[0m\n")
		}

		time.Sleep(120 * time.Millisecond)
	}
	fmt.Println();

	return nil;
}

