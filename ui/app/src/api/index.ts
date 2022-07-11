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
import autoBind from 'auto-bind'
import { ArchivesApi, CommonApi, EventsApi, ExperimentsApi, SchedulesApi, WorkflowsApi } from 'openapi'

import * as auth from './auth'
import http from './http'

/**
 * Due to we use OpenAPI generated as our API, it is important to make sure that `this` always points to
 * the original instance object during the API call.
 *
 * So we use `autoBind` to help us finish the job automatically. Please read the following code for more details.
 *
 */

class CommonApiBind extends CommonApi {
  constructor(...args: any) {
    super(...args)
    autoBind(this)
  }
}

class ArchivesApiBind extends ArchivesApi {
  constructor(...args: any) {
    super(...args)
    autoBind(this)
  }
}

class ExperimentsApiBind extends ExperimentsApi {
  constructor(...args: any) {
    super(...args)
    autoBind(this)
  }
}

class SchedulesApiBind extends SchedulesApi {
  constructor(...args: any) {
    super(...args)
    autoBind(this)
  }
}

class WorkflowsApiBind extends WorkflowsApi {
  constructor(...args: any) {
    super(...args)
    autoBind(this)
  }
}

class EventsApiBind extends EventsApi {
  constructor(...args: any) {
    super(...args)
    autoBind(this)
  }
}

const common = new CommonApiBind(undefined, undefined, http)
const archives = new ArchivesApiBind(undefined, undefined, http)
const experiments = new ExperimentsApiBind(undefined, undefined, http)
const schedules = new SchedulesApiBind(undefined, undefined, http)
const workflows = new WorkflowsApiBind(undefined, undefined, http)
const events = new EventsApiBind(undefined, undefined, http)

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
