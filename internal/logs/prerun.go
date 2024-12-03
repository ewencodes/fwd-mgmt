package logs

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func ToggleDebug(cmd *cobra.Command, args []string) {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
	})
	if debug, _ := cmd.Flags().GetBool("debug"); debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug logs enabled")
	}
}
