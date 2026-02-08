<script lang="ts">
  import { onMount } from "svelte";
  import { fly } from "svelte/transition";

  interface CalendarEvent {
    UID: string;
    HumanStart: string;
    Summary: string;
    Description: string;
  }

  interface ToastItem extends CalendarEvent {
    id: string;
  }

  // Svelte 5 Runes
  let toasts = $state<ToastItem[]>([]);
  const toastLimit = 10;

  function removeToast(id: string) {
    toasts = toasts.filter((t) => t.id !== id);
  }

  function addToast(event: CalendarEvent) {
    const id = `toast-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
    const newToast = { ...event, id };

    // Add to front of list (top of stack)
    toasts = [newToast, ...toasts].slice(0, toastLimit);

    // Auto-remove after 5 seconds
    setTimeout(() => {
      removeToast(id);
    }, 5000);
  }

  $effect(() => {
    const handler = (e: CustomEvent<CalendarEvent[]>) => {
      const events = e.detail;
      if (!Array.isArray(events)) return;

      events.forEach((event, index) => {
        setTimeout(() => {
          addToast(event);
        }, index * 200);
      });
    };

    window.addEventListener("jsoon:show-events", handler as EventListener);
    return () =>
      window.removeEventListener("jsoon:show-events", handler as EventListener);
  });
</script>

<!-- Notification layer: pointer-events-none container, pointer-events-auto items -->
<div
  class="fixed top-24 right-6 z-[9999] flex flex-col items-end pointer-events-none space-y-4"
>
  {#each toasts as toast (toast.id)}
    <div
      transition:fly={{ x: 300, duration: 300 }}
      class="neo-border neo-shadow-lg p-4 bg-white dark:bg-neo-dark w-80 pointer-events-auto relative"
    >
      <div class="flex justify-between items-start mb-2">
        <span
          class="bg-neo-purple text-white dark:bg-neo-pink dark:text-black px-2 py-0.5 text-xs font-black uppercase tracking-tight truncate max-w-[200px]"
        >
          {toast.HumanStart}
        </span>
        <button
          onclick={() => removeToast(toast.id)}
          class="text-xl leading-none hover:scale-110 transition-transform px-1 font-black cursor-pointer text-black dark:text-white"
        >
          ×
        </button>
      </div>
      <h3
        class="font-black uppercase mb-1 leading-tight text-black dark:text-white"
      >
        {toast.Summary}
      </h3>
      <p
        class="text-sm font-bold opacity-80 line-clamp-3 text-black dark:text-white"
      >
        {toast.Description}
      </p>
    </div>
  {/each}
</div>
