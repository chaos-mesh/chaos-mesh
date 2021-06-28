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
