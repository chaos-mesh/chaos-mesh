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
import * as archives from './archives'
import * as auth from './auth'
import * as common from './common'
import * as events from './events'
import * as experiments from './experiments'
import * as schedules from './schedules'
import * as workflows from './workflows'

const api = {
  auth,
  common,
  experiments,
  workflows,
  schedules,
  events,
  archives,
}

export default api
