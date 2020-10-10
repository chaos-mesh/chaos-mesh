module.exports = {
  docs: [
    {
      type: 'doc',
      id: 'overview/what_is_chaos_mesh',
    },
    {
      type: 'category',
      label: 'Overview',
      items: ['overview/what_is_chaos_mesh', 'overview/features', 'overview/architecture'],
    },
    {
      type: 'category',
      label: 'Concepts',
      items: ['concepts/chaos_engineering','concepts/blast_radius'],
    },
    {
      type: 'category',
      label: 'Getting Started',
      items: ['get_started/installation', 'get_started/get_started_on_kind', 'get_started/get_started_on_minikube'],
    },
    {
      type: 'category',
      label: 'User Guides',
      items: [
        'user_guides/run_chaos_experiment',
        'user_guides/experiment_scope',
      ],
    },
    {
      type: 'category',
      label: 'Chaos Experiments',
      items: [
        'chaos_experiments/podchaos_experiment',
        'chaos_experiments/networkchaos_experiment',
        'chaos_experiments/stresschaos_experiment',
        'chaos_experiments/timechaos_experiment',
        'chaos_experiments/iochaos_experiment',
        'chaos_experiments/kernelchaos_experiment',
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
      items: ['releases/v1.0.0', 'releases/v0.9.0', 'releases/v0.8.0'],
    },
  ],
}
