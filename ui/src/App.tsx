import { Provider } from 'react-redux'
import React from 'react'
import { BrowserRouter as Router } from 'react-router-dom'
import TopContainer from 'components/TopContainer'
import store from './store'

const App = () => (
  <Provider store={store}>
    <Router basename="/dashboard">
      <TopContainer />
    </Router>
  </Provider>
)

export default App
