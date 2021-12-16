/*
 * Copyright 2021 Chaos Mesh Authors.
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
import * as Yup from 'yup'

export const schemaBasic = Yup.object({
  name: Yup.string().trim().required('The task name is required'),
  deadline: Yup.string().trim().required('The deadline is required'),
})

export const schema = schemaBasic.shape({
  container: Yup.object({
    name: Yup.string().trim().required('The container name is required'),
    image: Yup.string().trim().required('The image is required'),
    command: Yup.array().of(Yup.string()),
  }),
  conditionalBranches: Yup.array()
    .of(
      Yup.object({
        target: Yup.string().trim().required('The target is required'),
        expression: Yup.string().trim().required('The expression is required'),
      })
    )
    .min(1)
    .required('The conditional branches should be defined'),
})
