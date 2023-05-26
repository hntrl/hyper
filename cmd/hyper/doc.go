package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/hntrl/hyper/internal/context"
	"github.com/hntrl/hyper/internal/interfaces"
	"github.com/hntrl/hyper/internal/runtime"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(docCommand)
}

var docCommand = &cobra.Command{
	Use:   "doc [FILE]",
	Short: "Prints prints out all the objects exported by a Schemafile",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		inFile := "./Schemafile"
		if len(args) > 0 {
			inFile = args[0]
		}
		inPath := filepath.Join(dir, inFile)

		manifestTree, err := context.ParseContextFromFile(inPath)
		if err != nil {
			panic(err)
		}
		builder := context.NewContextBuilder()
		process := runtime.NewProcess()
		interfaces.RegisterDefaults(builder, process)
		ctx, err := builder.ParseContext(*manifestTree, inPath)
		if err != nil {
			panic(err)
		}

		fmt.Println("\nobjects:\n-------------")
		w := tabwriter.NewWriter(os.Stdout, 1, 1, 2, ' ', 0)
		for k, v := range ctx.Items {
			fmt.Fprintf(w, "%s\t%T\n", k, v)
		}
		w.Flush()
	},
}
