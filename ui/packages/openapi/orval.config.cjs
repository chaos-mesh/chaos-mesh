module.exports = {
  openapi: {
    input: './swagger.yaml',
    output: {
      mode: 'split',
      target: '../../app/src/openapi/index.ts',
      client: 'react-query',
      override: {
        mutator: {
          path: '../../app/src/api/http.ts',
          name: 'customInstance',
        },
      },
    },
  },
}
