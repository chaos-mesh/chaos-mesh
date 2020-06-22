import { Container, ContainerProps } from '@material-ui/core'

import React from 'react'
import { styled } from '@material-ui/core/styles'

type ContentContainerProps = Omit<ContainerProps, 'maxWidth'>

const ContentContainer = styled((props: ContentContainerProps) => <Container maxWidth="xl" {...props} />)(
  ({ theme }) => ({
    position: 'relative',
    padding: theme.spacing(6),
  })
)

export default ContentContainer
