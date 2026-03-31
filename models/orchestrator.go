// Package models
package models

import (
	"errors"
	"os/exec"
)

func (o *Orchestrator) UpdateRemotes() error {
	if err := o.SyncRemotes(); err != nil {
		return err
	}
	return o.runTask(updateRemotesTask)
}

func (o *Orchestrator) SyncRemotes() error {
	return o.runTask(syncRemotesTask)
}

func NewOrchestrator(params *OrchestratorParams) (*Orchestrator, error) {
	if params == nil {
		return nil, errors.New("nil params")
	}

	if params.Cfg == nil {
		return nil, errors.New("nil config")
	}

	if params.Dispatcher == nil {
		return nil, errors.New("nil dispatcher")
	}

	return &Orchestrator{params}, nil
}

func (o *Orchestrator) RemoveUnsupportedManagers() {
	for i, m := range o.params.Cfg.Managers {
		for _, file := range m.Depends {
			if _, err := exec.LookPath(file); err != nil {
				o.params.Cfg.Managers = append(o.params.Cfg.Managers[:i], o.params.Cfg.Managers[i+1:]...)
			}
		}
	}
}
