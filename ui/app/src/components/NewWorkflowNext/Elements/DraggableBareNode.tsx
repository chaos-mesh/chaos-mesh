/*
 * Copyright 2022 Chaos Mesh Authors.
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
import { useDrag } from 'react-dnd'

import BareNode from '../BareNode'
import type { BareNodeProps } from '../BareNode'
import { SpecialTemplateType } from '../utils/convert'
import { ElementTypes } from './types'

interface DraggableBareNodeProps extends BareNodeProps {
  elementType: ElementTypes | SpecialTemplateType
  act?: string

  /**
   * A special click handler that receives `kind` and `act` as arguments.
   *
   * It will be covered by the original `onClick` event.
   *
   * @memberof DraggableBareNodeProps
   */
  onNodeClick?: (kind: string, act?: string) => void
}

const DraggableBareNode = ({ elementType, kind, act, onNodeClick, sx, ...rest }: DraggableBareNodeProps) => {
  const [{ isDragging }, drag] = useDrag(() => ({
    type: elementType,
    item: {
      kind,
      act,
    },
    collect: (monitor) => ({
      isDragging: monitor.isDragging(),
    }),
  }))

  return (
    <BareNode
      kind={kind}
      sx={{ cursor: isDragging ? 'grab' : 'pointer' }}
      {...(kind &&
        onNodeClick && {
          onClick: () => onNodeClick(kind, act),
        })}
      {...rest}
      ref={drag}
    />
  )
}

export default DraggableBareNode
