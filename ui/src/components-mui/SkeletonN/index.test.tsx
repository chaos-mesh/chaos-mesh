import SkeletonN from '.'
import { render } from '@testing-library/react'

describe('<Skeleton />', () => {
  it('matches snapshot', () => {
    const { container } = render(<SkeletonN n={3} />)

    expect(container).toMatchSnapshot()
  })
})
