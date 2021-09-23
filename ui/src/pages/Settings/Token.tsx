import { useStoreDispatch, useStoreSelector } from 'store'

import { Button } from '@material-ui/core'
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
  const dispatch = useStoreDispatch()

  const handleRemoveToken = () =>
    dispatch(
      setConfirm({
        title: `${T('common.logout', intl)} ${tokenName}`,
        description: T('common.logoutDesc', intl),
        handle: handleRemoveTokenConfirm,
      })
    )

  const handleRemoveTokenConfirm = () => {
    LS.remove('token')
    LS.remove('token-name')
    history.go(0)
  }

  return (
    <PaperTop title={T('settings.addToken.token')} subtitle={tokenName + ': ' + truncate(tokens[0].token)}>
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
