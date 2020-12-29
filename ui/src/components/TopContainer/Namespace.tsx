import { MenuItem, TextField } from '@material-ui/core'
import React, { useEffect } from 'react'
import { useHistory, useLocation } from 'react-router-dom'
import { useStoreDispatch, useStoreSelector } from 'store'

import T from 'components/T'
import api from 'api'
import clsx from 'clsx'
import { getNamespaces } from 'slices/experiments'
import { makeStyles } from '@material-ui/core/styles'
import { setNameSpace } from 'slices/globalStatus'

const useStyles = makeStyles((theme) => ({
  namespace: {
    minWidth: 180,
    '& .MuiInputBase-root': {
      height: 32,
      color: theme.palette.background.default,
    },
    '& .MuiFormLabel-root': {
      margin: 0,
      color: theme.palette.background.default,
    },
    '& .MuiSelect-icon': {
      color: theme.palette.background.default,
    },
    '& .MuiOutlinedInput-notchedOutline': {
      borderColor: `${theme.palette.background.default} !important`,
    },
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

  const handleSelectGlobalNamespace = (e: React.ChangeEvent<{ value: unknown }>) => {
    const ns = e.target.value as string

    api.auth.namespace(ns)
    dispatch(setNameSpace(ns))

    history.replace('/namespaceSetted')
    setTimeout(() => history.replace(pathname))
  }

  return (
    <TextField
      className={clsx(classes.namespace, 'nav-namespace')}
      select
      variant="outlined"
      label={T('common.chooseNamespace')}
      value={namespace}
      onChange={handleSelectGlobalNamespace}
    >
      {['All', ...namespaces].map((option) => (
        <MenuItem key={option} value={option}>
          {option}
        </MenuItem>
      ))}
    </TextField>
  )
}

export default ControlBar
