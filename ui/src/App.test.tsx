import React from 'react'
import { render } from '@testing-library/react'
import App from './App'

test('renders dashboard text', () => {
  const { getByText } = render(<App />)
  const dashboardTextInAppBar = getByText(/dashboard/i)

  expect(dashboardTextInAppBar).toBeInTheDocument()
})
