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
import { RouteProps } from 'react-router-dom'

import NewWorkflow from 'components/NewWorkflow'
import NewWorkflowNext from 'components/NewWorkflowNext'

type CustomRouteProps = RouteProps & { title: string }

const routes: CustomRouteProps[] = [
  {
    element: <Dashboard />,
    path: '/dashboard',
    title: 'Dashboard',
  },
  {
    element: <NewWorkflow />,
    path: '/workflows/new',
    title: 'New Workflow',
  },
  {
    element: <NewWorkflowNext />,
    path: '/workflows/new/next',
    title: 'New Workflow',
  },
  {
    element: <Workflows />,
    path: '/workflows',
    title: 'Workflows',
  },
  {
    element: <Workflow />,
    path: '/workflows/:uuid',
    title: 'Workflow',
  },
  {
    element: <NewSchedule />,
    path: '/schedules/new',
    title: 'New Schedule',
  },
  {
    element: <Schedules />,
    path: '/schedules',
    title: 'Schedules',
  },
  {
    element: <Schedule />,
    path: '/schedules/:uuid',
    title: 'Schedule',
  },
  {
    element: <NewExperiment />,
    path: '/experiments/new',
    title: 'New Experiment',
  },
  {
    element: <Experiments />,
    path: '/experiments',
    title: 'Experiments',
  },
  {
    element: <Experiment />,
    path: '/experiments/:uuid',
    title: 'Experiment',
  },
  {
    element: <Events />,
    path: '/events',
    title: 'Events',
  },
  {
    element: <Archives />,
    path: '/archives',
    title: 'Archives',
  },
  {
    element: <Archive />,
    path: '/archives/:uuid',
    title: 'Archive',
  },
  {
    element: <Settings />,
    path: '/settings',
    title: 'Settings',
  },
]

export default routes
