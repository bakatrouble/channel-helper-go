<script setup lang="ts">

</script>

<template>
    <div class="fullscreen">
        <slot/>
    </div>
</template>

<style scoped lang="sass">
.fullscreen
    width: 100%
    height: 100%
    display: flex
    justify-content: center
    align-items: center
    flex-direction: column
</style>