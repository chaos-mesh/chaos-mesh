import { getChaosMeshDashboardAPIMSW } from '@/openapi/index.msw'
import { setupServer } from 'msw/node'

export const server = setupServer(...getChaosMeshDashboardAPIMSW())
