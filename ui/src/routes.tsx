import ArchiveReport from 'pages/ArchiveReport'
import Archives from 'pages/Archives'
import Dashboard from 'pages/Dashboard'
import Events from 'pages/Events'
import ExperimentDetail from 'pages/ExperimentDetail'
import Experiments from 'pages/Experiments'
import NewExperiment from 'pages/Experiments/New'
import NewWorkflow from 'components/NewWorkflow'
import { RouteProps } from 'react-router'
import Settings from 'pages/Settings'
import Swagger from 'pages/Swagger'
import Workflows from 'pages/Workflows'

const routes: RouteProps[] = [
  {
    component: Dashboard,
    path: '/dashboard',
    exact: true,
  },
  {
    component: Experiments,
    path: '/experiments',
    exact: true,
  },
  {
    component: NewExperiment,
    path: '/experiments/new',
  },
  {
    component: ExperimentDetail,
    path: '/experiments/:uuid',
  },
  {
    component: Workflows,
    path: '/workflows',
    exact: true,
  },
  {
    component: NewWorkflow,
    path: '/workflows/new',
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
    component: ArchiveReport,
    path: '/archives/:uuid',
  },
  {
    component: Settings,
    path: '/settings',
    exact: true,
  },
  {
    component: Swagger,
    path: '/swagger',
    exact: true,
  },
]

export default routes
