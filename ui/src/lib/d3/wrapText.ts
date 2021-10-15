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
import * as d3 from 'd3'

export default function wrap(texts: d3.Selection<d3.BaseType, unknown, SVGElement, unknown>, width: number) {
  texts.each(function () {
    const text = d3.select(this)
    const words = text.text().split(/\s+/).reverse()
    let word
    let line: string[] = []
    let lineNumber = 0
    const lineHeight = 1.1
    const y = text.attr('y')
    const dy = parseFloat(text.attr('dy'))
    let tspan = text
      .text(null)
      .append('tspan')
      .attr('x', 0)
      .attr('y', y)
      .attr('dy', dy + 'em')

    while ((word = words.pop())) {
      line.push(word)
      tspan.text(line.join(' '))

      if (tspan.node()!.getComputedTextLength() > width) {
        line.pop()
        tspan.text(line.join(' '))
        line = [word]
        tspan = text
          .append('tspan')
          .attr('x', 0)
          .attr('y', y)
          .attr('dy', `${++lineNumber * lineHeight + dy}em`)
          .text(word)
      }
    }
  })
}
