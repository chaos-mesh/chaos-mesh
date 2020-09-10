import { Box, Container, MenuItem, Paper, TextField } from '@material-ui/core'
import { RootState, useStoreDispatch } from 'store'

import PaperTop from 'components/PaperTop'
import React from 'react'
import T from 'components/T'
import messages from 'i18n/messages'
import { setLang } from 'slices/settings'
import { useSelector } from 'react-redux'

const Settings = () => {
  const { settings } = useSelector((state: RootState) => state)
  const { lang } = settings
  const dispatch = useStoreDispatch()

  const handleChangeLang = (e: React.ChangeEvent<HTMLInputElement>) => dispatch(setLang(e.target.value))

  return (
    <Paper variant="outlined" style={{ height: '100%' }}>
      <PaperTop title={T('settings.title')} />

      <Container>
        <Box p={6} width={400} maxWidth="100%">
          <TextField
            variant="outlined"
            select
            margin="dense"
            fullWidth
            value={lang}
            label={T('settings.languages')}
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
      </Container>
    </Paper>
  )
}

export default Settings
