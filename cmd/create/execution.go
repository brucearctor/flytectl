package create

import (
	"context"
	"fmt"
	"strings"

	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	executionShort = "Create execution resources"
	executionLong  = `
Create the executions for given launchplan in a project and domain.
Currently the launchplan and where the execution needs to be launched , need to belong to same domain and project.

There are three steps in generating an execution.

- Generate the execution param yaml file containing in the default input parameters for the launch plan.
- If the default parameters exist then you can modify the default param values in the generated yaml file.
- Run the execution passing in the generated yaml file.

The following command would generate core.basic.lp.go_greet.inputs.yaml file by using the launch plan provided in argument.
This would use the latest version of the launch plan.
::

 bin/flytectl create execution core.basic.lp.go_greet  -d development  -p flytesnacks -g

The generated file would look similar to this

.. code-block:: yaml

	inputs:
	  am: false
	  day_of_week: ""
	  number: 0

The generated file can be modified to change the parameter values and can be passed to command for execution
::

 bin/flytectl create execution core.basic.lp.go_greet -f core.basic.lp.go_greet.inputs.yaml  -d development  -p flytesnacks

Usage
`
)

//go:generate pflags ExecutionConfig --default-var executionConfig

// ProjectConfig Config hold configuration for project create flags.
type ExecutionConfig struct {
	File     string `json:"file" pflag:",file for the execution params.If not specified defaults to <launchplan_name>.inputs.yaml"`
	Version  string `json:"version" pflag:",version of the launch plan to be executed."`
	GenParam bool   `json:"genParam" pflag:",bool flag to indicate the generation of params file."`
}

var (
	executionConfig = &ExecutionConfig{}
)

func createExecutionCommand(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	if len(args) == 0 {
		return errors.New("pass in the launch plan to execute")
	}
	lpName := args[0]
	var lp *admin.LaunchPlan
	var err error
	if lp, err = getLaunchPlan(ctx, lpName, cmdCtx); err != nil {
		return err
	}
	var paramMap map[string]interface{}
	if paramMap, err = getParamMap(lp); err != nil {
		return err
	}
	var fileName string
	// Generate the param file
	if executionConfig.File != "" {
		fileName = executionConfig.File
	} else {
		fileName = lpName + ".inputs.yaml"
	}
	if executionConfig.GenParam {
		if err = writeParamsToFile(paramMap, fileName); err != nil {
			return err
		}
		fmt.Printf("params written to file %v\n", fileName)
		fmt.Printf("run followup command\nflytectl create execution %v -f %v\n", lpName, fileName)
		return nil
	}
	var yamlData map[string]interface{}
	if yamlData, err = readParamsFromFile(fileName); err != nil {
		return err
	}
	// Convert to Literal map
	var inputs *core.LiteralMap
	if inputs, err = createLiteralMapFromParams(yamlData, lp.Spec.DefaultInputs.Parameters); err != nil {
		return err
	}
	exec, _err := cmdCtx.AdminClient().CreateExecution(ctx, &admin.ExecutionCreateRequest{
		Project: config.GetConfig().Project,
		Domain:  config.GetConfig().Domain,
		Name:    "f" + strings.ReplaceAll(uuid.New().String(), "-", "")[:19],
		Spec: &admin.ExecutionSpec{
			LaunchPlan: lp.Id,
			Metadata: &admin.ExecutionMetadata{
				Mode:      admin.ExecutionMetadata_MANUAL,
				Principal: "sdk",
				Nesting:   0,
			},
		},
		Inputs: inputs,
	})
	if _err != nil {
		return _err
	}
	fmt.Printf("workflow execution identifier %v\n", exec.Id)
	return nil
}
