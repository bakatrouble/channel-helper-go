<script setup lang="ts">
import ky from "ky";
import { useQuery } from "@tanstack/vue-query";
import { apiBase } from "@/api.ts";
import LoadingScreen from "@/screens/LoadingScreen.vue";

const { apiKey } = defineProps<{
  apiKey: string;
}>();

const { data: count, isLoading: countLoading } = useQuery({
  queryKey: ['count'],
  queryFn: async () => {
    const resp = await ky.get(`${apiBase}/${apiKey}/count`).json() as { count: number };
    return resp.count;
  },
});
</script>

<template>
  <template v-if="countLoading">
    <loading-screen />
  </template>
  <div class="app-screen" v-else>
    Unsent posts count: {{ count }}
  </div>
</template>

<style scoped lang="sass">
.app-screen
    padding: 16px 24px
</style>
