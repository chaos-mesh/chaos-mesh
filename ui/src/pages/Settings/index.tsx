import { Box, Grid } from '@material-ui/core'

import Other from './Other'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
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
        <Paper>
          <Box mx={3}>
            {securityMode && (
              <>
                <PaperTop title={T('settings.addToken.title')} />
                <Token />
                <Box my={6} />
                <TokensTable />
                <Box mb={6} />
              </>
            )}
          </Box>

          <PaperTop title={T('common.other')} />
          <Box mx={3}>
            <Other />
          </Box>
        </Paper>
      </Grid>
    </Grid>
  )
}

export default Settings
