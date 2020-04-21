import React, { FC } from 'react'
import { Paper, Toolbar, Typography } from '@material-ui/core'
import ListOutlinedIcon from '@material-ui/icons/ListOutlined'
import { makeStyles, Theme, createStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    paper: {
      borderTop: `1px solid rgba(0, 0, 0, 0.12)`,
      paddingLeft: theme.spacing(2),
      paddingRight: theme.spacing(2),
    },
    toolbar: { ...theme.mixins.toolbar, justifyContent: 'space-between' },
    row: {
      display: 'flex',
    },
    item: {
      display: 'flex',
      alignItems: 'center',
      marginRight: theme.spacing(2),
    },
  })
)

//  TODO: get real dynamic data by using global context hooks
export const CurrentStatus = () => {
  const classes = useStyles()

  return (
    <>
      <div className={classes.item}>
        <ListOutlinedIcon />
      </div>
      <div className={classes.item}>
        <Typography variant="subtitle2">Total: 26</Typography>
      </div>
      <div className={classes.item}>
        <Typography variant="subtitle2">Running: 23</Typography>
      </div>
      <div className={classes.item}>
        <Typography variant="subtitle2">Failed: 3</Typography>
      </div>
    </>
  )
}

const Bar: FC<{}> = ({ children }) => {
  const classes = useStyles()

  return (
    <Paper square elevation={0} className={classes.paper}>
      <Toolbar className={classes.toolbar}>
        <div className={classes.row}>{children}</div>

        <div className={classes.row}>
          <CurrentStatus />
        </div>
      </Toolbar>
    </Paper>
  )
}

export default Bar
