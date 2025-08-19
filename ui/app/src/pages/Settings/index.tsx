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
import { Stale } from '@/api/queryUtils'
import messages from '@/i18n/messages'
import Checkbox from '@/mui-extends/Checkbox'
import PaperTop from '@/mui-extends/PaperTop'
import SelectField from '@/mui-extends/SelectField'
import Space from '@/mui-extends/Space'
import { useGetCommonConfig } from '@/openapi'
import { useAuthStore } from '@/zustand/auth'
import { useSettingActions, useSettingStore } from '@/zustand/setting'
import { type SystemTheme, useSystemActions, useSystemStore } from '@/zustand/system'
import { Box, Chip, Grow, MenuItem, Typography } from '@mui/material'
import type { SelectChangeEvent } from '@mui/material'

import { T } from '@/components/T'

import logoWhite from '@/images/logo-white.svg'
import logo from '@/images/logo.svg'

import Token from './Token'

const Settings = () => {
  const theme = useSystemStore((state) => state.theme)
  const lang = useSystemStore((state) => state.lang)
  const { setTheme, setLang } = useSystemActions()
  const debugMode = useSettingStore((state) => state.debugMode)
  const enableKubeSystemNS = useSettingStore((state) => state.enableKubeSystemNS)
  const useNewPhysicalMachine = useSettingStore((state) => state.useNewPhysicalMachine)
  const { setDebugMode, setEnableKubeSystemNS, setUseNewPhysicalMachine } = useSettingActions()
  const tokenName = useAuthStore((state) => state.tokenName)

  const { data: config } = useGetCommonConfig({
    query: {
      enabled: false,
      staleTime: Stale.DAY,
    },
  })

  const handleChangeDebugMode = () => setDebugMode(!debugMode)
  const handleChangeEnableKubeSystemNS = () => setEnableKubeSystemNS(!enableKubeSystemNS)
  const handleChangeUseNewPhysicalMachine = () => setUseNewPhysicalMachine(!useNewPhysicalMachine)
  const handleChangeTheme = (e: SelectChangeEvent) => setTheme(e.target.value as SystemTheme)
  const handleChangeLang = (e: SelectChangeEvent) => setLang(e.target.value)

  return (
    <Grow in={true} style={{ transformOrigin: '0 0 0' }}>
      <div style={{ height: '100%' }}>
        <Space>
          <PaperTop title={<T id="settings.title" />} h1 divider />
          {config?.security_mode && tokenName && <Token />}
          <PaperTop title={<T id="experiments.title" />} />
          <Checkbox
            label={<T id="settings.debugMode.title" />}
            helperText={<T id="settings.debugMode.choose" />}
            checked={debugMode}
            onChange={handleChangeDebugMode}
          />
          <Checkbox
            label={<T id="settings.enableKubeSystemNS.title" />}
            helperText={<T id="settings.enableKubeSystemNS.choose" />}
            checked={enableKubeSystemNS}
            onChange={handleChangeEnableKubeSystemNS}
          />
          <Checkbox
            label={
              <Space spacing={1} direction="row" alignItems="center">
                <Box>
                  <T id="settings.useNewPhysicalMachineCRD.title" />
                </Box>
                <Chip label="Preview" color="primary" size="small" />
              </Space>
            }
            helperText={<T id="settings.useNewPhysicalMachineCRD.choose" />}
            checked={useNewPhysicalMachine}
            onChange={handleChangeUseNewPhysicalMachine}
          />
          <PaperTop title={<T id="settings.theme.title" />} />
          <SelectField
            label={<T id="settings.theme.title" />}
            helperText={<T id="settings.theme.choose" />}
            value={theme}
            onChange={handleChangeTheme}
            sx={{ width: 300 }}
          >
            <MenuItem value="light">
              <Typography>
                <T id="settings.theme.light" />
              </Typography>
            </MenuItem>
            <MenuItem value="dark">
              <Typography>
                <T id="settings.theme.dark" />
              </Typography>
            </MenuItem>
          </SelectField>
          <PaperTop title={<T id="settings.lang.title" />} />
          <SelectField
            label={<T id="settings.lang.title" />}
            helperText={<T id="settings.lang.choose" />}
            value={lang}
            onChange={handleChangeLang}
            sx={{ width: 300 }}
          >
            {Object.keys(messages).map((lang) => (
              <MenuItem key={lang} value={lang}>
                <Typography>
                  <T id={`settings.lang.${lang}`} />
                </Typography>
              </MenuItem>
            ))}
          </SelectField>

          <PaperTop title={<T id="common.version" />} />
          <Box>
            <img src={theme === 'light' ? logo : logoWhite} alt="Chaos Mesh" style={{ width: 192 }} />
            <Typography variant="body2" color="textSecondary">
              Git Version: {config?.version}
            </Typography>
          </Box>
        </Space>
      </div>
    </Grow>
  )
}

export default Settings
