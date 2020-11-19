import { Box, Divider, Grid, Paper, Typography } from '@material-ui/core'

import Other from './Other'
import React from 'react'
import T from 'components/T'
import Token from 'components/Token'
import TokensTable from './TokensTable'

const Title: React.FC = ({ children }) => (
  <>
    <Typography variant="h6" gutterBottom>
      {children}
    </Typography>
    <Divider />
    <Box mb={6} />
  </>
)

const Settings = () => (
  <Grid container justify="center">
    <Grid item sm={12} md={6} zeroMinWidth>
      <Paper variant="outlined">
        <Box p={6}>
          <Title>{T('settings.addToken.title')}</Title>
          <Token />
          <Box my={6} />
          <TokensTable />
          <Box mb={6} />
          <Title>{T('common.other')}</Title>
          <Other />
        </Box>
      </Paper>
    </Grid>
  </Grid>
)

export default Settings
