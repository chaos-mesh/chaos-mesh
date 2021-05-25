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
