import { Box, MenuItem, TextField } from '@material-ui/core'
import { RootState, useStoreDispatch } from 'store'
import { setLang, setTheme } from 'slices/settings'

import React from 'react'
import T from 'components/T'
import messages from 'i18n/messages'
import { useSelector } from 'react-redux'

const Other = () => {
  const { settings } = useSelector((state: RootState) => state)
  const { theme, lang } = settings
  const dispatch = useStoreDispatch()

  const handleChangeTheme = (e: React.ChangeEvent<HTMLInputElement>) => dispatch(setTheme(e.target.value))
  const handleChangeLang = (e: React.ChangeEvent<HTMLInputElement>) => dispatch(setLang(e.target.value))

  return (
    <>
      {/* Theme */}
      <Box mb={3}>
        <TextField
          variant="outlined"
          select
          margin="dense"
          fullWidth
          value={theme}
          label={T('settings.theme.title')}
          helperText={T('settings.theme.choose')}
          onChange={handleChangeTheme}
        >
          <MenuItem value="light">{T(`settings.theme.light`)}</MenuItem>
          <MenuItem value="dark">{T(`settings.theme.dark`)}</MenuItem>
        </TextField>
      </Box>

      {/* Langauge */}
      <Box mb={3}>
        <TextField
          variant="outlined"
          select
          margin="dense"
          fullWidth
          value={lang}
          label={T('settings.lang.title')}
          helperText={T('settings.lang.choose')}
          onChange={handleChangeLang}
        >
          {Object.keys(messages).map((lang) => (
            <MenuItem key={lang} value={lang}>
              {T(`settings.lang.${lang}`)}
            </MenuItem>
          ))}
        </TextField>
      </Box>
    </>
  )
}

export default Other
