<script setup lang="ts">
import { apiBase, type Settings } from "@/api.ts";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useForm } from "vee-validate";
import { watch } from "vue";
import ky from "ky";
import { useMutation } from "@tanstack/vue-query";

const { settings, apiKey } = defineProps<{
    settings?: Settings;
    apiKey?: string;
}>();
const open = defineModel<boolean>();

const emit = defineEmits<{
    close: [];
    refetch: [];
}>();

const validationSchema = toTypedSchema(z.object({
    group_threshold: z.int().min(0),
}))
const { values, setValues, errors, defineField, handleSubmit } = useForm<Settings>({ validationSchema });
const [ groupThreshold, groupThresholdAttrs ] = defineField('group_threshold');

watch(open, newOpen => {
    if (newOpen && settings) {
        setValues(settings);
    }
});

const { mutateAsync: save, isPending: savePending } = useMutation({
    mutationFn: async () => {
        await ky.post(`${apiBase}/${apiKey}/settings`, {
            json: values,
        });
    },
    onSuccess: () => {
        emit('refetch');
        emit('close');
    },
})

const submit = handleSubmit(() => save());
</script>

<template>
    <div class="fade" :data-active="open" @click="emit('close')"/>
    <div class="dialog-wrapper" :data-visible="open">
        <div class="dialog">
            <div class="title">Settings</div>
            <form @submit.prevent="submit">
                <div class="controls">
                    <label>Group threshold</label>
                    <input :class="errors.group_threshold && 'error'" type="number" v-model="groupThreshold" v-bind="groupThresholdAttrs" >
                    <label v-if="errors.group_threshold" class="error">{{ errors.group_threshold }}</label>
                </div>
                <div class="actions">
                    <button class="btn outline" @click="emit('close')">Close</button>
                    <button class="btn" :disabled="savePending" type="submit">Save</button>
                </div>
            </form>
        </div>
    </div>
</template>

<style scoped lang="sass">
.fade
    pointer-events: none
    position: fixed
    top: 0
    left: 0
    bottom: 0
    right: 0
    background: #0006
    opacity: 0
    transition: opacity 0.3s

    &[data-active="true"]
        opacity: 1
        pointer-events: auto

.dialog-wrapper
    flex-direction: column
    position: fixed
    top: 0
    left: 0
    bottom: 0
    right: 0
    justify-content: center
    align-items: center
    display: none
    padding: 48px 32px

    &[data-visible="true"]
        display: flex

    .dialog
        width: 100%
        max-width: 400px
        background-color: var(--tg-theme-bg-color)
        padding: 16px 24px
        border-radius: 8px

        .title
            margin-bottom: 16px
            font-size: 18px

        .controls
            margin-bottom: 16px

            label
                font-size: 12px
                color: var(--tg-theme-hint-color)

                &.error
                    color: red

            input
                width: 100%
                padding: 8px
                background-color: transparent
                border-radius: 4px
                border: 1px solid var(--tg-theme-hint-color)
                color: white
                font-size: 16px
                outline: none

                &:active, &:focus
                    border-color: var(--tg-theme-accent-text-color)

                &.error
                    border-color: red

        .actions
            display: flex
            gap: 8px


</style>
