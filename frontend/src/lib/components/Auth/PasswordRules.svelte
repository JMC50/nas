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

<div class="p-3 rounded-md bg-bg-elevated border border-border-default">
  <div class="text-[11px] text-fg-muted mb-1.5">Password requirements</div>
  <ul class="text-xs text-fg-secondary space-y-0.5">
    <li>• At least {rules.minLength} characters</li>
    {#if rules.requireUppercase}<li>• Uppercase letter</li>{/if}
    {#if rules.requireLowercase}<li>• Lowercase letter</li>{/if}
    {#if rules.requireNumber}<li>• Number</li>{/if}
    {#if rules.requireSpecial}<li>• Special character</li>{/if}
  </ul>
</div>
