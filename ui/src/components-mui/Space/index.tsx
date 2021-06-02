import { Box, BoxProps } from '@material-ui/core'

import { Theme } from '@material-ui/core/styles'
import { styled } from '@material-ui/styles'

interface SpaceProps {
  spacing?: number
  vertical?: boolean
}
type Props = BoxProps & SpaceProps

export default styled(({ spacing, vertical, children, ...rest }: Props) => <Box {...rest}>{children}</Box>)<
  Theme,
  SpaceProps
>(({ theme, spacing = 3, vertical = false }) => {
  const direction = vertical ? 'marginBottom' : 'marginRight'

  return {
    '& > *': {
      [direction]: theme.spacing(spacing),
      '&:last-child': {
        [direction]: 0,
      },
    },
  }
})
