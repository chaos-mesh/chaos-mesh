import { act, render, screen } from 'test-utils'
import i18n, { T } from '.'

describe('i18n() and <T />', () => {
  test('displays k8s.title with i18n(string)', async () => {
    await act(async () => {
      render(<div>{i18n('k8s.title')}</div>)
    })

    screen.getByText('Kubernetes')
  })

  test('displays k8s.title with <T />', async () => {
    await act(async () => {
      render(<T id="k8s.title" />)
    })

    screen.getByText('Kubernetes')
  })
})
