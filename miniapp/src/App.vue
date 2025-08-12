<script setup lang="ts">
import { useQuery } from "@tanstack/vue-query";
import ky, { HTTPError } from "ky";
import { onMounted, ref } from "vue";
import Telegram from "@/telegram.ts";
import MainAppScreen from "@/screens/MainAppScreen.vue";
import { apiBase } from "@/api.ts";
import LoadingScreen from "@/screens/LoadingScreen.vue";
import GtfoScreen from "@/screens/GtfoScreen.vue";
import type { WebAppInitData } from "telegram-web-app";

const themeParams = ref(Telegram.WebApp.themeParams);
const initDataRaw = ref('');
const initData = ref<WebAppInitData>(Telegram.WebApp.initDataUnsafe);

onMounted(() => {
    Telegram.WebApp.onEvent('themeChanged', () => {
        themeParams.value = Telegram.WebApp.themeParams;
        console.log('Theme changed:', Telegram.WebApp.themeParams);
    })

    let initDataRawValue = Telegram.WebApp.initData ?? '';
    if (!initDataRawValue && import.meta.env.DEV) {
        initDataRawValue = 'user=%7B%22id%22%3A98934915%2C%22first_name%22%3A%22bakatrouble%22%2C%22last_name%22%3A%22%E3%80%8C%E6%88%91%E6%80%9D%E3%81%86%E6%95%85%E3%81%AB%E6%88%91%E3%83%9F%E3%82%AB%E3%83%B3%E3%80%8D%22%2C%22username%22%3A%22bakatrouble%22%2C%22language_code%22%3A%22en%22%2C%22is_premium%22%3Atrue%2C%22allows_write_to_pm%22%3Atrue%2C%22photo_url%22%3A%22https%3A%5C%2F%5C%2Ft.me%5C%2Fi%5C%2Fuserpic%5C%2F320%5C%2FOxbE0VGvCH--tBq1jnYeP4l_euJ-uzmy_kOLinjbHxw.svg%22%7D&chat_instance=4166407415562333196&chat_type=sender&auth_date=1754928562&signature=YdpqL0jZiH-3ZnhUW6XPJZ8TDGjLjJrfOGXhmS2mvrVqWg9zEChDm_9drvkEmknXayBjOM4Z_CkOosvDtK0IBw&hash=e23db2f71c3822152417bc0af04a95b10995b5c8d2fc8cea2eb737b06777d7a2';
    }
    initDataRaw.value = initDataRawValue;
});

const { data: apiKey, isLoading: apiKeyLoading, error: apiKeyError } = useQuery({
    queryKey: ['apiKey', initDataRaw],
    queryFn: async () => {
        if (!initDataRaw.value)
            return null;
        const resp = await ky.get(`${apiBase}/apiKey`, {
            searchParams: {
                init_data: initDataRaw.value,
            },
        }).json() as { apiKey: string };
        return resp.apiKey;
    },
    retry: (failureCount, e) => {
        if (e instanceof HTTPError && ([403, 400].includes(e.response?.status))) {
            return false;
        }
        return failureCount < 3;
    }
});
</script>

<template>
    <div class="app-wrapper">
        <template v-if="!initDataRaw || apiKeyLoading">
            <loading-screen/>
        </template>
        <template v-else-if="apiKey">
            <main-app-screen :api-key="apiKey"/>
        </template>
        <template v-else-if="[400, 403].includes((apiKeyError as HTTPError)?.response?.status)">
            <gtfo-screen />
        </template>
        <template v-else>
            Error loading API key.
        </template>
    </div>
</template>

<style lang="sass">
*
    box-sizing: border-box

body, html
    margin: 0
    padding: 0
    width: 100vw
    height: 100vh
    overflow: hidden

.app-wrapper
    width: 100vw
    height: 100vh
    display: flex
    background: var(--tg-theme-bg-color)
    color: var(--tg-theme-text-color)
</style>
