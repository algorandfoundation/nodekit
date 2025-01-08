// @ts-check
import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";
import tailwind from "@astrojs/tailwind";

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
          label: "Troubleshooting",
          link: "/troubleshooting",
        },
        {
          label: "Command Reference",
          collapsed: true,
          autogenerate: { directory: "reference" },
        },
      ],
      components: {
        ThemeProvider: "./src/components/CustomThemeProvider.astro",
      },
      customCss: ["./src/tailwind.css"],
    }),
    tailwind({ applyBaseStyles: true }),
  ],
});
