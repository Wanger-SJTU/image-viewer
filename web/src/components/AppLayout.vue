<script setup lang="ts">
</script>

<template>
  <div class="app-layout">
    <header class="app-header">
      <h1>Image Viewer</h1>
      <nav>
        <router-link to="/">Gallery</router-link>
      </nav>
    </header>
    <main class="app-main">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #1a1a2e;
  color: #eee;
}

.app-header {
  display: flex;
  align-items: center;
  gap: 2rem;
  padding: 0.5rem 1.5rem;
  background: #16213e;
  border-bottom: 1px solid #0f3460;
  height: 48px;
  flex-shrink: 0;
}

.app-header h1 {
  font-size: 1.2rem;
  font-weight: 600;
  margin: 0;
}

.app-header a {
  color: #e94560;
  text-decoration: none;
}

.app-main {
  flex: 1;
  overflow: hidden;
}
</style>
