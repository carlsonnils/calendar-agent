import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

export default defineConfig({
    plugins: [svelte()],
    build: {
        outDir: "../../dist", // puts dist/ at calendar-agent/ root, next to your Go code
    },
    server: {
        proxy: {
            "/api": "http://localhost", // your Go server port
            "/login": "http://localhost", // your Go server port
        },
    },
});
