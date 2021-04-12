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
    </style>`
  )
}
