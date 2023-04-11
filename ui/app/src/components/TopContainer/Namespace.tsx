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
import { Autocomplete, TextField } from '@mui/material'
import { applyNSParam } from 'api/interceptors'
import { Stale } from 'api/queryUtils'
import { useGetCommonChaosAvailableNamespaces } from 'openapi'
import { useLocation, useNavigate } from 'react-router-dom'

import Paper from '@ui/mui-extends/esm/Paper'

import { useStoreDispatch, useStoreSelector } from 'store'

import { setNameSpace } from 'slices/globalStatus'

import i18n from 'components/T'

const Namespace = () => {
  const navigate = useNavigate()
  const { pathname } = useLocation()

  const { data: namespaces } = useGetCommonChaosAvailableNamespaces({
    query: {
      staleTime: Stale.DAY,
    },
  })

  const { namespace } = useStoreSelector((state) => state.globalStatus)
  const dispatch = useStoreDispatch()

  const handleSelectGlobalNamespace = (_: any, newVal: any) => {
    const ns = newVal

    applyNSParam(ns)
    dispatch(setNameSpace(ns))

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
