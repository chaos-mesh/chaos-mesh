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
	"fmt"
	"go.uber.org/zap/zapcore"
	"bufio"
	"bytes"
	
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

// NewDefaultZapLogger is the recommended way to create a new logger, you could call this function to initialize the root
// logger of your application, and provide it to your components, by fx or manually.
func NewDefaultZapLogger() (logr.Logger, error) {
	// change the configuration in the future if needed.
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return logr.Discard(), err
	}
	logger := zapr.NewLogger(zapLogger)
	return logger, nil
}

func NewZapLoggerWithWriter() logr.Logger {

	var b bytes.Buffer
	bWriter := bufio.NewWriter(&b)
	
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.FunctionKey = "function"
	
	core := zapcore.NewCore(zapcore.NewJSONEncoder(config.EncoderConfig), zapcore.AddSync(bWriter), config.Level)
	zapLogger := zap.New(core)
	zapLogger.Error("an error")
	logger := zapr.NewLogger(zapLogger)
	bWriter.Flush()
	fmt.Println(b.String())
	return logger
	}
