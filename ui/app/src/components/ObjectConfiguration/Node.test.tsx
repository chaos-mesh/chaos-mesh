/*
 * Copyright 2021 Chaos Mesh Authors.
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
import { render, screen } from 'test-utils'

import NodeConfiguration from './Node'

describe('<NodeConfiguration />', () => {
  describe('SimpleNode templates', () => {
    it('renders Suspend template correctly', () => {
      const template = {
        name: 'suspend-node',
        templateType: 'Suspend',
        deadline: '5m',
      }

      render(<NodeConfiguration template={template} />)

      expect(screen.getByText('suspend-node')).toBeInTheDocument()
      expect(screen.getByText('Suspend')).toBeInTheDocument()
      expect(screen.getByText('5m')).toBeInTheDocument()
    })

    it('renders StatusCheck template correctly', () => {
      const template = {
        name: 'status-check-node',
        templateType: 'StatusCheck',
      }

      render(<NodeConfiguration template={template} />)

      expect(screen.getByText('status-check-node')).toBeInTheDocument()
      expect(screen.getByText('StatusCheck')).toBeInTheDocument()
    })

    it('renders Schedule template correctly', () => {
      const template = {
        name: 'schedule-node',
        templateType: 'Schedule',
      }

      render(<NodeConfiguration template={template} />)

      expect(screen.getByText('schedule-node')).toBeInTheDocument()
      expect(screen.getByText('Schedule')).toBeInTheDocument()
    })

    it('renders without deadline when not provided', () => {
      const template = {
        name: 'simple-node',
        templateType: 'Suspend',
      }

      render(<NodeConfiguration template={template} />)

      expect(screen.getByText('simple-node')).toBeInTheDocument()
      expect(screen.queryByText('Deadline')).not.toBeInTheDocument()
    })
  })

  describe('TaskNode template', () => {
    it('renders Task template with container info correctly', () => {
      const template = {
        name: 'task-node',
        templateType: 'Task',
        task: {
          container: {
            name: 'test-container',
            image: 'nginx:latest',
          },
        },
      }

      render(<NodeConfiguration template={template} />)

      expect(screen.getByText('task-node')).toBeInTheDocument()
      expect(screen.getByText('test-container')).toBeInTheDocument()
      expect(screen.getByText('nginx:latest')).toBeInTheDocument()
    })

    it('renders Task template with command correctly', () => {
      const template = {
        name: 'task-with-command',
        templateType: 'Task',
        task: {
          container: {
            name: 'busybox-container',
            image: 'busybox:1.35',
            command: ['/bin/sh', '-c', 'echo hello'],
          },
        },
      }

      render(<NodeConfiguration template={template} />)

      expect(screen.getByText('task-with-command')).toBeInTheDocument()
      expect(screen.getByText('busybox-container')).toBeInTheDocument()
      expect(screen.getByText('busybox:1.35')).toBeInTheDocument()
    })

    it('renders Task template with conditional branches correctly', () => {
      const template = {
        name: 'task-with-branches',
        templateType: 'Task',
        task: {
          container: {
            name: 'main-container',
            image: 'alpine:3.18',
          },
        },
        conditionalBranches: [
          {
            target: 'success-node',
            expression: 'exitCode == 0',
          },
          {
            target: 'failure-node',
            expression: 'exitCode != 0',
          },
        ],
      }

      render(<NodeConfiguration template={template} />)

      expect(screen.getByText('success-node')).toBeInTheDocument()
      expect(screen.getByText('exitCode == 0')).toBeInTheDocument()
      expect(screen.getByText('failure-node')).toBeInTheDocument()
      expect(screen.getByText('exitCode != 0')).toBeInTheDocument()
    })
  })
})
