import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
    title: "Git Storm",
    description: "A changelog manager",
    markdown: {
        theme: {
            light: "catppuccin-latte",
            dark: "catppuccin-macchiato",
        },
    },
    themeConfig: {
        // https://vitepress.dev/reference/default-theme-config
        nav: [
            { text: "Home", link: "/" },
            { text: "Examples", link: "/markdown-examples" },
        ],

        sidebar: [
            {
                text: "Examples",
                items: [
                    { text: "Markdown Examples", link: "/markdown-examples" },
                    { text: "Runtime API Examples", link: "/api-examples" },
                ],
            },
        ],

        socialLinks: [
            { icon: "github", link: "https://github.com/vuejs/vitepress" },
        ],
    },
});
