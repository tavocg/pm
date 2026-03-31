/*
Copyright © 2026 Gustavo Calvo <tavo@tavo.cr>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"context"
	"os"
	"path"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tavocg/pm/dispatcher"
	"github.com/tavocg/pm/models"
)

var (
	cfgFile string
	orch    *models.Orchestrator
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pm",
	Short: "Run package manager tasks",
	Long: `pm reads its configuration and runs package manager tasks such as
syncing remote package definitions and updating installed packages.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pm.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find config directories.
		for _, d := range findConfigDirs() {
			viper.AddConfigPath(d)
		}

		viper.SetConfigType("yaml")
		viper.SetConfigName("pm")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	cobra.CheckErr(err)

	initOrchestrator()
}

func findConfigDirs() []string {
	dirs := []string{}

	if xdgCfgDir := os.Getenv("XDG_CONFIG_HOME"); xdgCfgDir != "" {
		dirs = append(dirs, xdgCfgDir)
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if defCfgDir := path.Join(home, ".config"); !slices.Contains(dirs, defCfgDir) {
			dirs = append(dirs, defCfgDir)
		}
	}

	dirs = append(dirs, "/usr/local/etc", "/etc")

	return dirs
}

func initOrchestrator() {
	var cfg models.OrchestratorConfig
	viper.Unmarshal(&cfg)

	ctx := context.Background()
	disp := dispatcher.DefaultDispatcher(ctx)

	params := &models.OrchestratorParams{
		Cfg:        &cfg,
		Dispatcher: disp,
	}

	var err error
	orch, err = models.NewOrchestrator(params)
	cobra.CheckErr(err)

	orch.RemoveUnsupportedManagers()
}
