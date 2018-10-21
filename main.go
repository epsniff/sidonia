package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = 0.1

func main() {

	root := &cobra.Command{
		Use:   "sidonia",
		Short: "sidonia command line tools",
	}

	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "show version of this binary",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(Version)
		},
	})

	root.AddCommand(server.ServerCLI)
	root.Execute()
}
