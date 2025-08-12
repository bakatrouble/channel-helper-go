import { createApp } from 'vue'
import App from './App.vue'
import { VueQueryPlugin } from "@tanstack/vue-query";
import mdiVue from 'mdi-vue/v3';
import { mdiHandFrontLeft } from '@mdi/js';

createApp(App)
    .use(VueQueryPlugin)
    .use(mdiVue, {
        icons: {
            mdiHandFrontLeft,
        },
    })
    .mount('#app')
