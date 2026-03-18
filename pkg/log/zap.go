// Copyright 2022 Chaos Mesh Authors.
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
//

package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// Environment variable to control log format, if set to JSON, structured logging will be used. Default is "console", and to use zap.NewDevelopmentConfig.
	LogFormat = "LOG_FORMAT"
	// Log format environment variable value indicating JSON
	LogFormatJSON = "json"
	// Log format environment variable value indicating console / human-readable form
	LogFormatConsole = "console"

	// Environment variable to control logging level, should parse to zapcore.Level. Default is "info".
	LogLevel = "LOG_LEVEL"

	// Environment variable to control logging key used for log message in structured logging. Default is defined by zap.NewProductionEncoderConfig.
	LogKeyMessage = "LOG_KEY_MESSAGE"

	// Environment variable to control logging key used for log timestamp in structured logging. Default is defined by zap.NewProductionEncoderConfig.
	LogKeyTimestamp = "LOG_KEY_TIMESTAMP"

	// Environment variable to control the maximum field size in structured logging before truncating the field value.
	LogMaxFieldSize = "LOG_MAX_FIELD_SIZE"

	// Environment variable to control the logging timestamp format used in structured logging. Valid values are "rfc3339", "rfc3339nano" and "epoch".
	LogTimestampFormat = "LOG_TIMESTAMP_FORMAT"
)

// NewDefaultZapLogger is the recommended way to create a new logger, you could call this function to initialize the root
// logger of your application, and provide it to your components, by fx or manually.
func NewDefaultZapLogger() (logr.Logger, error) {
	logLevel := zap.InfoLevel
	if envLevel := os.Getenv(LogLevel); envLevel != "" {
		if level, err := zapcore.ParseLevel(envLevel); err == nil {
			logLevel = level
		} else {
			fmt.Printf("invalid log level %q, falling back to default level %q\n", envLevel, logLevel)
		}
	}

	logFormat := LogFormatConsole
	if envFormat := os.Getenv(LogFormat); envFormat == LogFormatJSON {
		logFormat = LogFormatJSON
	}

	var config zap.Config

	if logFormat == LogFormatConsole {
		config = zap.NewDevelopmentConfig()
	} else {
		encoderConfig := zap.NewProductionEncoderConfig()

		if v := os.Getenv(LogKeyMessage); v != "" {
			encoderConfig.MessageKey = v
		}
		if v := os.Getenv(LogKeyTimestamp); v != "" {
			encoderConfig.TimeKey = v
		}

		if v := os.Getenv(LogTimestampFormat); v != "" {
			var timeEncoder zapcore.TimeEncoder

			if err := timeEncoder.UnmarshalText([]byte(v)); err == nil {
				encoderConfig.EncodeTime = timeEncoder
			} else {
				fmt.Printf("invalid timestamp format %q, falling back to default: %s\n", v, err)
			}

		}

		// If configured, truncate the fields to the configured size. This allows for reasonable configuration to prevent extremely
		// long logging lines (e.g. by logging an object with a large collection of events attached), which can cause issues with
		// log ingestion/collection, while guaranteeing that the output remains valid JSON.
		envMaxFieldSize := os.Getenv(LogMaxFieldSize)
		if envMaxFieldSize != "" {
			maxFieldSize, err := strconv.Atoi(envMaxFieldSize)
			if err == nil {
				encoderConfig.NewReflectedEncoder = func(w io.Writer) zapcore.ReflectedEncoder {
					enc := json.NewEncoder(newTruncatingWriter(w, maxFieldSize))
					enc.SetEscapeHTML(false)
					return enc
				}
			}
		}

		config = zap.NewProductionConfig()
		config.EncoderConfig = encoderConfig
	}

	zapLogger, err := config.Build(zap.IncreaseLevel(logLevel))
	if err != nil {
		return logr.Discard(), err
	}

	logger := zapr.NewLogger(zapLogger)
	return logger, nil
}

type truncatingWriter struct {
	writer    io.Writer
	enc       *json.Encoder
	maxLength int
}

func newTruncatingWriter(writer io.Writer, maxLength int) truncatingWriter {
	enc := json.NewEncoder(writer)
	enc.SetEscapeHTML(false)

	return truncatingWriter{
		writer:    writer,
		maxLength: maxLength,
		enc:       enc,
	}
}

func (tr truncatingWriter) Write(bytes []byte) (int, error) {
	output := bytes

	if len(bytes) > tr.maxLength {
		output = append([]byte("TRUNCATED "), bytes[:tr.maxLength]...)
	}

	// Always encode the field value as a JSON string, this ensures that the log field types will not change between logging
	// of truncated and untruncated values, while also ensuring that the log line remains valid JSON.
	return 0, tr.enc.Encode(strings.TrimRight(string(output), "\n"))
}

// NewZapLoggerWithWriter creates a new logger with io.writer
// The provided encoder presets NewDevelopmentEncoderConfig used by NewDevelopmentConfig do not enable function name logging.
// To enable function name, a non-empty value for config.EncoderConfig.FunctionKey.
func NewZapLoggerWithWriter(out io.Writer) logr.Logger {
	bWriter := out
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.FunctionKey = "function"
	core := zapcore.NewCore(zapcore.NewJSONEncoder(config.EncoderConfig), zapcore.AddSync(bWriter), config.Level)
	zapLogger := zap.New(core)
	logger := zapr.NewLogger(zapLogger)
	return logger
}
