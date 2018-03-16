package kniv

import (
	"errors"
	"fmt"
	"github.com/robertkrimen/otto"
)

type Filter struct {
	*BaseProcessor
	opt            *FilterOpt
	distinctKeyMap map[string]bool
	vm             *otto.Otto
}

type TaskType string
type TaskMode string

type FilterTask struct {
	taskType TaskType
	mode     TaskMode
	commands []string
	keys     []string
}

type FilterOpt struct {
	selectPayloadKeys                  []string
	selectJSCommands                   []string
	filterPayloadKeys                  []string
	filterEventIfEveryPayloadKeysExist []string
	filterEventIfSomePayloadKeysExist  []string
	transformJSCommands                []string
	distinctEventByPayloadKeys         []string
}

func NewFilter(queueSize int, opt *FilterOpt) *Filter {
	filter := &Filter{
		BaseProcessor: &BaseProcessor{
			Name:   "filter",
			inChan: make(chan Event, queueSize),
		},
		opt: opt,
	}
	return filter
}

func (f *Filter) filter(event Event) ([]Event, error) {
	if event.GetPayload().HasEveryPayloadKeys(f.opt.filterEventIfEveryPayloadKeysExist) ||
		event.GetPayload().HasSomePayloadKeys(f.opt.filterEventIfSomePayloadKeysExist) {
		return []Event{}, nil
	}

	duplicated, err := f.checkAndUpdateDuplicatedEvent(event)
	if err != nil {
		return nil, err
	}

	if duplicated {
		return []Event{}, nil
	}

	return []Event{event}, nil
}

func createStrForDuplicateChecking(payload EventPayload, keys []string) (string, error) {
	// FIXME convert distinctKey to hash
	distinctKey := ""
	for _, key := range keys {
		v, ok := payload[key]
		if !ok {
			return "", errors.New(key + " not found in filter")
		}

		vstr, ok := v.(fmt.Stringer)
		if !ok {
			return "", errors.New("value of " + key + " is not stringer in filter")
		}
		distinctKey += vstr.String()
	}
	return distinctKey, nil
}

func (f *Filter) checkAndUpdateDuplicatedEvent(event Event) (bool, error) {
	checkStr, err := createStrForDuplicateChecking(event.GetPayload(), f.opt.distinctEventByPayloadKeys)
	if err != nil {
		return false, err
	}

	_, exist := f.distinctKeyMap[checkStr]
	if exist {
		return false, nil
	}

	f.distinctKeyMap[checkStr] = true
	return exist, err
}
