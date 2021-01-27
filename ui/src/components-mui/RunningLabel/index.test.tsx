import RunningLabel from '.'
import customTheme from 'theme'
import { render } from '@testing-library/react'

describe('<RunningLabel />', () => {
  const { container } = render(<RunningLabel>Running</RunningLabel>)

  it('checks style', () => {
    expect(container.firstChild).toHaveStyle(`background: ${customTheme.palette.warning.main}`)
    expect(container.firstChild).toHaveStyle(`color: ${customTheme.palette.common.white}`)
  })
})
