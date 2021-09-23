import { Box, Grid, Grow, Typography } from '@material-ui/core'

import Experiments from './Experiments'
import Other from './Other'
import PaperTop from 'components-mui/PaperTop'
import Space from 'components-mui/Space'
import T from 'components/T'
import Token from './Token'
import logo from 'images/logo.svg'
import logoWhite from 'images/logo-white.svg'
import { useStoreSelector } from 'store'

const Settings = () => {
  const state = useStoreSelector((state) => state)
  const { securityMode, version } = state.globalStatus
  const { theme } = state.settings

  return (
    <Grow in={true} style={{ transformOrigin: '0 0 0' }}>
      <Grid container spacing={6}>
        <Grid item sm={12} md={6}>
          <Space spacing={6}>
            {securityMode && <Token />}
            <Experiments />
            <Other />

            <PaperTop title={T('common.version')} divider />
            <Box>
              <img src={theme === 'light' ? logo : logoWhite} alt="Chaos Mesh" style={{ width: 192 }} />
              <Typography variant="body2" color="textSecondary">
                Git Version: {version}
              </Typography>
            </Box>
          </Space>
        </Grid>
      </Grid>
    </Grow>
  )
}

export default Settings
