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
export default function insertCommonStyle() {
  document.head.insertAdjacentHTML(
    'beforeend',
    `
    <style>
      .chaos-chart.dark text {
        fill: #fff;
      }

      .chaos-chart .axis path,
      .chaos-chart .axis line {
        stroke: rgba(0, 0, 0, .12);
      }

      .chaos-chart.dark .axis path,
      .chaos-chart.dark .axis line {
        stroke: rgba(255, 255, 255, .12);
      }

      .chaos-events-legends {
        position: absolute;
        top: 0;
        right: -15px;
        width: 150px;
        height: 100%;
        overflow-y: scroll;
      }

      @media screen and (max-width: 768px) {
        .chaos-events-legends {
          display: none;
        }
      }

      .chaos-events-legends > div {
        display: flex;
        align-items: center;
        cursor: pointer;
      }

      .chaos-events-legends .square {
        width: 12px;
        height: 12px;
        border-radius: 50%;
      }

      .chaos-events-legends .experiment {
        margin-left: .375rem;
        color: rgba(0, 0, 0, .54);
        font-size: .75rem;
        font-weight: bold;
      }

      .chaos-events-legends.dark .experiment {
        color: rgba(255, 255, 255, .54);
      }

      .chaos-event-tooltip {
        position: absolute;
        top: 0;
        left: 0;
        padding: .75rem;
        background: #fafafa;
        font: 1rem;
        border: 1px solid rgba(0, 0, 0, .12);
        border-radius: 4px;
        opacity: 0;
        transition: top .25s ease, left .25s ease;
        z-index: 999;
      }

      .chaos-event-tooltip.dark {
        background: #303030;
        border: 1px solid rgba(255, 255, 255, .12);
      }

      .chaos-event-tooltip .secondary {
        color: rgba(0, 0, 0, .54);
      }

      .chaos-event-tooltip.dark .secondary {
        color: rgba(255, 255, 255, .54);
      }
    </style>`
  )
}
