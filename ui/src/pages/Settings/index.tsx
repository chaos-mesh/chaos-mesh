/*
 * Copyright 2021 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
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
