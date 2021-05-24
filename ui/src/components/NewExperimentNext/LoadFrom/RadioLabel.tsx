import { Box, Typography } from '@material-ui/core'

import { truncate } from 'lib/utils'

const RadioLabel = (label: string, uuid?: string) => (
  <Box display="flex" justifyContent="space-between" alignItems="center">
    <Typography>{label}</Typography>
    {uuid && (
      <Box ml={3}>
        <Typography variant="body2" color="textSecondary" title={uuid}>
          {truncate(uuid)}
        </Typography>
      </Box>
    )}
  </Box>
)

export default RadioLabel
