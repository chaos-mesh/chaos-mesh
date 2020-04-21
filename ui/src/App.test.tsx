import React from 'react'
import { render } from '@testing-library/react'
import App from './App'

// TODO: improvement test cases
test('renders dashboard text', () => {
  const { getByText } = render(<App />)
  const linkElement = getByText(/dashboard/i)
  expect(linkElement).toBeInTheDocument()
})
