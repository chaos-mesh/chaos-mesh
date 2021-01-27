import { Box, MenuItem, TextField, Typography } from '@material-ui/core'
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
          <MenuItem value="light">
            <Typography variant="body2">{T(`settings.theme.light`)}</Typography>
          </MenuItem>
          <MenuItem value="dark">
            <Typography variant="body2">{T(`settings.theme.dark`)}</Typography>
          </MenuItem>
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
              <Typography variant="body2">{T(`settings.lang.${lang}`)}</Typography>
            </MenuItem>
          ))}
        </TextField>
      </Box>
    </>
  )
}

export default Other
