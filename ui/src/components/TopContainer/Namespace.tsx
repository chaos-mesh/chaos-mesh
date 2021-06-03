import { useHistory, useLocation } from 'react-router-dom'
import { useStoreDispatch, useStoreSelector } from 'store'

import Autocomplete from '@material-ui/lab/Autocomplete'
import Paper from 'components-mui/Paper'
import T from 'components/T'
import { TextField } from '@material-ui/core'
import api from 'api'
import clsx from 'clsx'
import { getNamespaces } from 'slices/experiments'
import { makeStyles } from '@material-ui/styles'
import { setNameSpace } from 'slices/globalStatus'
import { useEffect } from 'react'

const useStyles = makeStyles((theme) => ({
  namespace: {
    minWidth: 180,
  },
}))

const ControlBar = () => {
  const classes = useStyles()
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
      className={clsx(classes.namespace, 'nav-namespace')}
      value={namespace}
      options={['All', ...namespaces]}
      onChange={handleSelectGlobalNamespace}
      disableClearable={true}
      renderInput={(params) => <TextField {...params} size="small" label={T('common.chooseNamespace')} />}
      PaperComponent={(props) => <Paper {...props} padding={0} />}
    />
  )
}

export default ControlBar
