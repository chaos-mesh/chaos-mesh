import { Stack, StackProps } from '@material-ui/core'

function Space({ children, ...rest }: StackProps) {
  return (
    <Stack spacing={3} {...rest}>
      {children}
    </Stack>
  )
}

export default Space
