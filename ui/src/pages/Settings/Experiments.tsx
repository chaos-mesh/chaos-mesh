import { Box, Checkbox, FormControl, FormControlLabel, FormHelperText, Typography } from '@material-ui/core'
import { RootState, useStoreDispatch } from 'store'
import { setDebugMode, setEnableKubeSystemNS } from 'slices/settings'

import React from 'react'
import T from 'components/T'
import { useSelector } from 'react-redux'

const Experiments = () => {
  const { settings } = useSelector((state: RootState) => state)
  const { debugMode, enableKubeSystemNS } = settings
  const dispatch = useStoreDispatch()

  const handleChangeDebugMode = (e: React.ChangeEvent<HTMLInputElement>) => dispatch(setDebugMode(e.target.checked))
  const handleChangeEnableKubeSystemNS = (e: React.ChangeEvent<HTMLInputElement>) =>
    dispatch(setEnableKubeSystemNS(e.target.checked))

  return (
    <>
      {/* devMode */}
      <Box mb={3}>
        <FormControl>
          <FormControlLabel
            control={<Checkbox color="primary" checked={debugMode} onChange={handleChangeDebugMode} />}
            label={<Typography variant="body2">{T('settings.debugMode.title')}</Typography>}
          />
          <FormHelperText>{T('settings.debugMode.choose')}</FormHelperText>
        </FormControl>
      </Box>

      {/* Enable kube-system */}
      <Box mb={3}>
        <FormControl>
          <FormControlLabel
            control={
              <Checkbox color="primary" checked={enableKubeSystemNS} onChange={handleChangeEnableKubeSystemNS} />
            }
            label={<Typography variant="body2">{T('settings.enableKubeSystemNS.title')}</Typography>}
          />
          <FormHelperText>{T('settings.enableKubeSystemNS.choose')}</FormHelperText>
        </FormControl>
      </Box>
    </>
  )
}

export default Experiments
