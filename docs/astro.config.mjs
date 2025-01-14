// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import tailwind from "@astrojs/tailwind";
import mdx from '@astrojs/mdx';

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
          label: "Guides",
          autogenerate: { directory: "guides" },
        },
        {
          label: "Command Reference",
          collapsed: true,
          autogenerate: { directory: "reference" },
        },
        {
          label: "Source Code",
          link: "https://github.com/algorandfoundation/nodekit/",
        },
        {
          label: "Issue Tracker",
          link: "https://github.com/algorandfoundation/nodekit/issues",
        },
        {
          label: "Report an issue",
          link: "https://github.com/algorandfoundation/nodekit/issues/new/choose",
        }
      ],
      components: {
        ThemeProvider: "./src/components/CustomThemeProvider.astro",
      },
      customCss: ["./src/tailwind.css"],
    }),
    mdx(),
    tailwind({ applyBaseStyles: true }),
  ],
});
