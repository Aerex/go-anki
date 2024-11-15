// Code generated by counterfeiter. DO NOT EDIT.
package editorfakes

import (
	"sync"

	"github.com/aerex/go-anki/pkg/editor"
)

type FakeEditor struct {
	CloneStub        func() error
	cloneMutex       sync.RWMutex
	cloneArgsForCall []struct {
	}
	cloneReturns struct {
		result1 error
	}
	cloneReturnsOnCall map[int]struct {
		result1 error
	}
	ConfirmUserErrorStub        func() bool
	confirmUserErrorMutex       sync.RWMutex
	confirmUserErrorArgsForCall []struct {
	}
	confirmUserErrorReturns struct {
		result1 bool
	}
	confirmUserErrorReturnsOnCall map[int]struct {
		result1 bool
	}
	CreateStub        func() error
	createMutex       sync.RWMutex
	createArgsForCall []struct {
	}
	createReturns struct {
		result1 error
	}
	createReturnsOnCall map[int]struct {
		result1 error
	}
	EditStub        func(interface{}) (error, []byte, bool)
	editMutex       sync.RWMutex
	editArgsForCall []struct {
		arg1 interface{}
	}
	editReturns struct {
		result1 error
		result2 []byte
		result3 bool
	}
	editReturnsOnCall map[int]struct {
		result1 error
		result2 []byte
		result3 bool
	}
	RemoveStub        func() error
	removeMutex       sync.RWMutex
	removeArgsForCall []struct {
	}
	removeReturns struct {
		result1 error
	}
	removeReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeEditor) Clone() error {
	fake.cloneMutex.Lock()
	ret, specificReturn := fake.cloneReturnsOnCall[len(fake.cloneArgsForCall)]
	fake.cloneArgsForCall = append(fake.cloneArgsForCall, struct {
	}{})
	stub := fake.CloneStub
	fakeReturns := fake.cloneReturns
	fake.recordInvocation("Clone", []interface{}{})
	fake.cloneMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeEditor) CloneCallCount() int {
	fake.cloneMutex.RLock()
	defer fake.cloneMutex.RUnlock()
	return len(fake.cloneArgsForCall)
}

func (fake *FakeEditor) CloneCalls(stub func() error) {
	fake.cloneMutex.Lock()
	defer fake.cloneMutex.Unlock()
	fake.CloneStub = stub
}

func (fake *FakeEditor) CloneReturns(result1 error) {
	fake.cloneMutex.Lock()
	defer fake.cloneMutex.Unlock()
	fake.CloneStub = nil
	fake.cloneReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeEditor) CloneReturnsOnCall(i int, result1 error) {
	fake.cloneMutex.Lock()
	defer fake.cloneMutex.Unlock()
	fake.CloneStub = nil
	if fake.cloneReturnsOnCall == nil {
		fake.cloneReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.cloneReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeEditor) ConfirmUserError() bool {
	fake.confirmUserErrorMutex.Lock()
	ret, specificReturn := fake.confirmUserErrorReturnsOnCall[len(fake.confirmUserErrorArgsForCall)]
	fake.confirmUserErrorArgsForCall = append(fake.confirmUserErrorArgsForCall, struct {
	}{})
	stub := fake.ConfirmUserErrorStub
	fakeReturns := fake.confirmUserErrorReturns
	fake.recordInvocation("ConfirmUserError", []interface{}{})
	fake.confirmUserErrorMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeEditor) ConfirmUserErrorCallCount() int {
	fake.confirmUserErrorMutex.RLock()
	defer fake.confirmUserErrorMutex.RUnlock()
	return len(fake.confirmUserErrorArgsForCall)
}

func (fake *FakeEditor) ConfirmUserErrorCalls(stub func() bool) {
	fake.confirmUserErrorMutex.Lock()
	defer fake.confirmUserErrorMutex.Unlock()
	fake.ConfirmUserErrorStub = stub
}

func (fake *FakeEditor) ConfirmUserErrorReturns(result1 bool) {
	fake.confirmUserErrorMutex.Lock()
	defer fake.confirmUserErrorMutex.Unlock()
	fake.ConfirmUserErrorStub = nil
	fake.confirmUserErrorReturns = struct {
		result1 bool
	}{result1}
}

func (fake *FakeEditor) ConfirmUserErrorReturnsOnCall(i int, result1 bool) {
	fake.confirmUserErrorMutex.Lock()
	defer fake.confirmUserErrorMutex.Unlock()
	fake.ConfirmUserErrorStub = nil
	if fake.confirmUserErrorReturnsOnCall == nil {
		fake.confirmUserErrorReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.confirmUserErrorReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *FakeEditor) Create() error {
	fake.createMutex.Lock()
	ret, specificReturn := fake.createReturnsOnCall[len(fake.createArgsForCall)]
	fake.createArgsForCall = append(fake.createArgsForCall, struct {
	}{})
	stub := fake.CreateStub
	fakeReturns := fake.createReturns
	fake.recordInvocation("Create", []interface{}{})
	fake.createMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeEditor) CreateCallCount() int {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	return len(fake.createArgsForCall)
}

func (fake *FakeEditor) CreateCalls(stub func() error) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.CreateStub = stub
}

func (fake *FakeEditor) CreateReturns(result1 error) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.CreateStub = nil
	fake.createReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeEditor) CreateReturnsOnCall(i int, result1 error) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.CreateStub = nil
	if fake.createReturnsOnCall == nil {
		fake.createReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.createReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeEditor) Edit(arg1 interface{}) (error, []byte, bool) {
	fake.editMutex.Lock()
	ret, specificReturn := fake.editReturnsOnCall[len(fake.editArgsForCall)]
	fake.editArgsForCall = append(fake.editArgsForCall, struct {
		arg1 interface{}
	}{arg1})
	stub := fake.EditStub
	fakeReturns := fake.editReturns
	fake.recordInvocation("Edit", []interface{}{arg1})
	fake.editMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeEditor) EditCallCount() int {
	fake.editMutex.RLock()
	defer fake.editMutex.RUnlock()
	return len(fake.editArgsForCall)
}

