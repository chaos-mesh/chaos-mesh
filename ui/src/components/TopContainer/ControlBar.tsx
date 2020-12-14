import { AppBar, Box, MenuItem, TextField, Toolbar } from '@material-ui/core'
import React, { useEffect, useState } from 'react'
import { useHistory, useLocation } from 'react-router-dom'
import { useStoreDispatch, useStoreSelector } from 'store'

import api from 'api'
import { makeStyles } from '@material-ui/core/styles'
import { setNameSpace } from 'slices/globalStatus'

const useStyles = makeStyles({
  toolbar: {
    minHeight: 48,
  },
  namespaces: {
    height: 48,
    minWidth: 180,
  },
})

const ControlBar = () => {
  const classes = useStyles()
  const history = useHistory()
  const { pathname } = useLocation()

  const { namespace } = useStoreSelector((state) => state.globalStatus)
  const dispatch = useStoreDispatch()

  const [namespaces, setNamespaces] = useState(['All'])

  const fetchNamespaces = () => {
    api.common
      .chaosAvailableNamespaces()
      .then(({ data }) => setNamespaces(['All', ...data]))
      .catch(console.error)
  }

  const handleSelectGlobalNamespace = (e: React.ChangeEvent<HTMLInputElement>) => {
    const ns = e.target.value

    api.auth.namespace(ns)
    dispatch(setNameSpace(ns))

    history.replace('/namespaceSetted')
    setTimeout(() => history.replace(pathname))
  }

  useEffect(fetchNamespaces, [])

  return (
    <AppBar position="relative" color="inherit" elevation={0}>
      <Toolbar className={classes.toolbar} disableGutters>
        <Box>
          <TextField
            className={classes.namespaces}
            variant="outlined"
            color="primary"
            select
            InputProps={{
              style: {
                height: '100%',
              },
            }}
            value={namespace}
            onChange={handleSelectGlobalNamespace}
          >
            {namespaces.map((option) => (
              <MenuItem key={option} value={option}>
                {option}
              </MenuItem>
            ))}
          </TextField>
        </Box>
      </Toolbar>
    </AppBar>
  )
}

export default ControlBar
