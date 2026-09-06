import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';

const input = readFileSync('src/components/forms/Input/input.css', 'utf8');
const forms = readFileSync('src/components/forms/forms.css', 'utf8');

describe('form density sizing', () => {
  it('keeps single-line controls within the 44px/36px height contract', () => {
    // Body prose uses 26px leading: with 20px padding and 3px borders it
    // inflated comfortable controls to 49px. Control leading must be explicit.
    expect(input).toMatch(/\.paper-input\s*\{[^}]*font-size:\s*var\(--paper-type-body-size\);[^}]*line-height:\s*1\.25;/s);
    expect(forms).toMatch(/\.paper-textarea\s*\{[^}]*line-height:\s*var\(--paper-type-body-line-height\)/s);
  });

  it('bounds the native unit selector to the compound field height', () => {
    expect(forms).toMatch(/\.paper-compound-field > \.paper-compound-field__unit\s*\{[^}]*block-size:\s*calc\(var\(--paper-control-height\) - var\(--paper-border-static-width\)\)/s);
  });

  it('lets the nearest density scope control segmented radio geometry', () => {
    expect(forms).not.toMatch(/\[data-density="compact"\]\s+\.paper-radio-select__segment/);
    expect(forms).toMatch(/\.paper-radio-select__segment\s*\{[^}]*min-inline-size:\s*var\(--paper-radio-min-width\)/s);
  });
});

it('uses control leading for chip inputs rather than inheriting prose leading', () => {
  expect(forms).toMatch(/\.paper-combobox__chips\s*\{[^}]*font-size:\s*var\(--paper-type-body-size\);[^}]*line-height:\s*1\.25;/s);
});
