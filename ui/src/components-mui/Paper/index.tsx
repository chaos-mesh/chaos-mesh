import { Paper as MUIPaper, PaperProps } from '@material-ui/core'

const Paper: React.FC<PaperProps> = ({ sx, children, ...rest }) => (
  <MUIPaper {...rest} variant="outlined" sx={{ position: 'relative', height: '100%', p: 4.5, ...sx }}>
    {children}
  </MUIPaper>
)

export default Paper
