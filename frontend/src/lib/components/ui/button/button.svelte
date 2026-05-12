<script lang="ts" module>
  import { tv, type VariantProps } from "tailwind-variants";

  export const buttonVariants = tv({
    base: "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-2 focus-visible:outline-border-focus disabled:opacity-50 disabled:pointer-events-none",
    variants: {
      variant: {
        default: "bg-accent text-accent-fg hover:bg-accent-hover",
        secondary: "bg-bg-elevated text-fg-primary hover:bg-bg-hover",
        outline: "border border-border-default bg-transparent text-fg-primary hover:bg-bg-hover",
        ghost: "bg-transparent text-fg-primary hover:bg-bg-hover",
        destructive: "bg-fg-danger text-bg-base hover:opacity-90",
        link: "bg-transparent text-fg-link underline-offset-4 hover:underline",
      },
      size: {
        sm: "h-8 px-3 text-xs",
        md: "h-9 px-4 text-sm",
        lg: "h-10 px-5 text-base",
        icon: "h-8 w-8 p-0",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "md",
    },
  });

  export type ButtonVariant = NonNullable<VariantProps<typeof buttonVariants>["variant"]>;
  export type ButtonSize = NonNullable<VariantProps<typeof buttonVariants>["size"]>;
</script>

<script lang="ts">
  import type { HTMLButtonAttributes } from "svelte/elements";
  import { cn } from "$lib/utils";

  interface Props extends HTMLButtonAttributes {
    variant?: ButtonVariant;
    size?: ButtonSize;
    class?: string;
    children?: import("svelte").Snippet;
  }

  let {
    variant = "default",
    size = "md",
    class: className = "",
    children,
    ...rest
  }: Props = $props();
</script>

<button class={cn(buttonVariants({ variant, size }), className)} {...rest}>
  {@render children?.()}
</button>
