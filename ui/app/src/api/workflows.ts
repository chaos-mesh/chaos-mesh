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
import { TemplateCustom } from '@/slices/workflows'

import http from './http'
import { RequestForm } from './workflows.type'

// TODO: refactor this interface, use the union type from golang struct
export interface APITemplate {
  name: string
  templateType: string
  deadline?: string
  children?: APITemplate[]
  task?: TemplateCustom
}

export const renderHTTPTask = (form: RequestForm) => http.post('/workflows/render-task/http', form)
export const parseHTTPTask = (t: APITemplate) => http.post('/workflows/parse-task/http', t)
