package models

const (
	syncRemotesTask taskStr = "sync-remotes"
)

func (o *Orchestrator) runTask(task taskStr) error {
	for _, m := range o.params.Cfg.Managers {
		for _, t := range m.Tasks {
			if t.Cmd == "" || t.Task != string(task) {
				continue
			}

			if t.Privileged {
				if err := o.params.Dispatcher.RunAsPrivileged(t.Cmd); err != nil {
					return err
				}
			} else {
				if err := o.params.Dispatcher.Run(t.Cmd); err != nil {
					return err
				}
			}

		}
	}

	return nil
}
