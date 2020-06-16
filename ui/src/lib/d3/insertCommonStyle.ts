export default function insertCommonStyle() {
  document.head.insertAdjacentHTML(
    'beforeend',
    `<style>
      .chaos-events-chart text {
        fill: rgba(0, 0, 0, 0.72);
      }

      .chaos-events-chart .axis path,
      .chaos-events-chart .axis line {
        stroke: rgba(0, 0, 0, 0.36);
      }

      .chaos-events-legends {
        position: absolute;
        top: 0;
        left: 50%;
        display: flex;
        transform: translateX(-50%);
      }

      .chaos-events-legends > div {
        display: flex;
        align-items: center;
        margin-right: 12px;
      }

      .chaos-events-legends > div:last-child {
        margin-right: 0;
      }
    </style>`
  )
}
