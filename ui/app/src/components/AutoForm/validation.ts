/*
 * Copyright 2022 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import { array, object, string } from 'yup'

import { Belong } from '.'

const scopeInitialValuesSchema = {
  selector: object({
    namespaces: array().min(1),
  }),
}

const workflowNodeInfoSchema = {
  name: string().trim().required(),
  deadline: string().trim().required(),
}

const workflowSchema = object({
  ...scopeInitialValuesSchema,
  ...workflowNodeInfoSchema,
})

export function chooseSchemaByBelong(belong: Belong) {
  switch (belong) {
    case Belong.Workflow:
      return workflowSchema
  }
}
