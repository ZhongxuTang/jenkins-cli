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
	Name     string
	Value    interface{}
	IsRecent bool
}

const recentPrefix = "[Recent] "

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
	return strUISelect(label, itemStrs)
}

func StrUISelectWithRecent(label string, itemStrs []string, recent []string) string {
	items, mapping := buildRecentItems(itemStrs, recent)
	return strUISelectItems(label, items, mapping)
}

func strUISelect(label string, itemStrs []string) string {
	if len(itemStrs) == 0 {
		//color.Red("No items to select")
		return ""
	}
	items := make([]BaseSelectItem, 0)
	for _, v := range itemStrs {
		items = append(items, BaseSelectItem{Name: v, Value: v})
	}
	return strUISelectItems(label, items, nil)
}

func strUISelectItems(label string, items []BaseSelectItem, mapping map[int]string) string {
	if len(items) == 0 {
		return ""
	}

	template := &promptui.SelectTemplates{
		Label:    "{{ . }}? (Press Ctrl+C or 'q' to quit)",
		Active:   "{{ if .IsRecent }}\U0001F34B {{ .Name | yellow }}{{ else }}\U0001F34B {{ .Name | cyan }}{{ end }}",
		Inactive: "{{ if .IsRecent }}  {{ .Name | yellow }}{{ else }}  {{ .Name | cyan }}{{ end }}",
		Selected: "{{ if .IsRecent }}\u2714 {{ .Name | yellow }}{{ else }}\u2714 {{ .Name | cyan }}{{ end }}",
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

	if mapping != nil {
		if value, ok := mapping[index]; ok {
			return value
		}
	}
	if items[index].Value == nil {
		return ""
	}
	if value, ok := items[index].Value.(string); ok {
		return value
	}
	return ""
}

// contains checks if a string contains a substring (case insensitive)
func contains(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

func buildRecentItems(items []string, recent []string) ([]BaseSelectItem, map[int]string) {
	if len(items) == 0 && len(recent) == 0 {
		return []BaseSelectItem{}, nil
	}
	recentSet := make(map[string]struct{}, len(recent))
	merged := make([]BaseSelectItem, 0, len(items)+len(recent))
	indexMapping := make(map[int]string)
	for _, value := range recent {
		if value == "" {
			continue
		}
		if _, ok := recentSet[value]; ok {
			continue
		}
		recentSet[value] = struct{}{}
		indexMapping[len(merged)] = value
		merged = append(merged, BaseSelectItem{
			Name:     recentPrefix + value,
			Value:    value,
			IsRecent: true,
		})
	}
	for _, value := range items {
		if value == "" {
			continue
		}
		merged = append(merged, BaseSelectItem{Name: value, Value: value})
	}
	if len(merged) == 0 {
		return []BaseSelectItem{}, nil
	}
	return merged, indexMapping
}
