/*
 * Copyright 2025 Chaos Mesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
import { defineConfig } from 'orval'

export default defineConfig({
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
        query: {
          version: 5,
          options: {
            retry: 1,
            retryDelay: 3000,
          },
        },
        mock: {
          delay: 0,
          required: true,
        },
      },
      mock: true,
    },
  },
})
