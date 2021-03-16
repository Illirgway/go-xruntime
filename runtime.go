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
	"unsafe"
)

// ATN! only for const, global, stack allocated, pooled and fixed from gc in calling fn (the most difficult case)
//      input strings/slices !!!

// ATN! SEE https://github.com/golang/go/issues/25484#issuecomment-391329481

// SEE go api runtime slicebytetostringtmp
//
// NOTE go compiler generate shortest code for `*string` arg type (at least for go1.13)
//go:nosplit
func AssignString2SliceUnsafe( /* const */ sp *string) ( /* const */ b []byte) {

	/*
		pStringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
		pSliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))

		pSliceHeader.Data = pStringHeader.Data
		pSliceHeader.Len, pSliceHeader.Cap = pStringHeader.Len, pStringHeader.Len
	*/

	// more GC stable code variant - b now GC driven, so s may be heap allocated string
	// + with fixing Cap of ret slice b
	// https://stackoverflow.com/a/59210739

	b = *(*[]byte)(unsafe.Pointer(sp))
	(*reflect.SliceHeader)(unsafe.Pointer(&b)).Cap = len(*sp)

	return b
}

// shorter and better for inlining version of AssignString2SliceUnsafe, but without Cap explicit initialization
// (cap(result) is undefined garbage)
// should be used with additional care
//
// NOTE go compiler generate shortest code for `*string` arg type (at least for go1.13)
//go:nosplit
func AssignString2SliceUnsafeRough( /* const */ sp *string) /* const */ []byte {
	return *(*[]byte)(unsafe.Pointer(sp))
}

// SEE https://github.com/golang/go/issues/25484#issuecomment-391401861

//go:nosplit
func AssignSlice2StringUnsafe( /* const */ b []byte) ( /* const */ s string) {

	/*
		pStringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
		pSliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))

		pStringHeader.Data = pSliceHeader.Data
		pStringHeader.Len = pSliceHeader.Len

		return s
	*/

	// inlined
	// using of this form leads to GC driven ret string s, so b may be heap allocated slice
	return *(*string)(unsafe.Pointer(&b))
}

// from runtime/string.go:stringStructOf
// inlined
//go:nosplit
func GetStringHeader(sp *string) *reflect.StringHeader {
	return (*reflect.StringHeader)(unsafe.Pointer(sp))
}

// inlined
//go:nosplit
func GetStringDataPointer(sp *string) uintptr {
	return (*reflect.StringHeader)(unsafe.Pointer(sp)).Data
}

// Hides pointer from the compiler escape analysis
// The most dangerous function in file, SHOULD BE USED WITH EXTREME CAREFULLY
//
// In fact it is copy of go src\runtime\stubs.go:noescape() under exported name
//
// USAGE (*TypeT)(xruntime.NoEscape(unsafe.Pointer(&valueOfTypeT)))
// NOTE can't simply use go:linkname due to loss of inlining
//
// inlined
//go:nosplit
//~go:noescape
func NoEscape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
}
