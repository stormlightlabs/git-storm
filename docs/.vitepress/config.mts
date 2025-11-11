import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
// https://vitepress.dev/reference/default-theme-config
export default defineConfig({
    title: "Storm",
    description: "Local-first changelog manager for git repositories",
    markdown: {
        theme: {
            light: "catppuccin-latte",
            dark: "catppuccin-macchiato",
        },
    },
    themeConfig: {
        nav: [
            { text: "Introduction", link: "/introduction" },
            { text: "Quickstart", link: "/quickstart" },
            { text: "Manual", link: "/manual" },
            { text: "Development", link: "/development" },
        ],
        sidebar: [
            {
                text: "Getting Started",
                items: [
                    { text: "Introduction", link: "/introduction" },
                    { text: "Quickstart", link: "/quickstart" },
                ],
            },
            {
                text: "Reference",
                items: [
                    { text: "Manual", link: "/manual" },
                    { text: "Development", link: "/development" },
                ],
            },
        ],
        socialLinks: [
            {
                icon: "github",
                link: "https://github.com/stormlightlabs/git-storm",
            },
            {
                icon: "bluesky",
                link: "http://bsky.app/profile/desertthunder.dev/",
            },
        ],
    },
    base: "/git-storm/",
});
