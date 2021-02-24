package create

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flyteidl/clients/go/coreutils"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/pkg/errors"
	"io/ioutil"
	"sigs.k8s.io/yaml"
	"strings"
)

const (
	executionShort = "Create execution resources"
	executionLong  = `
Create the execution.
`
)

//go:generate pflags ExecutionConfig --default-var executionConfig

// ProjectConfig Config hold configuration for project create flags.
type ExecutionConfig struct {
	File    string `json:"file" pflag:",file for the project definition."`
	Version string `json:"version" pflag:",version of the entity to be registered with flyte."`
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
	if executionConfig.Version != "" {
		lp, err = cmdCtx.AdminClient().GetLaunchPlan(ctx, &admin.ObjectGetRequest{
			Id: &core.Identifier{
				ResourceType: core.ResourceType_LAUNCH_PLAN,
				Project:      config.GetConfig().Project,
				Domain:       config.GetConfig().Domain,
				Name:         lpName,
				Version:      executionConfig.Version,
			},
		})
		if err != nil {
			return err
		}
	} else {
		var lpList *admin.LaunchPlanList
		lpList, err = cmdCtx.AdminClient().ListLaunchPlans(ctx, &admin.ResourceListRequest{
			Limit: 10,
			Id: &admin.NamedEntityIdentifier{
				Project: config.GetConfig().Project,
				Domain:  config.GetConfig().Domain,
				Name:    lpName,
			},
			SortBy: &admin.Sort{
				Key:       "created_at",
				Direction: admin.Sort_DESCENDING,
			},
		})
		if err != nil {
			return err
		}
		if len(lpList.LaunchPlans) == 0 {
			return errors.Errorf("no launch plans retrieved for %v", lpName)
		}
		lp = lpList.LaunchPlans[0]
	}
	//fmt.Println("fetched launch plan default inputs ", *lp.Spec.DefaultInputs)
	yamlMap := make(map[string]interface{})
	for k, v := range lp.Spec.DefaultInputs.Parameters {
		varTypeValue, err := coreutils.MakeDefaultLiteralForType(v.Var.Type)
		if err != nil {
			fmt.Println("error creating default value for literal type ", v.Var.Type)
			return err
		}
		if yamlMap[k], err = coreutils.FetchFromLiteral(varTypeValue); err != nil {
			return err
		}
		// Override if there is a default value
		switch v.Behavior.(type) {
		case *core.Parameter_Default:
			if yamlMap[k], err = coreutils.FetchFromLiteral(v.Behavior.(*core.Parameter_Default).Default); err != nil {
				return err
			}
			break
		case *core.Parameter_Required:
			break
		}
	}

	//yamlWithSections := make(map[string]interface{})
	//yamlWithSections["inputs"] = yamlMap
	//d, err := yaml.Marshal(&yamlWithSections)
	//if err != nil {
	//	fmt.Printf("error: %v", err)
	//}

	//var outputFile string
	//if executionConfig.File != "" {
	//	outputFile = executionConfig.File
	//} else {
	//	outputFile = lpName + ".inputs.yaml"
	//}
	//err = ioutil.WriteFile(outputFile, d, 0644)
	//if err != nil {
	//	return errors.New(fmt.Sprintf("unable to write in %v yaml file", executionConfig.File))
	//}

	data, _err := ioutil.ReadFile(lpName + ".inputs.yaml")
	if _err != nil {
		return errors.New(fmt.Sprintf("unable to read from %v yaml file", executionConfig.File))
	}
	fmt.Printf("Params: \n%v\n", string(data))
	yamlReadWithSection := make(map[string]map[string]interface{})
	yaml.Unmarshal(data, &yamlReadWithSection)
	//fmt.Printf("unmarshalled data %v\n", yamlReadWithSection["inputs"])
	//uuidVal := uuid.New()
	var yamlData map[string]interface{}
	yamlData = yamlReadWithSection["inputs"]
	// Convert to Literal map
	inputs := &core.LiteralMap{
		Literals: make(map[string]*core.Literal, len(yamlData)),
	}
	for k, v := range lp.Spec.DefaultInputs.Parameters {
		switch v.Var.GetType().GetType().(type) {
		case *core.LiteralType_Simple:
			simple := v.Var.GetType().GetType().(*core.LiteralType_Simple).Simple
			literalVal, _err := coreutils.MakeLiteralForSimpleType(simple, fmt.Sprintf("%v", yamlData[k]))
			if _err != nil {
				return _err
			}
			inputs.Literals[k]=literalVal
		}
	}
	//fmt.Printf("Args : \n%v \n", inputs)
	exec, _err := cmdCtx.AdminClient().CreateExecution(ctx, &admin.ExecutionCreateRequest{
		Project: config.GetConfig().Project,
		Domain:  config.GetConfig().Domain,
		Name:    "f" + strings.ReplaceAll(uuid.New().String(),"-","")[:19],
		Spec: &admin.ExecutionSpec{
			LaunchPlan: lp.Id,
			Metadata: &admin.ExecutionMetadata{
				Mode: admin.ExecutionMetadata_MANUAL,
				Principal: "sdk",
				Nesting: 0,
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

