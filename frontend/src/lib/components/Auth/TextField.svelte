<script lang="ts">
  interface Props {
    id: string;
    label: string;
    type?: "text" | "password";
    value: string;
    placeholder?: string;
    disabled?: boolean;
    required?: boolean;
    onInput: (next: string) => void;
  }

  let {
    id,
    label,
    type = "text",
    value,
    placeholder = "",
    disabled = false,
    required = false,
    onInput,
  }: Props = $props();

  function handle(event: Event) {
    const target = event.target as HTMLInputElement;
    onInput(target.value);
  }
</script>

<div>
  <label for={id} class="block text-[11px] text-fg-muted mb-1">{label}</label>
  <input
    {id}
    {type}
    {value}
    {placeholder}
    {disabled}
    {required}
    oninput={handle}
    class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-primary text-sm focus:border-border-focus outline-none disabled:opacity-60"
  />
</div>
