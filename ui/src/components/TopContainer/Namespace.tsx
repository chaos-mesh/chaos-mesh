import { Autocomplete, TextField } from '@material-ui/core'
import { useHistory, useLocation } from 'react-router-dom'
import { useStoreDispatch, useStoreSelector } from 'store'

import Paper from 'components-mui/Paper'
import T from 'components/T'
import api from 'api'
import { getNamespaces } from 'slices/experiments'
import { setNameSpace } from 'slices/globalStatus'
import { useEffect } from 'react'

const ControlBar = () => {
  const history = useHistory()
  const { pathname } = useLocation()

  const { namespace } = useStoreSelector((state) => state.globalStatus)
  const { namespaces } = useStoreSelector((state) => state.experiments)
  const dispatch = useStoreDispatch()

  useEffect(() => {
    dispatch(getNamespaces())
  }, [dispatch])

  const handleSelectGlobalNamespace = (_: any, newVal: any) => {
    const ns = newVal

    api.auth.namespace(ns)
    dispatch(setNameSpace(ns))

    history.replace('/namespaceSetted')
    setTimeout(() => history.replace(pathname))
  }

  return (
    <Autocomplete
      className="tutorial-namespace"
      sx={{ minWidth: 180 }}
      value={namespace}
      options={['All', ...namespaces]}
      onChange={handleSelectGlobalNamespace}
      disableClearable={true}
      renderInput={(params) => <TextField {...params} size="small" label={T('common.chooseNamespace')} />}
      PaperComponent={(props) => <Paper {...props} sx={{ p: 0 }} />}
    />
  )
}

export default ControlBar
