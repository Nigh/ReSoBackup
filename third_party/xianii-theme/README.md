# Xianii theme

Theme source: https://github.com/Nigh/xianii-theme

Pinned source: `07b10ae25e587f4a4c3b74a6ab9df08b8bec8c50`, design-system 2.4.0.

`tokens.css` is the unmodified upstream source of truth; `theme.go` maps its dark/light colors, alpha values and metrics to mygo Native UI. No CSS is loaded by the application. The sans token uses the platform sans-serif font with Chinese/emoji fallback. No font files are downloaded or bundled. Hover/pressed colors are derived from the theme's interaction tokens for native widgets. When updating the source, update the Go mapping and its tests together.

The upstream MIT license is retained in `LICENSE`. The application's license remains LGPL-2.1.
