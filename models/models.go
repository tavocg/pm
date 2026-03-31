// Package models
package models

import (
	"github.com/tavocg/pm/dispatcher"
)

type taskStr string

type Orchestrator struct {
	params *OrchestratorParams
}

type OrchestratorParams struct {
	Cfg        *OrchestratorConfig
	Dispatcher dispatcher.Dispatcher
}

type OrchestratorConfig struct {
	Managers []Manager `mapstructure:"managers"`
}

type Manager struct {
	Manager string   `mapstructure:"manager"`
	Depends []string `mapstructure:"depends"`
	Tasks   []Task   `mapstructure:"tasks"`
	Remotes []Remote `mapstructure:"remotes"`
}

type Remote struct {
	Remote   string    `mapstructure:"remote"`
	Packages []Package `mapstructure:"packages"`
}

type Package struct {
	Package string `mapstructure:"package"`
}

type Task struct {
	Task string `mapstructure:"task"`
	Cmd  string `mapstructure:"cmd"`
}
