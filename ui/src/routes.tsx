import { RouteProps } from 'react-router'

import Overview from './pages/Overview'
import Experiments from './pages/Experiments'
import ExperimentDetail from './pages/Experiments/Detail'
import Events from './pages/Events'
import Archive from './pages/Archive'

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
    path: '/experiment/:name',
    exact: true,
  },
  {
    component: Events,
    path: '/events',
    exact: true,
  },
  {
    component: Archive,
    path: '/archive',
    exact: true,
  },
]
