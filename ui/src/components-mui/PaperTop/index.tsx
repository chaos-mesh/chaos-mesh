import { Box, BoxProps, Typography } from '@material-ui/core'

interface PaperTopProps {
  title: string | JSX.Element
  subtitle?: string | JSX.Element
  boxProps?: BoxProps
}

const PaperTop: React.FC<PaperTopProps> = ({ title, subtitle, boxProps, children }) => (
  <Box {...boxProps} display="flex" justifyContent="space-between" width="100%">
    <div>
      <Typography component="div" gutterBottom={subtitle ? true : false}>
        {title}
      </Typography>
      {subtitle && (
        <Typography variant="body2" color="textSecondary">
          {subtitle}
        </Typography>
      )}
    </div>
    {children}
  </Box>
)

export default PaperTop
