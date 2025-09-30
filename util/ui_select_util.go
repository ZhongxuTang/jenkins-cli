package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/lemonsoul/jenkins-cli/config"
	"github.com/manifoldco/promptui"
)

type QueueSelectItem struct {
	Name      string
	QueueInfo config.Queue
}

type BaseSelectItem struct {
	Name  string
	Value interface{}
}

func QueueUISelect(label string, items []QueueSelectItem) int {

	template := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U0001F34B {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "\u2714 {{ .Name | red | cyan }}",
		Details: `
--------- Pepper ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Queue ID:" | faint }}	{{ .QueueInfo.Id }}
{{ "TaskName:" | faint }}	{{ .QueueInfo.TaskName }}
{{ "Params:" | faint }}	{{ .QueueInfo.Params }}`,
	}
	selectPrompt := &promptui.Select{
		Label:     label,
		Items:     items,
		Templates: template,
		Size:      10,
	}
	index, _, err := selectPrompt.Run()
	if err != nil {
		color.Yellow("failed to select")
		return -1
	}
	return index
}

func StrUISelect(label string, itemStrs []string) string {
	if len(itemStrs) == 0 {
		//color.Red("No items to select")
		return ""
	}
	items := make([]BaseSelectItem, 0)
	for _, v := range itemStrs {
		items = append(items, BaseSelectItem{Name: v, Value: v})
	}

	template := &promptui.SelectTemplates{
		Label:    "{{ . }}? (Press Ctrl+C or 'q' to quit)",
		Active:   "\U0001F34B {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "\u2714 {{ .Name | cyan }}",
		Details: `
--------- Pepper ----------
{{ "Name:" | faint }}	{{ .Name }}`,
	}

	searcher := func(input string, index int) bool {
		item := items[index]
		name := item.Name

		// Handle 'q' to quit
		if input == "q" || input == "Q" {
			fmt.Println()
			color.Yellow("👋 Exiting...")
			os.Exit(0)
		}

		// Default search behavior - case insensitive substring match
		if input == "" {
			return true
		}
		return contains(name, input)
	}

	selectPrompt := &promptui.Select{
		Label:     label,
		Items:     items,
		Templates: template,
		Size:      10,
		Searcher:  searcher,
	}
	index, _, err := selectPrompt.Run()
	if err != nil {
		if err.Error() == "^C" || err.Error() == "interrupt" {
			fmt.Println()
			color.Yellow("👋 Exiting...")
			os.Exit(0)
		}
		color.Yellow("failed to select")
		return ""
	}

	return itemStrs[index]
}

// contains checks if a string contains a substring (case insensitive)
func contains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}
