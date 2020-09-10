import { Provider } from 'react-redux'
import React from 'react'
import { BrowserRouter as Router } from 'react-router-dom'
import { ThemeProvider } from '@material-ui/core/styles'
import TopContainer from 'components/TopContainer'
import chaosMeshTheme from 'theme'
import store from './store'

const App = () => (
  <ThemeProvider theme={chaosMeshTheme}>
    <Provider store={store}>
      <Router basename="/dashboard">
        <TopContainer />
      </Router>
    </Provider>
  </ThemeProvider>
)

export default App
