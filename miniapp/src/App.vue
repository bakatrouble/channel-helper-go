<script setup lang="ts">
import qs from 'qs';
import { useQuery } from "@tanstack/vue-query";
import ky from "ky";
import AuthenticatedApp from "@/AuthenticatedApp.vue";
import { apiBase } from "@/api.ts";

const initDataRaw = (window as any).Telegram.WebApp.initData;
// const initDataRaw = 'user=%7B%22id%22%3A98934915%2C%22first_name%22%3A%22bakatrouble%22%2C%22last_name%22%3A%22%E3%80%8C%E6%88%91%E6%80%9D%E3%81%86%E6%95%85%E3%81%AB%E6%88%91%E3%83%9F%E3%82%AB%E3%83%B3%E3%80%8D%22%2C%22username%22%3A%22bakatrouble%22%2C%22language_code%22%3A%22en%22%2C%22is_premium%22%3Atrue%2C%22allows_write_to_pm%22%3Atrue%2C%22photo_url%22%3A%22https%3A%5C%2F%5C%2Ft.me%5C%2Fi%5C%2Fuserpic%5C%2F320%5C%2FOxbE0VGvCH--tBq1jnYeP4l_euJ-uzmy_kOLinjbHxw.svg%22%7D&chat_instance=4166407415562333196&chat_type=sender&auth_date=1754928562&signature=YdpqL0jZiH-3ZnhUW6XPJZ8TDGjLjJrfOGXhmS2mvrVqWg9zEChDm_9drvkEmknXayBjOM4Z_CkOosvDtK0IBw&hash=e23db2f71c3822152417bc0af04a95b10995b5c8d2fc8cea2eb737b06777d7a2';
const initData = qs.parse(initDataRaw);

const { data: apiKey, isLoading: apiKeyLoading } = useQuery({
  queryKey: ['apiKey'],
  queryFn: async () => {
    const resp = await ky.get(`${apiBase}/apiKey`, {
      searchParams: {
        init_data: initDataRaw,
      },
    }).json() as { apiKey: string };
    return resp.apiKey;
  }
})

const user = JSON.parse(initData.user as string || '{}');
console.log(initData, user, apiKey, apiKeyLoading);
</script>

<template>
  <div v-if="apiKeyLoading">
    Loading...
  </div>
  <div v-else-if="apiKey">
    <authenticated-app :api-key="apiKey" />
  </div>
  <div v-else>
    Error loading API key.
  </div>
</template>

<style lang="sass">
*
  color: white
</style>
