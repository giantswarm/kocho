package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

var (
	templatesDir            string = "default-templates"
	cloudFormationTemplates        = []string{
		"primary-cloudformation.tmpl",
		"secondary-cloudformation.tmpl",
		"standalone-cloudformation.tmpl",

		"primary-parameters.tmpl",
		"secondary-parameters.tmpl",
		"standalone-parameters.tmpl",
	}
	cloudConfigTemplates = []string{
		"primary-cloudconfig.tmpl",
		"secondary-cloudconfig.tmpl",
		"standalone-cloudconfig.tmpl",
	}
	ignitionTemplates = []string{
		"primary-ignition.tmpl",
		"secondary-ignition.tmpl",
		"standalone-ignition.tmpl",
	}

	flagTemplateDir string
	flagForce       bool
	flagUseIgnition bool
	cmdTemplateInit = &Command{
		Name:        "template-init",
		Description: "Initialise templates",
		Summary:     "Initialise templates from their default values",
		Run:         runTemplateInit,
	}
)

func init() {
	cmdTemplateInit.Flags.StringVar(&flagTemplateDir, "template-dir", "templates", "directory to write templates to")
	cmdTemplateInit.Flags.BoolVar(&flagForce, "force", false, "overwriting existing templates")
	cmdTemplateInit.Flags.BoolVar(&flagUseIgnition, "use-ignition", false, "use ignition configuration templates")
}

func runTemplateInit(args []string) (exit int) {
	// Create template directory if it doesn't exist
	if _, err := os.Stat(flagTemplateDir); err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(flagTemplateDir, os.ModePerm); err != nil {
			return exitError(fmt.Sprintf("couldn't create template directory: %s", flagTemplateDir), err)
		}
	}

	templates := make([]string, 0)
	templates = append(templates, cloudFormationTemplates...)
	if flagUseIgnition {
		templates = append(templates, ignitionTemplates...)
	} else {
		templates = append(templates, cloudConfigTemplates...)
	}
	// Write out each template file
	for _, template := range templates {
		fileData, err := Asset(path.Join(templatesDir, template))
		if err != nil {
			return exitError(fmt.Sprintf("couldn't access template asset: %s", template), err)
		}

		absPath, err := filepath.Abs(path.Join(flagTemplateDir, template))
		if err != nil {
			return exitError(fmt.Sprintf("couldn't determine path to template: %s", template), err)
		}

		_, statErr := os.Stat(absPath)

		// The file exists, and we don't want to overwrite, so continue
		if statErr == nil && !flagForce {
			continue
		}

		// The file exists, and we do want to overwrite, so delete the file
		if statErr == nil && flagForce {
			if err := os.Remove(absPath); err != nil {
				return exitError(fmt.Sprintf("couldn't remove template to overwrite: %s", template), err)
			}
		}

		// Write the file, creating it if it doesn't exist
		if err := ioutil.WriteFile(absPath, fileData, 0666); err != nil {
			return exitError(fmt.Sprintf("couldn't write template out: %s", absPath), err)
		}
	}

	return 0
}
