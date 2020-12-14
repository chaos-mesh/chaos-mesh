import { Box, Button, Grid, Typography } from '@material-ui/core'

import AddIcon from '@material-ui/icons/Add'
import { Link } from 'react-router-dom'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import React from 'react'
import T from 'components/T'
import { makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme) => ({
  container: {
    position: 'relative',
    height: `calc(94px + ${theme.spacing(5.25)})`,
    margin: theme.spacing(3),
    marginTop: 0,
  },
}))

const Welcome = () => {
  const classes = useStyles()

  return (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Paper>
          <PaperTop title={T('dashboard.welcome')} />
          <Box className={classes.container}>
            <Typography>
              <span role="img" aria-label="smiling face with sunglasses">
                😎
              </span>{' '}
              {T('dashboard.welcomeDesc')}
            </Typography>

            <Box position="absolute" bottom={0} width="100%">
              <Button variant="contained" color="primary" fullWidth>
                {T('common.tutorial')}
              </Button>
            </Box>
          </Box>
        </Paper>
      </Grid>
      <Grid item xs={12}>
        <Paper>
          <PaperTop title={T('dashboard.veteran')} />
          <Box className={classes.container}>
            <Typography>
              <span role="img" aria-label="firecracker">
                🧨
              </span>{' '}
              {T('dashboard.veteranDesc')}
            </Typography>

            <Box position="absolute" bottom={0} width="100%">
              <Button
                component={Link}
                to="/newExperiment"
                variant="contained"
                color="primary"
                fullWidth
                startIcon={<AddIcon />}
              >
                {T('newE.title')}
              </Button>
            </Box>
          </Box>
        </Paper>
      </Grid>
    </Grid>
  )
}

export default Welcome
