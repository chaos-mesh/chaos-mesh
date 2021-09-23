import { Box, BoxProps, Divider, Typography } from '@material-ui/core'

interface PaperTopProps {
  title: string | JSX.Element
  subtitle?: string | JSX.Element
  divider?: boolean
  boxProps?: BoxProps
}

const PaperTop: React.FC<PaperTopProps> = ({ title, subtitle, divider, boxProps, children }) => (
  <Box {...boxProps} display="flex" justifyContent="space-between" width="100%">
    <Box flex={1}>
      <Typography variant="h3" gutterBottom={subtitle || divider ? true : false}>
        {title}
      </Typography>
      {subtitle && (
        <Typography variant="body2" color="textSecondary">
          {subtitle}
        </Typography>
      )}
      {divider && <Divider />}
    </Box>
    {children}
  </Box>
)

export default PaperTop
