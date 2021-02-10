module.exports = {
  title: "driftctl",
  tagline: "Track infrastructure drift - driftctl",
  url: "https://docs.driftctl.com",
  baseUrl: "/",
  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",
  favicon: "img/favicon.ico",
  organizationName: "cloudskiff",
  projectName: "driftctl",
  themeConfig: {
    navbar: {
      title: "driftctl",
      logo: {
        alt: "driftctl Logo",
        src: "img/logo.svg",
        srcDark: "img/dark-logo.svg",
      },
      items: [
        {
          to: "/",
          activeBasePath: "/",
          label: "Docs",
          position: "left",
        },
        {
          type: "docsVersionDropdown",
          position: "right",
          dropdownActiveClassDisabled: true,
          dropdownItemsAfter: [
            {
              to: "/versions",
              label: "All versions",
            },
          ],
        },
        {
          href: "https://discord.gg/NMCBxtD7Nd",
          label: "Discord",
          position: "right",
          "aria-label": "Discord server",
        },
        {
          href: "https://twitter.com/getdriftctl",
          label: "Twitter",
          position: "right",
          "aria-label": "Twitter account",
        },
        {
          href: "https://github.com/cloudskiff/driftctl",
          label: "GitHub",
          position: "right",
          "aria-label": "GitHub repository",
        },
      ],
    },
    footer: {
      style: "dark",
      copyright: `Copyright © ${new Date().getFullYear()} CloudSkiff.`,
    },
  },
  presets: [
    [
      "@docusaurus/preset-classic",
      {
        docs: {
          path: "../docs",
          routeBasePath: "/",
          sidebarPath: require.resolve("./sidebars.js"),
          editUrl: "https://github.com/cloudskiff/driftctl",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
      },
    ],
  ],
};
