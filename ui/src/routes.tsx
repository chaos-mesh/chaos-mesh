import { RouteProps } from 'react-router'

import Overview from './pages/Overview'
import Experiments from './pages/Experiments'
import ExperimentDetail from './pages/Experiments/Detail'
import NewExperiment from './pages/Experiments/New'
import Events from './pages/Events'
import Archives from './pages/Archives'

export const routes: RouteProps[] = [
  {
    component: Overview,
    path: '/overview',
    exact: true,
  },
  {
    component: Experiments,
    path: '/experiments',
    exact: true,
  },
  {
    component: ExperimentDetail,
    path: '/experiments/:name',
    exact: true,
  },
  {
    component: NewExperiment,
    path: '/new-experiment',
    exact: true,
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
