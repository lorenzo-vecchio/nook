package tui

import (
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

type Prompter interface {
	Select(label string, options []string, defaultOption string) (string, error)
	MultiSelect(label string, options []string, defaults []string) ([]string, error)
	Input(label string, defaultVal string) (string, error)
	Confirm(label string, defaultVal bool) (bool, error)
}

type SurveyPrompter struct{}

func NewPrompter() Prompter {
	return &SurveyPrompter{}
}

func (p *SurveyPrompter) Select(label string, options []string, defaultOption string) (string, error) {
	result := ""
	prompt := &survey.Select{
		Message: label,
		Options: options,
		Default: defaultOption,
		Filter: func(filter, value string, index int) bool {
			lowerFilter := strings.ToLower(filter)
			return strings.Contains(strings.ToLower(value), lowerFilter)
		},
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

func (p *SurveyPrompter) MultiSelect(label string, options []string, defaults []string) ([]string, error) {
	result := []string{}
	prompt := &survey.MultiSelect{
		Message: label,
		Options: options,
		Default: defaults,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

func (p *SurveyPrompter) Input(label string, defaultVal string) (string, error) {
	result := ""
	prompt := &survey.Input{
		Message: label,
		Default: defaultVal,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

func (p *SurveyPrompter) Confirm(label string, defaultVal bool) (bool, error) {
	result := false
	prompt := &survey.Confirm{
		Message: label,
		Default: defaultVal,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}