func (fake *FakeEditor) EditCalls(stub func(interface{}) (error, []byte, bool)) {
	fake.editMutex.Lock()
	defer fake.editMutex.Unlock()
	fake.EditStub = stub
}

func (fake *FakeEditor) EditArgsForCall(i int) interface{} {
	fake.editMutex.RLock()
	defer fake.editMutex.RUnlock()
	argsForCall := fake.editArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeEditor) EditReturns(result1 error, result2 []byte, result3 bool) {
	fake.editMutex.Lock()
	defer fake.editMutex.Unlock()
	fake.EditStub = nil
	fake.editReturns = struct {
		result1 error
		result2 []byte
		result3 bool
	}{result1, result2, result3}
}

func (fake *FakeEditor) EditReturnsOnCall(i int, result1 error, result2 []byte, result3 bool) {
	fake.editMutex.Lock()
	defer fake.editMutex.Unlock()
	fake.EditStub = nil
	if fake.editReturnsOnCall == nil {
		fake.editReturnsOnCall = make(map[int]struct {
			result1 error
			result2 []byte
			result3 bool
		})
	}
	fake.editReturnsOnCall[i] = struct {
		result1 error
		result2 []byte
		result3 bool
	}{result1, result2, result3}
}

func (fake *FakeEditor) Remove() error {
	fake.removeMutex.Lock()
	ret, specificReturn := fake.removeReturnsOnCall[len(fake.removeArgsForCall)]
	fake.removeArgsForCall = append(fake.removeArgsForCall, struct {
	}{})
	stub := fake.RemoveStub
	fakeReturns := fake.removeReturns
	fake.recordInvocation("Remove", []interface{}{})
	fake.removeMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeEditor) RemoveCallCount() int {
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	return len(fake.removeArgsForCall)
}

func (fake *FakeEditor) RemoveCalls(stub func() error) {
	fake.removeMutex.Lock()
	defer fake.removeMutex.Unlock()
	fake.RemoveStub = stub
}

func (fake *FakeEditor) RemoveReturns(result1 error) {
	fake.removeMutex.Lock()
	defer fake.removeMutex.Unlock()
	fake.RemoveStub = nil
	fake.removeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeEditor) RemoveReturnsOnCall(i int, result1 error) {
	fake.removeMutex.Lock()
	defer fake.removeMutex.Unlock()
	fake.RemoveStub = nil
	if fake.removeReturnsOnCall == nil {
		fake.removeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.removeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeEditor) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.cloneMutex.RLock()
	defer fake.cloneMutex.RUnlock()
	fake.confirmUserErrorMutex.RLock()
	defer fake.confirmUserErrorMutex.RUnlock()
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	fake.editMutex.RLock()
	defer fake.editMutex.RUnlock()
	fake.removeMutex.RLock()
	defer fake.removeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeEditor) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ editor.Editor = new(FakeEditor)
