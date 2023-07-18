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
import Archives from 'pages/Archives'
import Archive from 'pages/Archives/Single'
import Dashboard from 'pages/Dashboard'
import Events from 'pages/Events'
import Experiments from 'pages/Experiments'
import NewExperiment from 'pages/Experiments/New'
import Experiment from 'pages/Experiments/Single'
import Schedules from 'pages/Schedules'
import NewSchedule from 'pages/Schedules/New'
import Schedule from 'pages/Schedules/Single'
import Settings from 'pages/Settings'
import Workflows from 'pages/Workflows'
import Workflow from 'pages/Workflows/Single'
import { Navigate, createHashRouter } from 'react-router-dom'

import NewWorkflow from 'components/NewWorkflow'
import NewWorkflowNext from 'components/NewWorkflowNext'
import TopContainer from 'components/TopContainer'

const router = createHashRouter([
  {
    path: '/',
    Component: TopContainer,
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: 'dashboard',
        Component: Dashboard,
      },
      {
        path: 'workflows/new',
        Component: NewWorkflow,
      },
      {
        path: 'workflows/new/next',
        Component: NewWorkflowNext,
      },
      {
        path: 'workflows',
        Component: Workflows,
      },
      {
        path: 'workflows/:uuid',
        Component: Workflow,
      },
      {
        path: 'schedules/new',
        Component: NewSchedule,
      },
      {
        path: 'schedules',
        Component: Schedules,
      },
      {
        path: 'schedules/:uuid',
        Component: Schedule,
      },
      {
        path: 'experiments/new',
        Component: NewExperiment,
      },
      {
        path: 'experiments',
        Component: Experiments,
      },
      {
        path: 'experiments/:uuid',
        Component: Experiment,
      },
      {
        path: 'events',
        Component: Events,
      },
      {
        path: 'archives',
        Component: Archives,
      },
      {
        path: 'archives/:uuid',
        Component: Archive,
      },
      {
        path: 'settings',
        Component: Settings,
      },
    ],
  },
])

export default router
