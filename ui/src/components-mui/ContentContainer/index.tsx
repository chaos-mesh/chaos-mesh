import { Container, ContainerProps } from '@material-ui/core'

import React from 'react'
import { styled } from '@material-ui/styles'

type ContentContainerProps = Omit<ContainerProps, 'maxWidth'>

const ContentContainer = styled((props: ContentContainerProps) => <Container maxWidth="xl" {...props} />)(
  ({ theme }) => ({
    position: 'relative',
    padding: `${theme.spacing(6)} !important`,
  })
)

export default ContentContainer
