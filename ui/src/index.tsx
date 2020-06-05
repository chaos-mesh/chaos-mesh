import * as serviceWorker from './serviceWorker'

import App from './App'
import React from 'react'
import ReactDOM from 'react-dom'

// Temporarily disable React.StrictMode in dev and prod.
// Progress: https://github.com/mui-org/material-ui/issues/13394
ReactDOM.render(<App />, document.getElementById('root'))

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister()
