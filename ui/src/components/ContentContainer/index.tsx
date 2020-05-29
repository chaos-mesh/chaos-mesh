import { Container, ContainerProps } from '@material-ui/core'
import { Theme, createStyles, makeStyles } from '@material-ui/core/styles'

import React from 'react'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      padding: theme.spacing(6),
    },
  })
)

const ContentContainer: React.FC<ContainerProps> = (props) => {
  const classes = useStyles()

  return <Container className={classes.root} maxWidth="xl" {...props} />
}

export default ContentContainer
