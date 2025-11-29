<script setup lang="ts">
import ky from "ky";
import { useQuery } from "@tanstack/vue-query";
import { apiBase, type Settings } from "@/api";
import LoadingScreen from "@/screens/LoadingScreen.vue";
import { onMounted, ref } from "vue";
import Telegram from "@/telegram.ts";
import SettingsDialog from "@/screens/SettingsDialog.vue";

const { apiKey } = defineProps<{
    apiKey: string;
}>();

const settingsOpen = ref(false);

onMounted(() => {
    Telegram.WebApp.SettingsButton.show();
    Telegram.WebApp.SettingsButton.onClick(() => {
        settingsOpen.value = true;
    });
})

const { data: count, isLoading: countLoading } = useQuery({
    queryKey: ['count'],
    queryFn: async () => {
        const resp = await ky.get(`${apiBase}/${apiKey}/count`).json() as { count: number };
        return resp.count;
    },
});

const { data: settings, isLoading: settingsLoading, refetch: settingsRefetch } = useQuery({
    queryKey: ['settings'],
    queryFn: async () => {
        const resp = await ky.get(`${apiBase}/${apiKey}/settings`).json() as { settings: Settings };
        return resp.settings;
    }
})
</script>

<template>
    <template v-if="countLoading || settingsLoading">
        <loading-screen/>
    </template>
    <div class="app-screen" v-else>
        <div class="content">
            <div class="title">Stats</div>
            <div class="grid">
                <div class="tile">
                    <span class="value">{{ count }}</span>
                    <span class="label">Unsent posts</span>
                </div>
            </div>
            <div class="title">Settings</div>
            <div class="grid">
                <div class="tile">
                    <span class="value">{{ settings!.group_threshold }}</span>
                    <span class="label">Group threshold</span>
                </div>
            </div>
        </div>
        <button class="btn" @click="settingsOpen = true">Settings</button>

        <settings-dialog
            v-model="settingsOpen"
            :settings="settings"
            :api-key="apiKey"
            @close="settingsOpen = false"
            @refetch="settingsRefetch"
        />
    </div>
</template>

<style scoped lang="sass">
.app-screen
    padding: 16px 24px
    width: 100vw
    height: 100vh
    display: flex
    flex-direction: column

    .content
        flex-grow: 1
        overflow: auto
        margin-bottom: 16px

.title
    font-size: 18px
    margin-bottom: 8px

.grid
    display: flex
    flex-direction: row
    flex-wrap: wrap
    gap: 8px

    &:not(:last-child)
        margin-bottom: 16px

    .tile
        width: 150px
        padding: 8px 12px
        border-radius: 8px
        border: 1px solid var(--tg-theme-hint-color)
        display: flex
        flex-direction: column
        align-items: center

        .value
            font-size: 48px
            font-weight: 600
            color: var(--tg-theme-accent-text-color)

</style>
