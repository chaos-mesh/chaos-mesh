import { setupServer } from 'msw/node'
import { getChaosMeshDashboardAPIMSW } from 'openapi/index.msw'

export const server = setupServer(...getChaosMeshDashboardAPIMSW())
