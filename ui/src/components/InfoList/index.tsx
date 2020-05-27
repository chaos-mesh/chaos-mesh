import React from 'react'

import { List, ListItem, ListItemText, Typography } from '@material-ui/core'
import { createStyles, Theme, makeStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    list: {
      width: '100%',
      backgroundColor: theme.palette.background.default,
    },
    listItem: {
      display: 'flex',
    },
    itemLabel: {
      minWidth: '6rem',
      marginRight: theme.spacing(4),
      textTransform: 'capitalize',
    },
  })
)

interface ItemProps {
  key: string
  value: string | React.ReactNode
}

interface InfoListProps {
  info: { [key: string]: string | React.ReactNode }
}

export default function Info({ info }: InfoListProps) {
  const classes = useStyles()

  const list = Object.keys(info).map((i) => {
    return {
      key: i,
      value: info[i],
    }
  })

  return (
    <List className={classes.list}>
      {list.map(({ key, value }: ItemProps) => {
        return (
          <ListItem key={key} className={classes.listItem}>
            <Typography className={classes.itemLabel} component="span" color="textPrimary">
              {key}
            </Typography>

            <Typography color="textPrimary">{value}</Typography>
            <ListItemText />
          </ListItem>
        )
      })}
    </List>
  )
}
