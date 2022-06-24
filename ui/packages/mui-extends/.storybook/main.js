module.exports = {
  stories: ['../stories/**/*.stories.mdx', '../stories/**/*.stories.@(js|jsx|ts|tsx)'],
  addons: ['@storybook/addon-links', '@storybook/addon-essentials', '@storybook/addon-interactions'],
  framework: '@storybook/react',
  features: {
    emotionAlias: false, // https://stackoverflow.com/questions/70253373/mui-v5-storybook-theme-and-font-family-do-not-work-in-storybook
  },
}
