import { Breadcrumbs, Link, Paper, Typography } from '@material-ui/core'
import { Link as RouterLink, useLocation } from 'react-router-dom'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import React from 'react'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    pager: {
      padding: theme.spacing(5),
    },
  })
)

interface BreadCrumbProps {
  name: string
  path?: string
}

function pathnameToBreadCrumbs(pathname: string) {
  const nameArray = pathname.slice(1).split('/')

  return nameArray.map((name, i) => {
    const b: BreadCrumbProps = {
      name: i === 0 ? name.charAt(0).toUpperCase() + name.slice(1) : name,
    }

    if (i < nameArray.length - 1) {
      b.path = '/' + nameArray.slice(0, i + 1).join('/')
    }

    return b
  })
}

export default function PageBar() {
  const classes = useStyles()
  const { pathname } = useLocation()
  const breadcrumbs = pathnameToBreadCrumbs(pathname)

  document.title = breadcrumbs.map((b) => b.name).join(' > ')

  return (
    <Paper className={classes.pager} square>
      <Breadcrumbs aria-label="breadcrumb">
        {breadcrumbs.map((b) => {
          return b.path ? (
            <Link key={b.name} component={RouterLink} to={b.path} color="primary">
              {b.name}
            </Link>
          ) : (
            <Typography key={b.name} color="inherit">
              {b.name}
            </Typography>
          )
        })}
      </Breadcrumbs>
    </Paper>
  )
}
