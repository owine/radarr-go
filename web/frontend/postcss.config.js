export default {
  plugins: {
    'postcss-preset-env': {
      stage: 1,
      features: {
        'custom-properties': true,
        'custom-media-queries': true,
        'nesting-rules': true,
        'custom-selectors': true,
      },
    },
    'autoprefixer': {},
  },
}