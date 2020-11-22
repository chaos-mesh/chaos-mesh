import { Box, Container, MenuItem, Paper, TextField } from '@material-ui/core'
import { RootState, useStoreDispatch } from 'store'
import { setLang, setTheme } from 'slices/settings'

import PaperTop from 'components-mui/PaperTop'
import React from 'react'
import T from 'components/T'
import messages from 'i18n/messages'
import { useSelector } from 'react-redux'

const Settings = () => {
  const { settings } = useSelector((state: RootState) => state)
  const { theme, lang } = settings
  const dispatch = useStoreDispatch()

  const handleChangeTheme = (e: React.ChangeEvent<HTMLInputElement>) => dispatch(setTheme(e.target.value))
  const handleChangeLang = (e: React.ChangeEvent<HTMLInputElement>) => dispatch(setLang(e.target.value))

  return (
    <Paper variant="outlined" style={{ height: '100%' }}>
      <PaperTop title={T('settings.title')} />

      <Container>
        <Box p={6} width={400} maxWidth="100%">
          <Box mb={2}>
            <TextField
              variant="outlined"
              select
              margin="dense"
              fullWidth
              value={theme}
              label={T('settings.theme')}
              helperText={T('settings.chooseTheme')}
              onChange={handleChangeTheme}
            >
              <MenuItem value="light">{T(`settings.themeLight`)}</MenuItem>
              <MenuItem value="dark">{T(`settings.themeDark`)}</MenuItem>
            </TextField>
          </Box>
          <Box mb={2}>
            <TextField
              variant="outlined"
              select
              margin="dense"
              fullWidth
              value={lang}
              label={T('settings.language')}
              helperText={T('settings.chooseInterfaceLanguage')}
              onChange={handleChangeLang}
            >
              {Object.keys(messages).map((lang) => (
                <MenuItem key={lang} value={lang}>
                  {T(`settings.${lang}`)}
                </MenuItem>
              ))}
            </TextField>
          </Box>
        </Box>
      </Container>
    </Paper>
  )
}

export default Settings
