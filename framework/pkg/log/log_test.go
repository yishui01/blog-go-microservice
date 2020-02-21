package log

import (
	"strings"
	"testing"
)

func BenchmarkFileRecord(b *testing.B) {
	zapLogger := New(DefaultFileCore())

	str := strings.Repeat("Hello World!", 10)
	b.Run("file", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			zapLogger.Info(str)
		}
	})

}
