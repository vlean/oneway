package main

import (
	"errors"
	"os"

	"gihub.com/vlean/oneway/config"
	_ "gihub.com/vlean/oneway/log"
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	root = &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			return errors.New("test")
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			cont, err := os.ReadFile(cfgFile)
			if err != nil {
				return
			}
			_, err = toml.Decode(string(cont), config.Global())
			return
		},
	}

	cfgFile = ""
)

func main() {
	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.toml", "-c config.toml")
	err := root.Execute()
	if err != nil {
		log.Infof("run error: %v", err)
	}
}

