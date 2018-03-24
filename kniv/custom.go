package kniv

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/robertkrimen/otto"
)

const customProcessorType = "custom"

type logicType string

// FIXME load by side effect
const (
	filterByJSType    logicType = "filter_event_by_js"
	transformByJSType logicType = "transform_by_js"
	distinctType      logicType = "distinct"
	selectPayloadType logicType = "select_payload"
)

type CustomProcessor struct {
	*BaseProcessor
	logics []CustomLogic
}

type LogicType string

type CustomLogic interface {
	Run(payload EventPayload) (EventPayload, error)
}

type FilterByJSLogic struct {
	commands []string
	vm       *otto.Otto
}

func (t *FilterByJSLogic) Run(payload EventPayload) (EventPayload, error) {
	ok, err := filterEventByJS(t.vm, payload, t.commands)
	if err != nil {
		return nil, err
	}
	if ok {
		return payload, nil
	}
	return nil, nil
}

func filterEventByJS(vm *otto.Otto, payload EventPayload, commands []string) (bool, error) {
	jsPayload, err := vm.ToValue(payload)
	if err != nil {
		return false, err
	}

	vm.Set("p", jsPayload)

	joinedCommands := strings.Join(commands, ";") + ";"
	result, err := vm.Run(joinedCommands)
	if err != nil {
		return false, err
	}

	return result.ToBoolean()
}

func NewFilterByJSLogic(commands []string) *FilterByJSLogic {
	return &FilterByJSLogic{
		vm:       otto.New(),
		commands: commands,
	}
}

type TransformByJSLogic struct {
	commands []string
	vm       *otto.Otto
}

func (t *TransformByJSLogic) Run(payload EventPayload) (EventPayload, error) {
	return transformPayloadByJS(t.vm, payload, t.commands)
}

func transformPayloadByJS(vm *otto.Otto, payload EventPayload, commands []string) (EventPayload, error) {
	jsPayload, err := vm.ToValue(payload)
	if err != nil {
		return nil, err
	}

	vm.Set("p", jsPayload)

	newCommands := append(commands, "result = p;")

	joinedCommands := strings.Join(newCommands, ";")
	result, err := vm.Run(joinedCommands)
	if err != nil {
		return nil, err
	}

	retObj, err := result.Export()
	if err != nil {
		return nil, err
	}

	retPayload, ok := retObj.(EventPayload)
	if !ok {
		return nil, errors.New("js execution is failed")
	}
	return retPayload, nil
}

func NewTransformByJSLogic(commands []string) *TransformByJSLogic {
	return &TransformByJSLogic{
		vm:       otto.New(),
		commands: commands,
	}
}

type DistinctLogic struct {
	distinctKeyMap map[string]bool
	keys           []string
}

func (t *DistinctLogic) Run(payload EventPayload) (EventPayload, error) {
	distinct, err := checkAndUpdateDuplicatedEvent(t.distinctKeyMap, payload, t.keys)
	if err != nil {
		return nil, err
	}
	if distinct {
		return nil, err
	}
	return payload, nil
}

func NewDistinctLogic(keys []string) *DistinctLogic {
	return &DistinctLogic{
		distinctKeyMap: map[string]bool{},
		keys:           keys,
	}
}

type SelectPayloadLogic struct {
	keys []string
}

func (t *SelectPayloadLogic) Run(payload EventPayload) (EventPayload, error) {
	newPayload := EventPayload{}
	for _, key := range t.keys {
		newPayload[key] = payload[key]
	}
	return newPayload, nil
}

func NewSelectPayloadLogic(keys []string) *SelectPayloadLogic {
	return &SelectPayloadLogic{
		keys: keys,
	}
}

func NewCustomProcessor(queueSize int, logics []CustomLogic) *CustomProcessor {
	customProcessor := &CustomProcessor{
		BaseProcessor: &BaseProcessor{
			Name:   "custom",
			inChan: make(chan Event, queueSize),
		},
		logics: logics,
	}
	customProcessor.BaseProcessor.Process = customProcessor.Process
	return customProcessor
}

func (p *CustomProcessor) Process(event Event) ([]Event, error) {
	payload := event.GetPayload()

	newPayload := payload
	for _, logic := range p.logics {
		log.Printf("%d: %T is started. payload: %#v", event.GetId(), logic, newPayload)
		p, err := logic.Run(newPayload)

		if err != nil {
			return nil, err
		}

		if p == nil {
			log.Printf("%d: event is filtered by %T", event.GetId(), logic)
			return nil, nil
		}
		newPayload = p
	}
	event.SetPayload(newPayload)
	return []Event{event}, nil
}

func createStrForDuplicateChecking(payload EventPayload, keys []string) (string, error) {
	// FIXME convert distinctKey to hash for efficient memory management
	distinctKey := ""
	for _, key := range keys {
		v, ok := payload[key]
		if !ok {
			return "", errors.New(key + " not found in filter")
		}

		distinctKey += fmt.Sprint(v)
	}
	return distinctKey, nil
}

func checkAndUpdateDuplicatedEvent(distinctKeyMap map[string]bool, payload EventPayload, keys []string) (bool, error) {
	checkStr, err := createStrForDuplicateChecking(payload, keys)
	if err != nil {
		return false, err
	}

	_, exist := distinctKeyMap[checkStr]
	if exist {
		return false, nil
	}

	distinctKeyMap[checkStr] = true
	return exist, err
}
