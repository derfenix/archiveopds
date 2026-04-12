package cli

import (
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "archiveopds",
		Short: "OPDS-сервер для каталогов книг в крупных архивах",
	}
	root.AddCommand(newServe(), newConfigCmd())
	return root
}
