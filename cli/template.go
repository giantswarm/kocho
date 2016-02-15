package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	templateInitCmd = &cobra.Command{
		Use:   "template-init",
		Short: "Initialise templates",
		Long:  "Initialise templates from their default, in-built values",
		Run:   runTemplateInit,
	}

	templatesDir string = "default-templates"
	templates           = []string{
		"primary-cloudconfig.tmpl",
		"secondary-cloudconfig.tmpl",
		"standalone-cloudconfig.tmpl",

		"primary-cloudformation.tmpl",
		"secondary-cloudformation.tmpl",
		"standalone-cloudformation.tmpl",

		"primary-parameters.tmpl",
		"secondary-parameters.tmpl",
		"standalone-parameters.tmpl",
	}

	flagTemplateDir string
	flagForce       bool
)

func init() {
	templateInitCmd.Flags().StringVar(&flagTemplateDir, "template-dir", "templates", "directory to write templates to")
	templateInitCmd.Flags().BoolVar(&flagForce, "force", false, "overwriting existing templates")

	RootCmd.AddCommand(templateInitCmd)
}

func runTemplateInit(cmd *cobra.Command, args []string) {
	// Create template directory if it doesn't exist
	if _, err := os.Stat(flagTemplateDir); err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(flagTemplateDir, os.ModePerm); err != nil {
			fmt.Printf("couldn't create template directory: %s\n", err)
			return
		}
	}

	// Write out each template file
	for _, template := range templates {
		fileData, err := Asset(path.Join(templatesDir, template))
		if err != nil {
			fmt.Printf("couldn't read template: %s\n", err)
			return
		}

		absPath, err := filepath.Abs(path.Join(flagTemplateDir, template))
		if err != nil {
			fmt.Printf("couldn't determine path to template: %s\n", err)
			return
		}

		_, statErr := os.Stat(absPath)

		// The file exists, and we don't want to overwrite, so continue
		if statErr == nil && !flagForce {
			continue
		}

		// The file exists, and we do want to overwrite, so delete the file
		if statErr == nil && flagForce {
			if err := os.Remove(absPath); err != nil {
				fmt.Printf("couldn't remove template to overwrite: %s\n", err)
				return
			}
		}

		// Write the file, creating it if it doesn't exist
		if err := ioutil.WriteFile(absPath, fileData, 0666); err != nil {
			fmt.Printf("couldn't write template out: %s\n", err)
			return
		}
	}
}
