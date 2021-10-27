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
import { Box, MenuItem, TextField, Typography } from '@material-ui/core'
import { setLang, setTheme } from 'slices/settings'
import { useStoreDispatch, useStoreSelector } from 'store'

import T from 'components/T'
import messages from 'i18n/messages'

const Other = () => {
  const { settings } = useStoreSelector((state) => state)
  const { theme, lang } = settings
  const dispatch = useStoreDispatch()

  const handleChangeTheme = (e: React.ChangeEvent<HTMLInputElement>) => dispatch(setTheme(e.target.value))
  const handleChangeLang = (e: React.ChangeEvent<HTMLInputElement>) => dispatch(setLang(e.target.value))

  return (
    <>
      {/* Theme */}
      <Box>
        <TextField
          select
          size="small"
          style={{ width: '50%' }}
          label={T('settings.theme.title')}
          helperText={T('settings.theme.choose')}
          value={theme}
          onChange={handleChangeTheme}
        >
          <MenuItem value="light">
            <Typography>{T(`settings.theme.light`)}</Typography>
          </MenuItem>
          <MenuItem value="dark">
            <Typography>{T(`settings.theme.dark`)}</Typography>
          </MenuItem>
        </TextField>
      </Box>

      {/* Language */}
      <Box>
        <TextField
          select
          size="small"
          style={{ width: '50%' }}
          label={T('settings.lang.title')}
          helperText={T('settings.lang.choose')}
          value={lang}
          onChange={handleChangeLang}
        >
          {Object.keys(messages).map((lang) => (
            <MenuItem key={lang} value={lang}>
              <Typography>{T(`settings.lang.${lang}`)}</Typography>
            </MenuItem>
          ))}
        </TextField>
      </Box>
    </>
  )
}

export default Other
