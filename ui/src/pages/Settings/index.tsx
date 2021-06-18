import { Grid, Grow, Typography } from '@material-ui/core'

import Experiments from './Experiments'
import Other from './Other'
import Paper from 'components-mui/Paper'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import T from 'components/T'
import Token from 'components/Token'
import TokensTable from './TokensTable'
import logo from 'images/logo.svg'
import logoWhite from 'images/logo-white.svg'
import { useStoreSelector } from 'store'

const Settings = () => {
  const state = useStoreSelector((state) => state)
  const { securityMode, version } = state.globalStatus
  const { theme } = state.settings

  return (
    <Grow in={true} style={{ transformOrigin: '0 0 0' }}>
      <Grid container>
        <Grid item sm={12} md={8}>
          <Paper>
            <Space>
              {securityMode && (
                <>
                  <PaperTop title={T('settings.addToken.title')} />
                  <Token />
                  <TokensTable />
                </>
              )}

              <PaperTop title={T('experiments.title')} />
              <Experiments />

              <PaperTop title={T('common.other')} />
              <Other />

              <PaperTop title={T('common.version')} />
              <img src={theme === 'light' ? logo : logoWhite} alt="Chaos Mesh" style={{ width: 192 }} />
              <Typography variant="body2" color="textSecondary">
                Git Version: {version}
              </Typography>
            </Space>
          </Paper>
        </Grid>
      </Grid>
    </Grow>
  )
}

export default Settings
