package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/hntrl/hyper/src/hyper/domain"
	"github.com/hntrl/hyper/src/hyper/interfaces"
	"github.com/hntrl/hyper/src/hyper/runtime"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(docCommand)
}

var docCommand = &cobra.Command{
	Use:   "doc [FILE]",
	Short: "Prints out all the objects exported by a hyper context",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		inFile := "./index.hyper"
		if len(args) > 0 {
			inFile = args[0]
		}
		inPath := filepath.Join(dir, inFile)

		manifestTree, err := domain.ParseContextFromFile(inPath)
		if err != nil {
			panic(err)
		}
		builder := domain.NewContextBuilder()
		process := runtime.NewProcess()
		interfaces.RegisterDefaults(builder, process)
		ctx, err := builder.ParseContext(*manifestTree, inPath)
		if err != nil {
			panic(err)
		}

		fmt.Println("\nobjects:\n-------------")
		w := tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)
		for k, v := range ctx.Items {
			fmt.Fprintf(w, "%s\t%T\n", k, v.HostItem)
		}
		w.Flush()
	},
}
