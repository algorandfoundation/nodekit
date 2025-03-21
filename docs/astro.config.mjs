// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import tailwind from "@astrojs/tailwind";
import mdx from '@astrojs/mdx';
import {resolve} from "path";

// https://astro.build/config
export default defineConfig({
  integrations: [
    starlight({
      title: "NodeKit",
      logo: {
        light: "./public/nodekit-light.png",
        dark: "./public/nodekit-dark.png",
        alt: "NodeKit for Algorand",
        replacesTitle: true,
      },
      social: {
        github: "https://github.com/algorandfoundation/nodekit",
      },
      sidebar: [
        {
          label: 'Running A Node',
          collapsed: true,
          items: [
            {
              label: 'NodeKit Overview',
              link: 'nodes/nodekit-overview',
            },
            {
              label: 'NodeKit Quick Start',
              link: 'nodes/nodekit-quick-start',
            },
            {
              label: 'NodeKit Reference',
              link: 'nodes/nodekit-reference/commands',
            },
          ],
        },
      ],
      components: {
        ThemeProvider: "./src/components/CustomThemeProvider.astro",
      },
      disable404Route: true,
      customCss: ["./src/tailwind.css"],
    }),
    mdx(),
    tailwind({ applyBaseStyles: true }),
  ],
  vite: {
    resolve: {
      alias: {
        '@assets': resolve('./src/assets'),
        '@images': resolve('./src/assets/images'),
        '@diagrams': resolve('./src/assets/diagrams/svg'),
      },
    },

    // plugins: [tailwindcss()],
  },
});
