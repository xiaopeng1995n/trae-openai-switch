<script>
  import { cn } from '../../lib/utils.js';
  export let className = '';
</script>

<div
  class={cn('rounded-lg border bg-card text-card-foreground shadow-sm', className)}
  {...$$restProps}
>
  <slot />
</div>