import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import { ngrok } from "vite-plugin-ngrok";

// https://vite.dev/config/
export default defineConfig({
    base: '/miniapp/',
    plugins: [
        vue(),
        vueDevTools(),
        ngrok({
            authtoken: process.env.NGROK_AUTHTOKEN,
            domain: 'driving-monitor-marginally.ngrok-free.app',
        }),
    ],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url))
        },
    },
    server: {
        allowedHosts: [
            'driving-monitor-marginally.ngrok-free.app',
        ],
        proxy: {
            '/api': {
                target: "http://localhost:8001",
                changeOrigin: true,
                secure: false,
                rewrite: path => path.replace('/api', ''),
            }
        },
    },
})
