import { Button, Checkbox, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@material-ui/core'
import { setConfirm, setTokenName, setTokens } from 'slices/globalStatus'

import LS from 'lib/localStorage'
import PaperContainer from 'components-mui/PaperContainer'
import { RootState } from 'store'
import T from 'components/T'
import { TokenFormValues } from 'components/Token'
import api from 'api'
import { useIntl } from 'react-intl'
import { useSelector } from 'react-redux'
import { useStoreDispatch } from 'store'

const TokensTable = () => {
  const intl = useIntl()

  const { tokens, tokenName } = useSelector((state: RootState) => state.globalStatus)
  const dispatch = useStoreDispatch()

  const handleUseToken = (_token: TokenFormValues) => () => {
    dispatch(setTokenName(_token.name))
    api.auth.token(_token.token)
  }

  const handleRemoveToken = (token: TokenFormValues) => () =>
    dispatch(
      setConfirm({
        title: `${intl.formatMessage({ id: 'common.delete' })} ${token.name}`,
        description: intl.formatMessage({ id: 'settings.addToken.deleteDesc' }),
        handle: handleRemoveTokenConfirm(token.name),
      })
    )

  const handleRemoveTokenConfirm = (n: string) => () => {
    const current = tokens.filter(({ name }) => name !== n)

    if (current.length) {
      dispatch(setTokens(current))

      if (n === tokenName) {
        api.auth.resetToken()
        dispatch(setTokenName(''))
      }
    } else {
      LS.remove('token')
      LS.remove('token-name')
      window.location.reload()
    }
  }

  return (
    <TableContainer component={PaperContainer}>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell />
            <TableCell>{T('common.name')}</TableCell>
            <TableCell>{T('settings.addToken.token')}</TableCell>
            <TableCell>{T('common.status')}</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {tokens.map((token) => {
            const key = `${token.name}:${token.token}`

            return (
              <TableRow key={key}>
                <TableCell padding="checkbox">
                  <Checkbox indeterminate checked={true} onChange={handleRemoveToken(token)} />
                </TableCell>
                <TableCell>{token.name}</TableCell>
                <TableCell>{'*'.repeat(12)}</TableCell>
                <TableCell>
                  <Button
                    onClick={handleUseToken(token)}
                    variant="outlined"
                    color="primary"
                    size="small"
                    disabled={token.name === tokenName}
                  >
                    {token.name === tokenName ? T('common.using') : T('common.use')}
                  </Button>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </TableContainer>
  )
}

export default TokensTable
