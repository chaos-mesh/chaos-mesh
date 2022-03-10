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

package extension

//
//import (
//	"fmt"
//
//	"github.com/pkg/errors"
//
//	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
//	"github.com/chaos-mesh/chaos-mesh/pkg/time"
//)
//
//type Config struct {
//	inner time.Config
//}
//
//func (c *Config) DeepCopy() tasks.Object {
//	return &Config{*c.inner.DeepCopy().(*time.Config)}
//}
//
//func (c *Config) Add(a tasks.Addable) error {
//	A, OK := a.(*Config)
//	if OK {
//		err := c.inner.Add(&A.inner)
//		if err != nil {
//			return err
//		}
//		return nil
//	}
//
//	return errors.Wrapf(tasks.ErrCanNotAdd, "expect type : *extension.Config, got : %T", a)
//}
//
//func (c *Config) New(values interface{}) (tasks.Injectable, error) {
//	skew, err := time.GetSkew()
//	if err != nil {
//		return nil, err
//	}
//	skew.SkewConfig = *c.inner.DeepCopy().(*time.Config)
//	groupProcessHandler, ok := values.(*tasks.ProcessGroupHandler)
//	if !ok {
//		return nil, errors.New(fmt.Sprintf("type %t is not *tasks.ProcessGroupHandler", values))
//	}
//	_, ok = groupProcessHandler.Main.(*time.Skew)
//	if !ok {
//		return nil, errors.New(fmt.Sprintf("type %t is not *Skew", groupProcessHandler.Main))
//	}
//	newGroupProcessHandler :=
//		tasks.NewProcessGroupHandler(groupProcessHandler.Logger, &skew)
//	return &newGroupProcessHandler, nil
//}
//
//func (c *Config) Assign(injectable tasks.Injectable) error {
//	groupProcessHandler, ok := injectable.(*tasks.ProcessGroupHandler)
//	if !ok {
//		return errors.New(fmt.Sprintf("type %t is not *tasks.ProcessGroupHandler", injectable))
//	}
//	I, ok := groupProcessHandler.Main.(*time.Skew)
//	if !ok {
//		return errors.New(fmt.Sprintf("type %t is not *Skew", groupProcessHandler.Main))
//	}
//
//	I.SkewConfig = (*c).inner
//	return nil
//}
