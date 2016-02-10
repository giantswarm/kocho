package aws

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/juju/errgo"
)

const (
	generatedCloudformationPath = "/tmp/aws-cloudformation.json"

	primaryCloudFormationTemplateName    = "primary-cloudformation.tmpl"
	secondaryCloudFormationTemplateName  = "secondary-cloudformation.tmpl"
	standaloneCloudFormationTemplateName = "standalone-cloudformation.tmpl"

	cloudFormationStackTag = "aws:cloudformation:stack-name"
)

type primaryCloudformation struct {
	Name              string
	Machines          []int // the index of the machines to iterate over in the template
	MachineReferences string
	Type              string
	VPCCIDR           string
}

type secondaryCloudformation struct {
	Type    string
	VPCCIDR string
}

type standaloneCloudformation struct {
	Type    string
	VPCCIDR string
}

type machineReference struct {
	Ref string
}

func createMachineReferences(clusterSize int) string {
	machineReferences := []machineReference{}
	for i := 0; i < clusterSize; i++ {
		machineReferences = append(machineReferences, machineReference{Ref: fmt.Sprintf("Machine%d", i)})
	}
	jsonList, err := json.Marshal(machineReferences)
	if err != nil {
		panic(err)
	}
	return string(jsonList)
}

func createPrimaryCloudformationTemplate(name string, clusterSize int, templateDir string, vpccidr string) (string, error) {
	machineIds := make([]int, clusterSize)
	for id, _ := range machineIds {
		machineIds[id] = id
	}

	cloudFormationTemplatePath := path.Join(templateDir, primaryCloudFormationTemplateName)

	return parseCloudformationTemplate(cloudFormationTemplatePath, primaryCloudformation{
		Name:              name,
		Machines:          machineIds,
		MachineReferences: createMachineReferences(clusterSize),
		Type:              "primary",
		VPCCIDR:           vpccidr,
	})
}

func createSecondaryCloudformationTemplate(templateDir string, vpccidr string) (string, error) {
	cloudFormationTemplatePath := path.Join(templateDir, secondaryCloudFormationTemplateName)

	return parseCloudformationTemplate(cloudFormationTemplatePath, secondaryCloudformation{
		Type:    "secondary",
		VPCCIDR: vpccidr,
	})
}

func createStandaloneCloudformationTemplate(templateDir string, vpccidr string) (string, error) {
	cloudFormationTemplatePath := path.Join(templateDir, standaloneCloudFormationTemplateName)

	return parseCloudformationTemplate(cloudFormationTemplatePath, standaloneCloudformation{
		Type:    "standalone",
		VPCCIDR: vpccidr,
	})
}

func parseCloudformationTemplate(templatePath string, cfg interface{}) (string, error) {
	f, err := os.Create(generatedCloudformationPath)
	if err != nil {
		return generatedCloudformationPath, errgo.Mask(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	absoluteCloudFormationTemplatePath, err := filepath.Abs(templatePath)
	if err != nil {
		return "", errgo.Mask(err)
	}

	templateData, err := ioutil.ReadFile(absoluteCloudFormationTemplatePath)
	if err != nil {
		return "", errgo.Mask(err)
	}

	var tmpl *template.Template
	if tmpl, err = template.New("cloudformation").Parse(string(templateData)); err != nil {
		return generatedCloudformationPath, errgo.Mask(err)
	}

	if err = tmpl.Execute(f, cfg); err != nil {
		return generatedCloudformationPath, errgo.Mask(err)
	}

	return generatedCloudformationPath, nil
}
