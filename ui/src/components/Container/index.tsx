import React, { FC } from 'react'
import { Container, ContainerProps } from '@material-ui/core'

import { makeStyles, Theme, createStyles } from '@material-ui/core/styles'

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    container: {
      padding: theme.spacing(5),
    },
  })
)

const ContentContainer: FC<ContainerProps> = (props) => {
  const classes = useStyles()

  return <Container maxWidth="xl" className={classes.container} {...props} />
}

export default ContentContainer
