package models

const (
	syncRemotesTask   taskStr = "sync-remotes"
	updateRemotesTask taskStr = "update-remotes"
)

func (o *Orchestrator) runTask(task taskStr) error {
	for _, m := range o.params.Cfg.Managers {
		for _, t := range m.Tasks {
			if t.Cmd == "" || t.Task != string(task) {
				continue
			}

			cmd := o.params.Dispatcher.Cmd(t.Cmd)
			if t.Privileged {
				cmd = cmd.WithPrivileged()
			}

			if t.Interactive {
				cmd = cmd.WithInteractive()
			}

			if err := o.params.Dispatcher.Run(cmd); err != nil {
				return err
			}
		}
	}

	return nil
}
