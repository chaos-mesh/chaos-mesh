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
import Archive from 'pages/Archives/Single'
import Archives from 'pages/Archives'
import Dashboard from 'pages/Dashboard'
import Events from 'pages/Events'
import Experiment from 'pages/Experiments/Single'
import Experiments from 'pages/Experiments'
import NewExperiment from 'pages/Experiments/New'
import NewSchedule from 'pages/Schedules/New'
import NewWorkflow from 'components/NewWorkflow'
import { RouteProps } from 'react-router'
import Schedule from 'pages/Schedules/Single'
import Schedules from 'pages/Schedules'
import Settings from 'pages/Settings'
import Workflow from 'pages/Workflows/Single'
import Workflows from 'pages/Workflows'

const routes: RouteProps[] = [
  {
    component: Dashboard,
    path: '/dashboard',
    exact: true,
  },
  {
    component: NewWorkflow,
    path: '/workflows/new',
  },
  {
    component: Workflows,
    path: '/workflows',
    exact: true,
  },
  {
    component: Workflow,
    path: '/workflows/:uuid',
  },
  {
    component: NewSchedule,
    path: '/schedules/new',
  },
  {
    component: Schedules,
    path: '/schedules',
    exact: true,
  },
  {
    component: Schedule,
    path: '/schedules/:uuid',
  },
  {
    component: NewExperiment,
    path: '/experiments/new',
  },
  {
    component: Experiments,
    path: '/experiments',
    exact: true,
  },
  {
    component: Experiment,
    path: '/experiments/:uuid',
  },
  {
    component: Events,
    path: '/events',
    exact: true,
  },
  {
    component: Archives,
    path: '/archives',
    exact: true,
  },
  {
    component: Archive,
    path: '/archives/:uuid',
  },
  {
    component: Settings,
    path: '/settings',
    exact: true,
  },
]

export default routes
