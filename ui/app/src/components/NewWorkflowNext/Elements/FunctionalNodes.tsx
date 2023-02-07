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
import Space from '@ui/mui-extends/esm/Space'

import { SpecialTemplateType } from '../utils/convert'
import DraggableBareNode from './DraggableBareNode'
import { ElementTypes, ElementsProps } from './types'

export default function FunctionalNodes({ onElementClick }: ElementsProps) {
  return (
    <Space>
      <DraggableBareNode
        elementType={SpecialTemplateType.Serial}
        kind={SpecialTemplateType.Serial}
        onNodeClick={onElementClick}
      >
        Serial
      </DraggableBareNode>
      <DraggableBareNode
        elementType={SpecialTemplateType.Parallel}
        kind={SpecialTemplateType.Parallel}
        onNodeClick={onElementClick}
      >
        Parallel
      </DraggableBareNode>
      <DraggableBareNode elementType={ElementTypes.Suspend} kind="Suspend" onNodeClick={onElementClick}>
        Suspend
      </DraggableBareNode>
    </Space>
  )
}
