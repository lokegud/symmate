<script lang="ts">
  interface Props {
    type?: 'text' | 'email' | 'password' | 'search' | 'url' | 'tel' | 'number';
    value?: string;
    placeholder?: string;
    label?: string;
    hint?: string;
    error?: string;
    disabled?: boolean;
    required?: boolean;
    id?: string;
    name?: string;
    autocomplete?: string;
    oninput?: (e: Event) => void;
    onchange?: (e: Event) => void;
  }

  let {
    type = 'text',
    value = $bindable(''),
    placeholder,
    label,
    hint,
    error,
    disabled = false,
    required = false,
    id,
    name,
    autocomplete,
    oninput,
    onchange,
  }: Props = $props();

  const inputId = id ?? `input-${Math.random().toString(36).slice(2)}`;
</script>

<div class="field" class:field--error={!!error} class:field--disabled={disabled}>
  {#if label}
    <label class="field__label" for={inputId}>
      {label}
      {#if required}<span class="field__required" aria-label="required">*</span>{/if}
    </label>
  {/if}

  <input
    {type}
    {name}
    {placeholder}
    {disabled}
    {required}
    {autocomplete}
    id={inputId}
    bind:value
    class="field__input"
    aria-describedby={hint || error ? `${inputId}-hint` : undefined}
    aria-invalid={!!error}
    oninput={oninput}
    onchange={onchange}
  />

  {#if error || hint}
    <p
      class="field__hint"
      class:field__hint--error={!!error}
      id="{inputId}-hint"
      role={error ? 'alert' : undefined}
    >
      {error ?? hint}
    </p>
  {/if}
</div>

<style>
  .field {
    display: flex;
    flex-direction: column;
    gap: var(--stack-xs);
  }

  .field__label {
    font-family: var(--font-body);
    font-size: var(--text-sm);
    font-weight: var(--weight-medium);
    color: var(--color-text);
    cursor: pointer;
  }

  .field__required {
    color: var(--color-error);
    margin-left: 2px;
  }

  .field__input {
    height: 40px;
    padding: 0 var(--inset-md);
    background: var(--color-surface);
    border: 1.5px solid var(--color-border);
    border-radius: var(--radius-md);
    font-family: var(--font-body);
    font-size: var(--text-base);
    color: var(--color-text);
    outline: none;
    transition:
      border-color var(--duration-micro) var(--ease-out),
      box-shadow var(--duration-micro) var(--ease-out);
    width: 100%;
  }

  .field__input::placeholder {
    color: var(--color-text-subtle);
  }

  .field__input:focus {
    border-color: var(--color-accent);
    box-shadow: 0 0 0 3px var(--color-accent-subtle);
  }

  .field--error .field__input {
    border-color: var(--color-error);
  }
  .field--error .field__input:focus {
    box-shadow: 0 0 0 3px var(--color-error-subtle);
  }

  .field--disabled .field__input {
    opacity: 0.5;
    cursor: not-allowed;
    background: var(--color-surface-raised);
  }

  .field__hint {
    font-size: var(--text-sm);
    color: var(--color-text-muted);
    margin: 0;
  }

  .field__hint--error {
    color: var(--color-error);
  }
</style>
