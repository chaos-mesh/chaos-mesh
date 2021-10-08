import { Box, Button } from '@material-ui/core'
import { useStoreDispatch, useStoreSelector } from 'store'

import Cookies from 'js-cookie'
import GoogleIcon from '@material-ui/icons/Google'
import LS from 'lib/localStorage'
import PaperTop from 'components-mui/PaperTop'
import T from 'components/T'
import { setConfirm } from 'slices/globalStatus'
import { truncate } from 'lib/utils'
import { useHistory } from 'react-router-dom'
import { useIntl } from 'react-intl'

const Token = () => {
  const history = useHistory()
  const intl = useIntl()

  const { tokens, tokenName } = useStoreSelector((state) => state.globalStatus)
  const tokenDesc =
    tokenName === 'gcp' ? (
      <Box display="flex" alignItems="center">
        {T('settings.addToken.gcp')}
        <GoogleIcon sx={{ ml: 1 }} />
      </Box>
    ) : (
      tokenName + ': ' + truncate(tokens[0].token)
    )
  const dispatch = useStoreDispatch()

  const handleRemoveToken = () =>
    dispatch(
      setConfirm({
        title: T('common.logout', intl),
        description: T('common.logoutDesc', intl),
        handle: handleRemoveTokenConfirm,
      })
    )

  const handleRemoveTokenConfirm = () => {
    if (tokenName === 'gcp') {
      Cookies.remove('access_token')
      Cookies.remove('refresh_token')
      Cookies.remove('expiry')
    } else {
      LS.remove('token')
      LS.remove('token-name')
    }

    history.go(0)
  }

  return (
    <PaperTop title={T('settings.addToken.token')} subtitle={tokenDesc}>
      <Button
        variant="outlined"
        size="small"
        color="secondary"
        sx={{ width: 64, height: 32 }}
        onClick={handleRemoveToken}
      >
        {T('common.logout')}
      </Button>
    </PaperTop>
  )
}

export default Token
