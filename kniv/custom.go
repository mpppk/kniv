package kniv

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/robertkrimen/otto"
)

type CustomProcessor struct {
	*BaseProcessor
	tasks []FilterTask
}

type TaskType string
type TaskMode string

type FilterTask interface {
	Run(payload EventPayload) (EventPayload, error)
}

type FilterByJSTask struct {
	commands []string
	vm       *otto.Otto
}

func (t *FilterByJSTask) Run(payload EventPayload) (EventPayload, error) {
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

func NewFilterByJSTask(commands []string) *FilterByJSTask {
	return &FilterByJSTask{
		vm:       otto.New(),
		commands: commands,
	}
}

type TransformByJSTask struct {
	commands []string
	vm       *otto.Otto
}

func (t *TransformByJSTask) Run(payload EventPayload) (EventPayload, error) {
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

func NewTransformByJSTask(commands []string) *TransformByJSTask {
	return &TransformByJSTask{
		vm:       otto.New(),
		commands: commands,
	}
}

type DistinctTask struct {
	distinctKeyMap map[string]bool
	keys           []string
}

func (t *DistinctTask) Run(payload EventPayload) (EventPayload, error) {
	distinct, err := checkAndUpdateDuplicatedEvent(t.distinctKeyMap, payload, t.keys)
	if err != nil {
		return nil, err
	}
	if distinct {
		return nil, err
	}
	return payload, nil
}

func NewDistinctTask(keys []string) *DistinctTask {
	return &DistinctTask{
		distinctKeyMap: map[string]bool{},
		keys:           keys,
	}
}

type SelectPayloadTask struct {
	keys []string
}

func (t *SelectPayloadTask) Run(payload EventPayload) (EventPayload, error) {
	newPayload := EventPayload{}
	for _, key := range t.keys {
		newPayload[key] = payload[key]
	}
	return newPayload, nil
}

func NewSelectPayloadTask(keys []string) *SelectPayloadTask {
	return &SelectPayloadTask{
		keys: keys,
	}
}

func NewCustomProcessor(queueSize int, tasks []FilterTask) *CustomProcessor {
	customProcessor := &CustomProcessor{
		BaseProcessor: &BaseProcessor{
			Name:   "custom",
			inChan: make(chan Event, queueSize),
		},
		tasks: tasks,
	}
	customProcessor.BaseProcessor.Process = customProcessor.Process
	return customProcessor
}

func (p *CustomProcessor) Process(event Event) ([]Event, error) {
	payload := event.GetPayload()

	newPayload := payload
	for _, task := range p.tasks {
		log.Printf("%d: %T is started. payload: %#v", event.GetId(), task, newPayload)
		p, err := task.Run(newPayload)

		if err != nil {
			return nil, err
		}

		if p == nil {
			log.Printf("%d: event is filtered by %T", event.GetId(), task)
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
