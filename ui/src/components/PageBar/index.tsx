import React from 'react'
import { Link as RouterLink } from 'react-router-dom'
import { Typography, Breadcrumbs, Link, Paper } from '@material-ui/core'

import { makeStyles, Theme, createStyles } from '@material-ui/core/styles'

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

interface PageBarProps {
  breadcrumbs: BreadCrumbProps[]
}

export default function PageBar(props: PageBarProps) {
  const { breadcrumbs } = props
  const classes = useStyles()

  return (
    <Paper className={classes.pager} square>
      <Breadcrumbs aria-label="breadcrumb">
        {breadcrumbs.map((b: BreadCrumbProps, index: number) => {
          const isLast = index === breadcrumbs.length - 1

          if (isLast) {
            return (
              <Typography color="inherit" key={b.name}>
                {b.name}
              </Typography>
            )
          }

          return (
            <Link
              color="primary"
              component={RouterLink as any}
              to={b.path}
              key={b.name}
            >
              {b.name}
            </Link>
          )
        })}
      </Breadcrumbs>
    </Paper>
  )
}
