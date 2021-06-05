import { Box, BoxProps } from '@material-ui/core'

import { styled } from '@material-ui/styles'

interface SpaceProps {
  spacing?: number
  vertical?: boolean
}
type Props = BoxProps & SpaceProps

export default styled(({ spacing, vertical, children, ...rest }: Props) => (
  <Box {...rest} display="flex">
    {children}
  </Box>
))(({ theme, spacing = 3, vertical = false }) => {
  const direction = vertical ? 'marginBottom' : 'marginRight'

  return {
    flexDirection: vertical ? 'column' : 'row',
    '& > *': {
      [direction]: theme.spacing(spacing),
      '&:last-child': {
        [direction]: 0,
      },
    },
  }
})
