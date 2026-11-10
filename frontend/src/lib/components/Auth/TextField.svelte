<script lang="ts">
  interface Props {
    id: string;
    label: string;
    type?: "text" | "password";
    value: string;
    placeholder?: string;
    disabled?: boolean;
    required?: boolean;
    autocomplete?: HTMLInputElement["autocomplete"];
    invalid?: boolean;
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
    autocomplete = undefined,
    invalid = false,
    onInput,
  }: Props = $props();

  function handle(event: Event) {
    const target = event.target as HTMLInputElement;
    onInput(target.value);
  }
</script>

<div>
  <label for={id} class="sg-label">{label}</label>
  <input
    {id}
    {type}
    {value}
    {placeholder}
    {disabled}
    {required}
    {autocomplete}
    aria-invalid={invalid || undefined}
    oninput={handle}
    class="sg-input sg-input-form"
  />
</div>
