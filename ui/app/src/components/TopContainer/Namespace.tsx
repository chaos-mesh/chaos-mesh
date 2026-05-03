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
import { applyNSParam } from '@/api/interceptors'
import { Stale } from '@/api/queryUtils'
import Paper from '@/mui-extends/Paper'
import { useGetCommonChaosAvailableNamespaces, useGetCommonConfig } from '@/openapi'
import { useAuthStore } from '@/zustand/auth'
import { Autocomplete, TextField } from '@mui/material'
import { useEffect } from 'react'
import { useLocation, useNavigate } from 'react-router'

import i18n from '@/components/T'

const Namespace = () => {
  const navigate = useNavigate()
  const { pathname } = useLocation()

  const { data: config } = useGetCommonConfig({
    query: {
      staleTime: Stale.DAY,
    },
  })

  const tokenName = useAuthStore((state) => state.tokenName)
  const namespace = useAuthStore((state) => state.namespace)
  const setNameSpace = useAuthStore((state) => state.actions.setNameSpace)

  // Only enable the query when:
  // - Config is loaded AND
  //   - Security mode is disabled, OR
  //   - Security mode is enabled AND token is available
  const shouldFetchNamespaces = config !== undefined && (!config.security_mode || (config.security_mode && !!tokenName))

  const { data: namespaces } = useGetCommonChaosAvailableNamespaces({
    query: {
      enabled: shouldFetchNamespaces,
      staleTime: Stale.DAY,
    },
  })

  // Update namespace parameter when namespaces become available and "All" is selected
  useEffect(() => {
    if (namespace === 'All' && namespaces && namespaces.length > 0) {
      applyNSParam(namespace, namespaces)
    }
  }, [namespace, namespaces])

  const handleSelectGlobalNamespace = (_: any, newVal: any) => {
    const ns = newVal

    applyNSParam(ns, namespaces)
    setNameSpace(ns)

    navigate('/namespaceSetted', { replace: true })
    setTimeout(() => navigate(pathname, { replace: true }))
  }

  return (
    <Autocomplete
      className="tutorial-namespace"
      sx={{ minWidth: 180 }}
      value={namespace}
      options={['All', ...(namespaces || [])]}
      onChange={handleSelectGlobalNamespace}
      disableClearable={true}
      renderInput={(params) => <TextField {...params} size="small" label={i18n('common.chooseNamespace')} />}
      PaperComponent={(props) => <Paper {...props} sx={{ p: 0 }} />}
    />
  )
}

export default Namespace
