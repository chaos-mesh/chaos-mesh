import React from 'react'
import { render } from '@testing-library/react'
import App from './App'

test('renders overview link', () => {
  const { getByText } = render(<App />)
  const linkElement = getByText(/overview/i)
  expect(linkElement).toBeInTheDocument()
})
