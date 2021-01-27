import { Button, DialogActions, DialogContentText, DialogTitle } from '@material-ui/core'

import ConfirmDialog from '.'
import { shallow } from 'enzyme'

describe('<ConfirmDialog />', () => {
  let confirmed = false
  function onConfirm() {
    confirmed = true
  }
  const wrapper = shallow(
    <ConfirmDialog
      open={true}
      setOpen={() => {}}
      title="Test ConfirmDialog"
      description="A description"
      onConfirm={onConfirm}
    />
  )

  it('renders title', () => {
    expect(wrapper.find(DialogTitle).text()).toBe('Test ConfirmDialog')
  })

  it('renders desc', () => {
    expect(wrapper.find(DialogContentText)).toHaveLength(1)
  })

  it('renders custom desc', () => {
    function Foo() {
      return <div>Bar</div>
    }
    const wrapper = shallow(
      <ConfirmDialog open={true} setOpen={() => {}} title="Test ConfirmDialog">
        <Foo />
      </ConfirmDialog>
    )

    expect(wrapper.find(DialogContentText)).toHaveLength(0)
    expect(wrapper.find(Foo)).toHaveLength(1)
  })

  it('simulates confirm', () => {
    wrapper.find(DialogActions).find(Button).at(1).simulate('click')

    expect(confirmed).toBe(true)
  })
})
