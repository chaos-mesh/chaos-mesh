import { Box, BoxProps } from '@material-ui/core'

import EmptyStreet from 'images/assets/undraw_empty_street.svg'
import EmptyStreetDark from 'images/assets/undraw_empty_street-dark.svg'
import { styled } from '@material-ui/styles'
import { useStoreSelector } from 'store'

const StyledBox = styled(Box)({
  position: 'absolute',
  top: '50%',
  left: '50%',
  transform: 'translate3d(-50%, -50%, 0)',
})

interface NotFoundProps extends BoxProps {
  illustrated?: boolean
}

const NotFound: React.FC<NotFoundProps> = ({ illustrated = false, children, ...rest }) => {
  const { theme } = useStoreSelector((state) => state.settings)

  return (
    <StyledBox {...rest}>
      {illustrated && (
        <Box mb={6}>
          <img style={{ width: '50%' }} src={theme === 'light' ? EmptyStreet : EmptyStreetDark} alt="Not found" />
        </Box>
      )}
      {children}
    </StyledBox>
  )
}

export default NotFound
