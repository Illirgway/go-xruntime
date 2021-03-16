/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/
 *
 * (c) 2020-2021 Illirgway
 */

package xruntime

import (
	"reflect"
	"testing"
	"unsafe"
)

// go test -count=1 -o runtime.exe -gcflags "-m -m" -v 2> runtime.log
// go tool objdump -S -s "go-xruntime" runtime.exe > runtime.disasm

type interfaceHeader struct {
	typePtr uintptr
	dataPtr uintptr
}

type fakeObject struct {
	Param string
}

type fakeInterface interface {
	interfaceDataPointer(argE2H interface{}) uintptr
	string2slice(s string) []byte
	string2sliceRough(s string) []byte
	slice2string(b []byte) string
}

type fakeInterfaceReceiver struct{}

// NoEscape

// prevent from inlining and simplifying by compiler
//go:noinline
//go:nosplit
func (f *fakeInterfaceReceiver) interfaceDataPointer(argE2H interface{}) uintptr {
	return (*interfaceHeader)(unsafe.Pointer(&argE2H)).dataPtr
}

func abs(val int) int {

	if val < 0 {
		return int(-val)
	}

	return int(val)
}

// go test -count=1 -o runtime.exe -gcflags "-m -m" -v -run "^TestNoEscape$" 2> runtime.log
// go tool objdump -S -s "go-xruntime" runtime.exe > runtime.disasm

func TestNoEscape(t *testing.T) {

	const (
		Str string = "fake stack allocated object"
	)

	var guarateedStackValuePtr uintptr

	guarateedStackValuePtr = uintptr(unsafe.Pointer(&guarateedStackValuePtr))

	var r fakeInterface = new(fakeInterfaceReceiver)

	var stackAllocated = fakeObject{
		Param: Str,
	}

	argPtr := r.interfaceDataPointer((*fakeObject)(NoEscape(unsafe.Pointer(&stackAllocated))))

	ptr := uintptr(unsafe.Pointer(&stackAllocated))

	if abs(int(ptr)-int(guarateedStackValuePtr)) > int(512+unsafe.Sizeof(stackAllocated)) {
		t.Errorf("stackAllocated variable is not in stack after NoEscape: %x real addr, but should be near %x", ptr, guarateedStackValuePtr)
		return
	}

	if argPtr != ptr {
		t.Errorf("argPtr value %x mismatch with stackAllocated addr value %x, must be equal", argPtr, ptr)
	}
}

// AssignString2SliceUnsafe

// prevent from inlining and simplifying by compiler
//go:noinline
//go:nosplit
func (f *fakeInterfaceReceiver) string2slice(s string) []byte {
	return AssignString2SliceUnsafe(&s)
}

// go test -count=1 -o runtime.exe -gcflags "-m -m" -v -run "^TestAssignString2SliceUnsafe$" 2> runtime.log
// go tool objdump -S -s "go-xruntime" runtime.exe > runtime.disasm

func TestAssignString2SliceUnsafe(t *testing.T) {

	var r fakeInterface = new(fakeInterfaceReceiver)

	s := "this is test string"
	// typecast to interface{} allocate stringHeader on heap, buy reuse original underlying char array
	b := r.string2slice(s)

	if sp, bp := ((*reflect.StringHeader)(unsafe.Pointer(&s))).Data, ((*reflect.SliceHeader)(unsafe.Pointer(&b))).Data; sp != bp {
		t.Errorf("slice data pointer point to another underlying array: %x (must be %x)", bp, sp)
	}

	if len(s) != len(b) {
		t.Errorf("slice and string lengths mismatch: len(s) = %d, len(b) = %d", len(s), len(b))
	}

	if cap(b) == 0 {
		t.Errorf("slice capacity must be initialized")
	}

	if cap(b) != len(s) {
		t.Errorf("wrong slice capacity value, must be %d, got %d", len(s), cap(b))
	}
}

// AssignString2SliceUnsafeInline

// prevent from inlining and simplifying by compiler
//go:noinline
//go:nosplit
func (f *fakeInterfaceReceiver) string2sliceRough(s string) []byte {
	return AssignString2SliceUnsafeRough(&s)
}

// go test -count=1 -o runtime.exe -gcflags "-m -m" -v -run "^TestAssignString2SliceUnsafeRough" 2> runtime.log
// go tool objdump -S -s "go-xruntime" runtime.exe > runtime.disasm

func TestAssignString2SliceUnsafeRough(t *testing.T) {

	var r fakeInterface = new(fakeInterfaceReceiver)

	s := "this is test string inline"
	// typecast to interface{} allocate stringHeader on heap, buy reuse original underlying char array
	b := r.string2sliceRough(s)

	if sp, bp := ((*reflect.StringHeader)(unsafe.Pointer(&s))).Data, ((*reflect.SliceHeader)(unsafe.Pointer(&b))).Data; sp != bp {
		t.Errorf("slice data pointer point to another underlying array: %x (must be %x)", bp, sp)
	}

	if len(s) != len(b) {
		t.Errorf("slice and string lengths mismatch: len(s) = %d, len(b) = %d", len(s), len(b))
	}

	if cap(b) != 0 {
		t.Errorf("slice capacity must be 0, got %d", cap(b))
	}
}

// AssignSlice2StringUnsafe

// prevent from inlining and simplifying by compiler
//go:noinline
//go:nosplit
func (f *fakeInterfaceReceiver) slice2string(b []byte) string {
	return AssignSlice2StringUnsafe(b)
}

// go test -count=1 -o runtime.exe -gcflags "-m -m" -v -run "^TestAssignString2SliceUnsafe$" 2> runtime.log
// go tool objdump -S -s "go-xruntime" runtime.exe > runtime.disasm

func TestAssignSlice2StringUnsafe(t *testing.T) {

	var r fakeInterface = new(fakeInterfaceReceiver)

	b := []byte("this is test byte raw string")
	// typecast to interface{} allocate stringHeader on heap, buy reuse original underlying char array
	s := r.slice2string(b)

	if sp, bp := ((*reflect.StringHeader)(unsafe.Pointer(&s))).Data, ((*reflect.SliceHeader)(unsafe.Pointer(&b))).Data; sp != bp {
		t.Errorf("string data pointer point to another underlying array: %x (must be %x)", sp, bp)
	}

	if len(s) != len(b) {
		t.Errorf("slice and string lengths mismatch: len(s) = %d, len(b) = %d", len(s), len(b))
	}
}
