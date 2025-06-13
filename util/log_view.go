package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

func main() {
	file, err := os.Open("jenkins.log") // 模拟本地日志文件
	if err != nil {
		fmt.Println("无法打开日志文件:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	fmt.Println("====== Jenkins Log View ======")

	for scanner.Scan() {
		line := scanner.Text()
		printLogLine(line)
		time.Sleep(100 * time.Millisecond) // 模拟逐行输出
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("读取日志出错:", err)
	}
}

func printLogLine(line string) {
	switch {
	case strings.Contains(line, "INFO"):
		color.Set(color.FgGreen)
	case strings.Contains(line, "WARN"):
		color.Set(color.FgYellow)
	case strings.Contains(line, "ERROR"):
		color.Set(color.FgRed)
	default:
		color.Set(color.FgWhite)
	}
	fmt.Println(line)
	color.Unset()
}
