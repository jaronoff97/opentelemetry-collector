// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Code generated by "pdata/internal/cmd/pdatagen/main.go". DO NOT EDIT.
// To regenerate this file run "make genpdata".

package pcommon

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/pdata/internal"
)

func TestNewIntSlice(t *testing.T) {
	ms := NewIntSlice()
	assert.Equal(t, 0, ms.Len())
	ms.FromRaw([]int{1, 2, 3})
	assert.Equal(t, 3, ms.Len())
	assert.Equal(t, []int{1, 2, 3}, ms.AsRaw())
	ms.SetAt(1, int(5))
	assert.Equal(t, []int{1, 5, 3}, ms.AsRaw())
	ms.FromRaw([]int{3})
	assert.Equal(t, 1, ms.Len())
	assert.Equal(t, int(3), ms.At(0))

	cp := NewIntSlice()
	ms.CopyTo(cp)
	ms.SetAt(0, int(2))
	assert.Equal(t, int(2), ms.At(0))
	assert.Equal(t, int(3), cp.At(0))
	ms.CopyTo(cp)
	assert.Equal(t, int(2), cp.At(0))

	mv := NewIntSlice()
	ms.MoveTo(mv)
	assert.Equal(t, 0, ms.Len())
	assert.Equal(t, 1, mv.Len())
	assert.Equal(t, int(2), mv.At(0))
	ms.FromRaw([]int{1, 2, 3})
	ms.MoveTo(mv)
	assert.Equal(t, 3, mv.Len())
	assert.Equal(t, int(1), mv.At(0))
}

func TestIntSliceReadOnly(t *testing.T) {
	raw := []int{1, 2, 3}
	state := internal.StateReadOnly
	ms := IntSlice(internal.NewIntSlice(&raw, &state))

	assert.Equal(t, 3, ms.Len())
	assert.Equal(t, int(1), ms.At(0))
	assert.Panics(t, func() { ms.Append(1) })
	assert.Panics(t, func() { ms.EnsureCapacity(2) })
	assert.Equal(t, raw, ms.AsRaw())
	assert.Panics(t, func() { ms.FromRaw(raw) })

	ms2 := NewIntSlice()
	ms.CopyTo(ms2)
	assert.Equal(t, ms.AsRaw(), ms2.AsRaw())
	assert.Panics(t, func() { ms2.CopyTo(ms) })

	assert.Panics(t, func() { ms.MoveTo(ms2) })
	assert.Panics(t, func() { ms2.MoveTo(ms) })
}

func TestIntSliceAppend(t *testing.T) {
	ms := NewIntSlice()
	ms.FromRaw([]int{1, 2, 3})
	ms.Append(5, 5)
	assert.Equal(t, 5, ms.Len())
	assert.Equal(t, int(5), ms.At(4))
}

func TestIntSliceEnsureCapacity(t *testing.T) {
	ms := NewIntSlice()
	ms.EnsureCapacity(4)
	assert.Equal(t, 4, cap(*ms.getOrig()))
	ms.EnsureCapacity(2)
	assert.Equal(t, 4, cap(*ms.getOrig()))
}
