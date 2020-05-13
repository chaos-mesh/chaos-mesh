import React, { FC } from 'react'
import { Box, Paper, Toolbar, Typography } from '@material-ui/core'
import ListOutlinedIcon from '@material-ui/icons/ListOutlined'
import { makeStyles, Theme, createStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    paper: {
      paddingLeft: theme.spacing(2),
      paddingRight: theme.spacing(2),
      borderTop: `1px solid rgba(0, 0, 0, 0.12)`,
    },
    toolbar: { ...theme.mixins.toolbar, justifyContent: 'space-between' },
  })
)

export const CurrentStatus = () => {
  // flexbox: https://material-ui.com/system/flexbox/#api
  const boxProps = {
    display: 'flex',
    alignItems: 'center',
    mr: 2,
  }

  return (
    <>
      <Box {...boxProps}>
        <ListOutlinedIcon />
      </Box>
      <Box {...boxProps}>
        <Typography variant="subtitle2">Total: 26</Typography>
      </Box>
      <Box {...boxProps}>
        <Typography variant="subtitle2">Running: 23</Typography>
      </Box>
      <Box {...boxProps}>
        <Typography variant="subtitle2">Failed: 3</Typography>
      </Box>
    </>
  )
}

const Bar: FC<{}> = ({ children }) => {
  const classes = useStyles()

  return (
    <Paper square elevation={0} className={classes.paper}>
      <Toolbar className={classes.toolbar}>
        <Box display="flex">{children}</Box>

        <Box display="flex">
          <CurrentStatus />
        </Box>
      </Toolbar>
    </Paper>
  )
}

export default Bar
