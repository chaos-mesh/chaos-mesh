// Copyright 2023 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log_test

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/log"
)

func TestZapLoggerSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zap Logger")
}

var _ = Describe("NewDefaultZapLogger", func() {
	Context("no environment", func() {
		It("uses console output", func() {
			var logTime time.Time

			output, err := interceptLogging(func() {
				logger, err := log.NewDefaultZapLogger()
				Expect(err).NotTo(HaveOccurred())

				logTime = time.Now().Truncate(time.Second)
				logger.Info("test", "key", "value123")
			})
			Expect(err).NotTo(HaveOccurred())

			fields := strings.Fields(output)

			if len(fields) < 5 {
				Expect(len(fields)).To(BeNumerically(">=", 5))
			}

			// Default is ISO8601, painfully close to but not the same as RFC3339
			ts, err := time.Parse("2006-01-02T15:04:05Z0700", fields[0])
			Expect(err).NotTo(HaveOccurred())

			Expect(ts.Before(logTime)).To(BeFalse())
			Expect(ts.After(time.Now())).To(BeFalse())

			Expect(fields[1]).To(Equal("INFO"))
			Expect(fields[2]).To(HavePrefix("log/zap_test.go:"))

			message := strings.TrimSpace(strings.Join(fields[3:], " "))
			Expect(message).To(Equal(`test {"key": "value123"}`))
		})
	})

	Context("JSON format", func() {
		It("logs valid JSON", func() {
			var logTime int64
			var output string

			wrapEnv(map[string]string{log.LogFormat: "json"}, func() {
				var err error
				output, err = interceptLogging(func() {
					logger, err := log.NewDefaultZapLogger()
					Expect(err).NotTo(HaveOccurred())

					logTime = time.Now().UnixNano()
					logger.Info("test", "key", "value123")
				})
				Expect(err).NotTo(HaveOccurred())
			})

			var data map[string]interface{}
			err := json.Unmarshal([]byte(output), &data)
			Expect(err).NotTo(HaveOccurred())

			Expect(data).To(HaveKey("ts"))
			tsSeconds := math.Floor(data["ts"].(float64))
			tsNanos := int64((data["ts"].(float64) - tsSeconds) * float64(time.Second))
			ts := int64(tsSeconds)*int64(time.Second) + tsNanos
			Expect(ts).To(BeNumerically(">", logTime))
			Expect(ts).To(BeNumerically("<", time.Now().UnixNano()))

			Expect(data).To(HaveKey("level"))
			Expect(data["level"]).To(Equal("info"))

			Expect(data).To(HaveKey("msg"))
			Expect(data["msg"]).To(Equal("test"))

			Expect(data).To(HaveKey("caller"))
			Expect(data["caller"]).To(HavePrefix("log/zap_test.go:"))

			Expect(data).To(HaveKey("key"))
			Expect(data["key"]).To(Equal("value123"))
		})

		It("uses configured timestamp key", func() {
			var logTime int64
			var output string

			wrapEnv(map[string]string{log.LogFormat: "json", log.LogKeyTimestamp: "timestamp"}, func() {
				var err error
				output, err = interceptLogging(func() {
					logger, err := log.NewDefaultZapLogger()
					Expect(err).NotTo(HaveOccurred())

					logTime = time.Now().UnixNano()
					logger.Info("test", "key", "value123")
				})
				Expect(err).NotTo(HaveOccurred())
			})

			var data map[string]interface{}
			err := json.Unmarshal([]byte(output), &data)
			Expect(err).NotTo(HaveOccurred())

			Expect(data).To(HaveKey("timestamp"))
			tsSeconds := math.Floor(data["timestamp"].(float64))
			tsNanos := int64((data["timestamp"].(float64) - tsSeconds) * float64(time.Second))
			ts := int64(tsSeconds)*int64(time.Second) + tsNanos
			Expect(ts).To(BeNumerically(">=", logTime))
			Expect(ts).To(BeNumerically("<", time.Now().UnixNano()))
		})

		It("uses configured message key", func() {
			var output string

			wrapEnv(map[string]string{log.LogFormat: "json", log.LogKeyMessage: "message"}, func() {
				var err error
				output, err = interceptLogging(func() {
					logger, err := log.NewDefaultZapLogger()
					Expect(err).NotTo(HaveOccurred())

					logger.Info("test", "key", "value123")
				})
				Expect(err).NotTo(HaveOccurred())
			})

			var data map[string]interface{}
			err := json.Unmarshal([]byte(output), &data)
			Expect(err).NotTo(HaveOccurred())

			Expect(data).To(HaveKey("message"))
			Expect(data["message"]).To(Equal("test"))
		})

		It("uses configured timestamp format", func() {
			var logTime time.Time
			var output string

			wrapEnv(map[string]string{log.LogFormat: "json", log.LogTimestampFormat: "rfc3339nano"}, func() {
				var err error
				output, err = interceptLogging(func() {
					logger, err := log.NewDefaultZapLogger()
					Expect(err).NotTo(HaveOccurred())

					logTime = time.Now()
					logger.Info("test", "key", "value123")
				})
				Expect(err).NotTo(HaveOccurred())
			})

			var data map[string]interface{}
			err := json.Unmarshal([]byte(output), &data)
			Expect(err).NotTo(HaveOccurred())

			Expect(data).To(HaveKey("ts"))
			ts, err := time.Parse(time.RFC3339Nano, data["ts"].(string))
			Expect(err).NotTo(HaveOccurred())

			Expect(ts.After(logTime)).To(BeTrue())
			Expect(ts.Before(time.Now())).To(BeTrue())
		})

		It("respects LOG_LEVEL", func() {
			var output string

			wrapEnv(map[string]string{log.LogLevel: "error"}, func() {
				var err error
				output, err = interceptLogging(func() {
					logger, err := log.NewDefaultZapLogger()
					Expect(err).NotTo(HaveOccurred())

					logger.Info("test", "key", "value123")
				})
				Expect(err).NotTo(HaveOccurred())
			})

			Expect(output).To(Equal(""))

			wrapEnv(map[string]string{log.LogLevel: "info"}, func() {
				var err error
				output, err = interceptLogging(func() {
					logger, err := log.NewDefaultZapLogger()
					Expect(err).NotTo(HaveOccurred())

					logger.Info("test", "key", "value123")
				})
				Expect(err).NotTo(HaveOccurred())
			})

			Expect(output).NotTo(Equal(""))
		})

		It("logs interface{} objects", func() {
			var output string

			wrapEnv(map[string]string{log.LogFormat: "json"}, func() {
				var err error
				output, err = interceptLogging(func() {
					logger, err := log.NewDefaultZapLogger()
					Expect(err).NotTo(HaveOccurred())

					type testStruct struct {
						String string
						Int    int
					}

					logger.Info("test", "obj", testStruct{String: "hello!", Int: 42})
				})
				Expect(err).NotTo(HaveOccurred())
			})

			var data map[string]interface{}
			err := json.Unmarshal([]byte(output), &data)
			Expect(err).NotTo(HaveOccurred())

			Expect(data).To(HaveKey("obj"))
			Expect(data["obj"]).To(Equal(map[string]interface{}{"String": "hello!", "Int": float64(42)}))
		})

		It("logs interface{} objects as JSON string with truncation enabled", func() {
			var output string

			wrapEnv(map[string]string{log.LogFormat: "json", log.LogMaxFieldSize: "100"}, func() {
				var err error
				output, err = interceptLogging(func() {
					logger, err := log.NewDefaultZapLogger()
					Expect(err).NotTo(HaveOccurred())

					type testStruct struct {
						String string
						Int    int
					}

					logger.Info("test", "obj", testStruct{String: "hello!", Int: 42})
				})
				Expect(err).NotTo(HaveOccurred())
			})

			var data map[string]interface{}
			err := json.Unmarshal([]byte(output), &data)
			Expect(err).NotTo(HaveOccurred())

			Expect(data).To(HaveKey("obj"))
			Expect(data["obj"]).To(Equal(`{"String":"hello!","Int":42}`))
		})

		It("truncates large logged objects", func() {
			var output string

			wrapEnv(map[string]string{log.LogFormat: "json", log.LogMaxFieldSize: "100"}, func() {
				var err error
				output, err = interceptLogging(func() {
					logger, err := log.NewDefaultZapLogger()
					Expect(err).NotTo(HaveOccurred())

					type testStruct struct {
						String string
						Int    int
					}
					object := []testStruct{}
					for i := 0; i < 20; i++ {
						object = append(object, testStruct{String: fmt.Sprintf("struct #%d", i), Int: i})
					}

					logger.Info("test", "obj", object)
				})
				Expect(err).NotTo(HaveOccurred())
			})

			var data map[string]interface{}
			err := json.Unmarshal([]byte(output), &data)
			Expect(err).NotTo(HaveOccurred())

			Expect(data).To(HaveKey("obj"))
			Expect(data["obj"]).To(HaveLen(110))
			Expect(data["obj"]).To(HavePrefix("TRUNCATED "))
		})
	})
})

func interceptLogging(f func()) (string, error) {
	stderr := os.Stderr

	r, w, err := os.Pipe()
	if err != nil {
		return "", fmt.Errorf("pipe creation: %w", err)
	}
	os.Stderr = w

	f()

	os.Stderr = stderr
	w.Close()

	output, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read from pipe: %w", err)
	}

	return string(output), nil
}

func wrapEnv(values map[string]string, f func()) {
	saved := map[string]string{}

	for k, v := range values {
		existing, set := os.LookupEnv(k)
		if set {
			saved[k] = existing
		}

		os.Setenv(k, v)
	}
	defer func() {
		for k := range values {
			if existing, ok := saved[k]; ok {
				os.Setenv(k, existing)
			} else {
				os.Unsetenv(k)
			}
		}
	}()

	f()

}
