package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"

	interfaces "github.com/hntrl/lang-interfaces"
	runtime "github.com/hntrl/lang-runtime"
	"github.com/hntrl/lang/context"
	"github.com/hntrl/lang/language"

	"github.com/spf13/cobra"
)

var (
	inFile  string
	outFile string
)

func init() {
	buildCommand.PersistentFlags().StringVarP(&inFile, "input", "i", "./Schemafile", "The entrypoint for the bundle")
	buildCommand.PersistentFlags().StringVarP(&outFile, "out", "o", "", "The path where the output should be sent")
	rootCmd.AddCommand(buildCommand)
}

var buildCommand = &cobra.Command{
	Use:   "build",
	Short: "Bundles a Schemafile into a single package",
	Long:  "Serializes a Schemafile and all subsequent imports to be used in the runtime",
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()

		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		inPath := filepath.Join(dir, inFile)

		manifestTree, err := language.ParseContextFromFile(inPath)
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

		buffer, err := builder.Serialize()
		if err != nil {
			panic(err)
		}

		if outFile == "" {
			outFile = ctx.Name
		}
		outPath := filepath.Join(dir, outFile)

		file, err := os.Create(outPath)
		if err != nil {
			panic(err)
		}

		writer := bufio.NewWriter(file)
		nBytes, err := writer.Write(buffer.Bytes())
		if err != nil {
			panic(err)
		}
		fmt.Printf("wrote %v bytes in %s", nBytes, time.Since(start))
	},
}
