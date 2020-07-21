import Archives from './pages/Archives'
import Events from './pages/Events'
import ExperimentDetail from './pages/ExperimentDetail'
import Experiments from './pages/Experiments'
import NewExperiment from 'components/NewExperiment'
import Overview from './pages/Overview'
import { RouteProps } from 'react-router'

const routes: RouteProps[] = [
  {
    component: Overview,
    path: '/overview',
    exact: true,
  },
  {
    component: NewExperiment,
    path: '/newExperiment',
  },
  {
    component: Experiments,
    path: '/experiments',
    exact: true,
  },
  {
    component: ExperimentDetail,
    path: '/experiments/:name',
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
]

export default routes
