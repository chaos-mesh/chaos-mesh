export default function insertCommonStyle() {
  document.head.insertAdjacentHTML(
    'beforeend',
    `<style>
      .chaos-chart text {
        fill: rgba(0, 0, 0, 0.54);
        font-weight: bold;
      }

      .chaos-chart-dark text {
        fill: #fff;
        font-weight: bold;
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
        max-width: 100%;
        max-height: 36px;
        flex-wrap: wrap;
      }

      .chaos-events-legends > div {
        display: flex;
        align-items: center;
        margin-left: 1rem;
        cursor: pointer;
      }
    </style>`
  )
}
