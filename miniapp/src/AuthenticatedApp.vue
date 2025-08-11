<script setup lang="ts">
import ky from "ky";
import { useQuery } from "@tanstack/vue-query";
import { apiBase } from "@/api.ts";

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
  <div v-if="countLoading">
    Loading...
  </div>
  <div v-else>
    Unsent posts count: {{ count }}
  </div>
</template>

<style scoped lang="sass">

</style>
