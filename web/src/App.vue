<script setup lang="ts">
</script>

<template>
  <div class="app">
    <header class="header">
      <h1>GoMath</h1>
      <p>数学题答案解析</p>
    </header>
    <main>
      <RouterView />
    </main>
  </div>
</template>

<style scoped>
.app {
  min-height: 100vh;
}
.header {
  padding: 1rem 2rem;
  border-bottom: 1px solid #eee;
}
.header h1 {
  margin: 0;
  font-size: 1.5rem;
}
.header p {
  margin: 0.25rem 0 0;
  color: #666;
  font-size: 0.9rem;
}
main {
  padding: 1.5rem 2rem;
}
</style>
