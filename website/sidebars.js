module.exports = {
  docs: [
    {
      type: 'doc',
      id: 'overview',
    },
    {
      type: 'category',
      label: 'Getting Started',
      items: ['installation/installation', 'installation/get_started_on_kind', 'installation/get_started_on_minikube'],
    },
    {
      type: 'category',
      label: 'User Guide',
      items: [
        'user_guides/run_chaos_experiment',
        'user_guides/pause_experiment',
        {
          type: 'category',
          label: 'Configure Chaos',
          items: [
            'user_guides/podchaos_experiment',
            'user_guides/networkchaos_experiment',
            'user_guides/stresschaos_experiment',
            'user_guides/timechaos_experiment',
            'user_guides/iochaos_experiment',
            'user_guides/kernelchaos_experiment',
          ],
        },
        'user_guides/experiment_scope',
        'user_guides/sidecar_configmap',
        'user_guides/sidecar_template',
      ],
    },
    {
      type: 'category',
      label: 'Use Cases',
      items: ['use_cases/multi_data_centers'],
    },
    {
      type: 'category',
      label: 'Development Guide',
      items: [
        'development_guides/development_overview',
        'development_guides/set_up_the_development_environment',
        'development_guides/develop_a_new_chaos',
      ],
    },
    {
      type: 'doc',
      id: 'faqs',
    },
    {
      type: 'category',
      label: 'Releases',
      items: ['releases/v0.8.0'],
    },
  ],
}
