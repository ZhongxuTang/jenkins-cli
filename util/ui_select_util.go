package util

import (
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
		Label:    "{{ . }}?",
		Active:   "\U0001F34B {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "\u2714 {{ .Name | cyan }}",
		Details: `
--------- Pepper ----------
{{ "Name:" | faint }}	{{ .Name }}`,
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
		return ""
	}
	return itemStrs[index]
}
