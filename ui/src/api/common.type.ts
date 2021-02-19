export interface Config {
  security_mode: boolean
  dns_server_create: boolean
  version: string
}

export interface RBACConfigParams {
  namespace: string
  role: 'manager' | 'viewer'
}
