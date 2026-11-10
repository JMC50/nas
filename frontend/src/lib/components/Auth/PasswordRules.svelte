<script lang="ts">
  interface Rules {
    minLength: number;
    requireUppercase: boolean;
    requireLowercase: boolean;
    requireNumber: boolean;
    requireSpecial: boolean;
  }

  interface Props {
    rules: Rules;
  }

  let { rules }: Props = $props();
</script>

<div class="p-3 rounded-control border border-border-default bg-bg-base">
  <div class="text-xs font-semibold text-fg-secondary mb-1.5">Password requirements</div>
  <ul class="text-xs text-fg-secondary flex flex-col gap-1 list-disc pl-4 marker:text-fg-muted">
    <li>At least {rules.minLength} characters</li>
    {#if rules.requireUppercase}<li>Uppercase letter</li>{/if}
    {#if rules.requireLowercase}<li>Lowercase letter</li>{/if}
    {#if rules.requireNumber}<li>Number</li>{/if}
    {#if rules.requireSpecial}<li>Special character</li>{/if}
  </ul>
</div>
