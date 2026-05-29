import { createRouter, createWebHashHistory } from 'vue-router'
import GalleryView from '../views/GalleryView.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'gallery',
      component: GalleryView,
    },
  ],
})

export default router
