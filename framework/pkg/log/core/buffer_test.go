// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package core

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

//突然发现测试还是有点用的，一下就测出哪里有bug，表示很爽【滑稽】
func TestBufferWrites(t *testing.T) {
	buf := NewPool(0).Get()
	tests := []struct {
		desc string
		f    func()
		want string
	}{
		{"AppendByte", func() { buf.AppendByte('v') }, "v"},
		{"AppendString", func() { buf.AppendString("foo") }, "foo"},
		{"AppendIntPositive", func() { buf.AppendInt(42) }, "42"},
		{"AppendIntNegative", func() { buf.AppendInt(-42) }, "-42"},
		{"AppendUint", func() { buf.AppendUint(42) }, "42"},
		{"AppendBool", func() { buf.AppendBool(true) }, "true"},
		{"AppendFloat64", func() { buf.AppendFloat(3.14, 64) }, "3.14"},
		{"AppendFloat32", func() { buf.AppendFloat(float64(float32(3.14)), 32) }, "3.14"},
		{"AppendWrite", func() { buf.Write([]byte("foo")) }, "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			buf.Reset()
			tt.f()
			assert.Equal(t, tt.want, buf.String(), "Unexpected buffer.String().")
			assert.Equal(t, tt.want, string(buf.Bytes()), "Unexpected string(buffer.Bytes()).")
			assert.Equal(t, len(tt.want), buf.Len(), "Unexpected buffer length.")
			// We're not writing more than a kibibyte in tests.
			assert.Equal(t, _size, buf.Cap(), "Expected buffer capacity to remain constant.")
		})
	}
}

//直接操作byte数组
//BenchmarkBuffers/ByteSlice              47869220                23.3 ns/op
//BenchmarkBuffers/ByteSlice-2            58386488                22.1 ns/op
//BenchmarkBuffers/ByteSlice-4            58312720                21.5 ns/op
//BenchmarkBuffers/ByteSlice-8            53119909                21.6 ns/op

//标准库bytes.Buffer
//BenchmarkBuffers/BytesBuffer            44326240                27.4 ns/op
//BenchmarkBuffers/BytesBuffer-2          43723957                27.9 ns/op
//BenchmarkBuffers/BytesBuffer-4          43525412                27.2 ns/op
//BenchmarkBuffers/BytesBuffer-8          43477472                27.4 ns/op

//自定义结构操作，底层是用 strconv.AppendXXX对byte数组进行操作的
//BenchmarkBuffers/CustomBuffer           53194318                22.6 ns/op
//BenchmarkBuffers/CustomBuffer-2         52060285                23.5 ns/op
//BenchmarkBuffers/CustomBuffer-4         52032519                23.0 ns/op
//BenchmarkBuffers/CustomBuffer-8         53200686                22.9 ns/op
//PASS
//ok      github.com/zuiqiangqishao/framework/pkg/log/core        16.073s

//go test -bench BenchmarkBuffers  -run =^$ -cpu 1,2,4,8
func BenchmarkBuffers(b *testing.B) {
	//因为我们使用strconv.AppendFoo系列函数来对buffer进行读写非常自由，
	//所以我们不会使用标准库的bytes.Buffer(而不引起一堆额外的分配），
	//不过，让我们确保我们不会损失任何纳秒
	// Because we use the strconv.AppendFoo functions so liberally, we can't
	// use the standard library's bytes.Buffer anyways (without incurring a
	// bunch of extra allocations). Nevertheless, let's make sure that we're
	// not losing any precious nanoseconds.
	str := strings.Repeat("A", 1024)
	s := []byte(str)
	slice := make([]byte, 1024)
	buf := bytes.NewBuffer(slice)
	custom := NewPool(0).Get()

	b.Run("ByteSlice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			slice = append(slice, str...)
			slice = slice[:0]
		}
	})
	b.Run("BytesBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			//buf.WriteString(str)
			buf.Write(s)
			buf.Reset()
		}
	})
	b.Run("CustomBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			custom.Write(s)
			//custom.AppendString(str)
			custom.Reset()
		}
	})
}
