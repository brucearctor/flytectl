package create

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/lyft/flytectl/cmd/config"
	cmdCore "github.com/lyft/flytectl/cmd/core"
	"github.com/lyft/flyteidl/clients/go/coreutils"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	"gopkg.in/yaml.v2"
)

func getLaunchPlan(ctx context.Context, lpName string, cmdCtx cmdCore.CommandContext) (*admin.LaunchPlan, error) {
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
			return nil, err
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
			return nil, err
		}
		if len(lpList.LaunchPlans) == 0 {
			return nil, fmt.Errorf("no launch plans retrieved for %v", lpName)
		}
		lp = lpList.LaunchPlans[0]
	}
	return lp, nil
}

func getParamMap(lp *admin.LaunchPlan) (map[string]interface{}, error) {
	paramMap := make(map[string]interface{})
	for k, v := range lp.Spec.DefaultInputs.Parameters {
		varTypeValue, err := coreutils.MakeDefaultLiteralForType(v.Var.Type)
		if err != nil {
			fmt.Println("error creating default value for literal type ", v.Var.Type)
			return nil, err
		}
		if paramMap[k], err = coreutils.FetchFromLiteral(varTypeValue); err != nil {
			return nil, err
		}
		// Override if there is a default value
		if paramsDefault, ok := v.Behavior.(*core.Parameter_Default); ok {
			if paramMap[k], err = coreutils.FetchFromLiteral(paramsDefault.Default); err != nil {
				return nil, err
			}
		}
	}
	return paramMap, nil
}

func writeParamsToFile(paramMap map[string]interface{}, fileName string) error {
	yamlWithSections := make(map[string]interface{})
	yamlWithSections["inputs"] = paramMap
	d, err := yaml.Marshal(&yamlWithSections)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	err = ioutil.WriteFile(fileName, d, 0600)
	if err != nil {
		return fmt.Errorf("unable to write in %v yaml file", executionConfig.File)
	}
	return nil
}

func readParamsFromFile(fileName string) (map[string]interface{}, error) {
	data, _err := ioutil.ReadFile(fileName)
	if _err != nil {
		return nil, fmt.Errorf("unable to read from %v yaml file", executionConfig.File)
	}
	fmt.Printf("Params: \n%v\n", string(data))
	yamlRead := make(map[string]map[string]interface{})
	if _err = yaml.Unmarshal(data, &yamlRead); _err != nil {
		return nil, _err
	}
	return yamlRead["inputs"], nil
}

func createLiteralMapFromParams(paramsMap map[string]interface{}, defaultParams map[string]*core.Parameter) (*core.LiteralMap, error) {
	inputs := &core.LiteralMap{
		Literals: make(map[string]*core.Literal, len(paramsMap)),
	}
	for k, v := range defaultParams {
		if paramsSimple, ok := v.Var.GetType().GetType().(*core.LiteralType_Simple); ok {
			literalVal, err := coreutils.MakeLiteralForSimpleType(paramsSimple.Simple, fmt.Sprintf("%v", paramsMap[k]))
			if err != nil {
				return nil, err
			}
			inputs.Literals[k] = literalVal
		} else {
			return nil, fmt.Errorf("unsupported type %v", v.Var.GetType().GetType())
		}
	}
	return inputs, nil
}
