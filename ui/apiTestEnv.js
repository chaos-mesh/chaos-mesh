const JSDOMEnvironment = require('jest-environment-jsdom')

class APITestEnvironment extends JSDOMEnvironment {
  constructor(config, context) {
    super(config, context)
    this.global.mockTestFailed = false
  }
  async handleTestEvent(event, state) {
    if (event.name === 'test_fn_failure' && state.currentlyRunningTest.name === 'mock data identifier test') {
      this.global.mockTestFailed = true
    }
  }
}

module.exports = APITestEnvironment
