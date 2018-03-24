package kniv

import "fmt"

type CustomLogicSetting struct {
	Type     logicType
	Commands []string
	Keys     []string
}

type CustomProcessorArgs struct {
	Logics []CustomLogicSetting
}

type CustomProcessorGenerator struct{}

func (g *CustomProcessorGenerator) Generate(intfArgs interface{}) (Processor, error) {
	args, ok := intfArgs.(CustomProcessorArgs)
	if !ok {
		return nil, fmt.Errorf("invalid delay processor config: %#v", intfArgs)
	}

	logics, err := argsToLogics(args.Logics)
	if err != nil {
		return nil, err
	}
	return NewCustomProcessor(100000, logics), nil // FIXME queue size
}

func (g *CustomProcessorGenerator) GetType() string {
	return customProcessorType
}

func argsToLogics(logicSettings []CustomLogicSetting) (logics []CustomLogic, err error) {
	for _, logicSetting := range logicSettings {
		logic, err := logicSettingToLogic(logicSetting)
		if err != nil {
			return nil, err
		}
		logics = append(logics, logic)
	}
	return logics, nil
}

func logicSettingToLogic(setting CustomLogicSetting) (logic CustomLogic, err error) {
	switch setting.Type {
	case filterByJSType:
		{
			logic = NewFilterByJSLogic(setting.Commands)
		}
	case transformByJSType:
		{
			logic = NewTransformByJSLogic(setting.Commands)
		}
	case distinctType:
		{
			logic = NewDistinctLogic(setting.Keys)
		}
	case selectPayloadType:
		{
			logic = NewSelectPayloadLogic(setting.Keys)
		}
	default:
		return nil, fmt.Errorf("logic type %s not found", setting.Type)
	}
	return logic, nil
}
