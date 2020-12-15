export default function insertCommonStyle() {
  document.head.insertAdjacentHTML(
    'beforeend',
    `<style>
      .chaos-chart-dark text {
        fill: #fff;
      }

      .chaos-chart .axis path,
      .chaos-chart .axis line {
        stroke: rgba(0, 0, 0, 0.12);
      }

      .chaos-chart-dark .axis path,
      .chaos-chart-dark .axis line {
        stroke: rgba(255, 255, 255, 0.12);
      }

      .chaos-events-legends {
        position: absolute;
        top: 0;
        left: 0;
        display: flex;
        max-height: 36px;
        flex-wrap: wrap;
        overflow-y: scroll;
      }

      .chaos-events-legends > div {
        display: flex;
        align-items: center;
        margin-right: 1rem;
        cursor: pointer;
      }
    </style>`
  )
}
