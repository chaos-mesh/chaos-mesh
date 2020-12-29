import { Box, Grid, Paper, Typography } from '@material-ui/core'

import Other from './Other'
import React from 'react'
import T from 'components/T'
import Token from 'components/Token'
import TokensTable from './TokensTable'
import { useStoreSelector } from 'store'

const Settings = () => {
  const { securityMode } = useStoreSelector((state) => state.globalStatus)

  return (
    <Grid container>
      <Grid item sm={12} md={6}>
        <Paper variant="outlined">
          <Box p={6}>
            {securityMode && (
              <>
                <Typography variant="h6" gutterBottom>
                  {T('settings.addToken.title')}
                </Typography>
                <Token />
                <Box my={6} />
                <TokensTable />
                <Box mb={6} />
              </>
            )}
            <Typography variant="h6" gutterBottom>
              {T('common.other')}
            </Typography>
            <Other />
          </Box>
        </Paper>
      </Grid>
    </Grid>
  )
}

export default Settings
