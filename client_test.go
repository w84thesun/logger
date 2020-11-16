package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_mapToSlice(t *testing.T) {
	type args struct {
		fields Fields
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "basic",
			args: args{fields: map[string]interface{}{"1": 1}},
			want: []interface{}{"1", 1},
		},
		{
			name: "ignore fields",
			args: args{fields: map[string]interface{}{
				"1":          1,
				"namespace":  "awesome",
				"@timestamp": "kill me",
				"message":    "kill me",
				"level":      "kill me",
				"service":    "kill me",
			}},
			want: []interface{}{"1", 1, "namespace", "awesome"},
		},
		{
			name: "no accepted fields",
			args: args{fields: map[string]interface{}{
				"@timestamp": "kill me",
				"message":    "kill me",
				"level":      "kill me",
				"service":    "kill me",
			}},
			want: []interface{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.fields.Flatten()
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestLoggerImpl_With(t *testing.T) {
	logger, _ := New(LoggingConfig{
		Service:   "testing",
		Namespace: "default",
	})

	logger.Info("okay")

	logger.Namespace("custom").With(Fields{"hello": "world"}).Info("modified")

	logger.Info("should be clear")
}

func BenchmarkLoggerImpl_Info(b *testing.B) {
	logger, _ := New(LoggingConfig{
		Service:       "testing",
		Namespace:     "default",
		DisableStdout: true,
		Level:         "info",
	})

	b.ReportAllocs()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Namespace("test").With(Fields{"a": "b"}).Info("hello there")
	}
}

func BenchmarkLoggerImpl_Error(b *testing.B) {
	logger, _ := New(DefaultConfig)

	b.ReportAllocs()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Error("test")
	}
}

func BenchmarkLoggerImpl_Errorf(b *testing.B) {
	logger, _ := New(DefaultConfig)

	b.ReportAllocs()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Errorf("test")
	}
}

func BenchmarkLoggerImpl_Errorf2(b *testing.B) {
	logger, _ := New(DefaultConfig)

	b.ReportAllocs()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Errorf("test: %s", "test")
	}
}
