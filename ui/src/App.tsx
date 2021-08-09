import { Provider } from 'react-redux'
import { BrowserRouter as Router } from 'react-router-dom'
import ThemeProvider from './ThemeProvider'
import TopContainer from 'components/TopContainer'
import store from './store'

const App = () => (
  <Provider store={store}>
    <Router>
      <ThemeProvider>
        <TopContainer />
      </ThemeProvider>
    </Router>
  </Provider>
)

export default App
